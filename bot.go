package winbeebot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"log"
	"strings"
	"mime/multipart"
	"io"
	"io/ioutil"
	"net/url"
	"go.uber.org/zap"
)

//go:generate go run ./scripts/generate

// Bot is the default Bot struct used to send and receive messages to the telegram API.
type Bot struct {
	// Token stores the bot's secret token obtained from t.me/BotFather, and used to interact with telegram's API.
	Token string
	Debug bool
	Log         *zap.Logger
	// The bot's User info, as returned by Bot.GetMe. Populated when created through the NewBot method.
	User
	// The bot client to use to make requests
	BotClient
	Admins []int64
	ExtraData *interface{}
}

// BotOpts declares all optional parameters for the NewBot function.
type BotOpts struct {
	// BotClient allows for passing in custom configurations of BotClient, such as handling extra errors or providing
	// metrics.
	BotClient BotClient
	// DisableTokenCheck can be used to disable the token validity check.
	// Useful when running in time-constrained environments where the startup time should be minimised, and where the
	// token can be assumed to be valid (eg lambdas).
	// Warning: Disabling the token check will mean that the Bot.User struct will no longer be populated.
	DisableTokenCheck bool
	// Request opts to use for checking token validity with Bot.GetMe. Can be slow - a high timeout (eg 10s) is
	// recommended.
	RequestOpts *RequestOpts
}

// NewBot returns a new Bot struct populated with the necessary defaults.
func NewBot(token string,adminsId []int64,log *zap.Logger, debug bool, extraData *interface{},opts *BotOpts) (*Bot, error) {
	botClient := BotClient(&BaseBotClient{
		Client:             http.Client{},
		UseTestEnvironment: false,
		DefaultRequestOpts: nil,
	})

	// Large timeout on the initial GetMe request as this can sometimes be slow.
	getMeReqOpts := &RequestOpts{
		Timeout: 10 * time.Second,
		APIURL:  DefaultAPIURL,
	}

	checkTokenValidity := true
	if opts != nil {
		if opts.BotClient != nil {
			botClient = opts.BotClient
		}

		if opts.RequestOpts != nil {
			getMeReqOpts = opts.RequestOpts
		}
		checkTokenValidity = !opts.DisableTokenCheck
	}

	b := Bot{
		Token:     token,
		BotClient: botClient,
		Admins: adminsId,
		Log: log,
		Debug: debug,
		ExtraData: extraData,
	}

	if checkTokenValidity {
		// Get bot info. This serves two purposes:
		// 1. Check token is valid.
		// 2. Populate the bot struct "User" field.
		botUser, err := b.GetMe(&GetMeOpts{RequestOpts: getMeReqOpts})
		if err != nil {
			return nil, fmt.Errorf("failed to check bot token: %w", err)
		}
		b.User = *botUser
	}

	return &b, nil
}

// UseMiddleware allows you to wrap the existing bot client to enhance functionality
//
// Deprecated: Instead of using middlewares, consider implementing the BotClient interface.
func (bot *Bot) UseMiddleware(mw func(client BotClient) BotClient) *Bot {
	bot.BotClient = mw(bot.BotClient)
	return bot
}

var ErrNilBotClient = errors.New("nil BotClient")

func (bot *Bot) Request(method string, params map[string]string, data map[string]NamedReader, opts *RequestOpts) (json.RawMessage, error) {
	if bot.BotClient == nil {
		return nil, ErrNilBotClient
	}

	ctx, cancel := bot.BotClient.TimeoutContext(opts)
	defer cancel()

	return bot.BotClient.RequestWithContext(ctx, bot.Token, method, params, data, opts)
}
func (bot *Bot) UpdateExtraData(data *interface{}){
	bot.ExtraData = data
}
func (bot *Bot) AddAdmin(id int64){
	bot.Admins = append(bot.Admins, id)
}
func (bot *Bot) RemoveAdmin(adminId int64){
	for idx,id := range bot.Admins{
		if id == adminId{
			bot.Admins = append(bot.Admins[:idx], bot.Admins[idx+1:]...)
			break
		}
	}
	
}
func hasFilesNeedingUpload(files []RequestFile) bool {
	for _, file := range files {
		if file.Data.NeedsUpload() {
			return true
		}
	}

	return false
}
// Request sends a Chattable to Telegram, and returns the APIResponse.
func (bot *Bot) RequestTgbotapi(c Chattable) (*APIResponse, error) {
	params, err := c.params()
	if err != nil {
		return nil, err
	}

	if t, ok := c.(Fileable); ok {
		files := t.files()

		// If we have files that need to be uploaded, we should delegate the
		// request to UploadFile.
		if hasFilesNeedingUpload(files) {
			return bot.UploadFiles(t.method(), params, files)
		}

		// However, if there are no files to be uploaded, there's likely things
		// that need to be turned into params instead.
		for _, file := range files {
			params[file.Name] = file.Data.SendData()
		}
	}

	return bot.MakeRequest(c.method(), params)
}

// Send will send a Chattable item to Telegram and provides the
// returned Message.
func (bot *Bot) Send(c Chattable) (Message, error) {
	resp, err := bot.RequestTgbotapi(c)
	if err != nil {
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal(resp.Result, &message)

	return message, err
}

const apiEndpoint ="https://api.telegram.org/bot%s/%s"
// MakeRequest makes a request to a specific endpoint with our token.
func (bot *Bot) MakeRequest(endpoint string, params Params) (*APIResponse, error) {
	if bot.Debug {
		log.Printf("Endpoint: %s, params: %v\n", endpoint, params)
	}

	method := fmt.Sprintf(apiEndpoint, bot.Token, endpoint)

	values := buildParams(params)

	req, err := http.NewRequest("POST", method, strings.NewReader(values.Encode()))
	if err != nil {
		return &APIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bot.BotClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.Debug {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Code:               apiResp.ErrorCode,
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}
// UploadFiles makes a request to the API with files.
func (bot *Bot) UploadFiles(endpoint string, params Params, files []RequestFile) (*APIResponse, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	// This code modified from the very helpful @HirbodBehnam
	// https://github.com/go-telegram-bot-api/telegram-bot-api/issues/354#issuecomment-663856473
	go func() {
		defer w.Close()
		defer m.Close()

		for field, value := range params {
			if err := m.WriteField(field, value); err != nil {
				w.CloseWithError(err)
				return
			}
		}

		for _, file := range files {
			if file.Data.NeedsUpload() {
				name, reader, err := file.Data.UploadData()
				if err != nil {
					w.CloseWithError(err)
					return
				}

				part, err := m.CreateFormFile(file.Name, name)
				if err != nil {
					w.CloseWithError(err)
					return
				}

				if _, err := io.Copy(part, reader); err != nil {
					w.CloseWithError(err)
					return
				}

				if closer, ok := reader.(io.ReadCloser); ok {
					if err = closer.Close(); err != nil {
						w.CloseWithError(err)
						return
					}
				}
			} else {
				value := file.Data.SendData()

				if err := m.WriteField(file.Name, value); err != nil {
					w.CloseWithError(err)
					return
				}
			}
		}
	}()

	if bot.Debug {
		log.Printf("Endpoint: %s, params: %v, with %d files\n", endpoint, params, len(files))
	}

	method := fmt.Sprintf(apiEndpoint, bot.Token, endpoint)

	req, err := http.NewRequest("POST", method, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := bot.BotClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return &apiResp, err
	}

	if bot.Debug {
		log.Printf("Endpoint: %s, response: %s\n", endpoint, string(bytes))
	}

	if !apiResp.Ok {
		var parameters ResponseParameters

		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}

		return &apiResp, &Error{
			Message:            apiResp.Description,
			ResponseParameters: parameters,
		}
	}

	return &apiResp, nil
}

func buildParams(in Params) url.Values {
	if in == nil {
		return url.Values{}
	}

	out := url.Values{}

	for key, value := range in {
		out.Set(key, value)
	}

	return out
}
func (bot *Bot) decodeAPIResponse(responseBody io.Reader, resp *APIResponse) ([]byte, error) {
	if !bot.Debug {
		dec := json.NewDecoder(responseBody)
		err := dec.Decode(resp)
		return nil, err
	}

	// if debug, read response body
	data, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bot *Bot) GetChatTgbotapi(config ChatInfoConfig) (Chat, error) {
	resp, err := bot.RequestTgbotapi(config)
	if err != nil {
		return Chat{}, err
	}

	var chat Chat
	err = json.Unmarshal(resp.Result, &chat)

	return chat, err
}