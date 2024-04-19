package ext

import (
	"strings"

	"github.com/selfmakeit/winbeebot"
)

// TODO: extend to be used as a generic cancel context?
type Context struct {
	// winbeebot.Update is inlined so that we can access all fields immediately if necessary.
	*winbeebot.Update
	// Data represents update-local storage.
	// This can be used to pass data across handlers - for example, to cache operations relevant to the current update,
	// such as admin checks.
	Data map[string]interface{}

	// EffectiveMessage is the message which triggered the update, if possible.
	EffectiveMessage *winbeebot.Message
	// EffectiveChat is the chat the update was triggered in, if possible.
	EffectiveChat *winbeebot.Chat
	// EffectiveUser is the user who triggered the update, if possible.
	// Note: when adding a user, the user who ADDED should be the EffectiveUser;
	// they caused the update. If a user joins naturally, then they are the EffectiveUser.
	//
	// WARNING: It may be better to rely on EffectiveSender instead, which allows for easier use
	// in the case of linked channels, anonymous admins, or anonymous channels.
	EffectiveUser *winbeebot.User
	// EffectiveSender is the sender of the update. This can be either:
	//  - a user
	//  - an anonymous admin of the current chat, speaking through the chat
	//  - the linked channel of the current chat
	//  - an anonymous user, speaking through a channel
	EffectiveSender *winbeebot.Sender
}

// NewContext populates a context with the relevant fields from the current update.
// It takes a data field in the case where custom data needs to be passed.
func NewContext(update *winbeebot.Update, data map[string]interface{}) *Context {
	var msg *winbeebot.Message
	var chat *winbeebot.Chat
	var user *winbeebot.User
	var sender *winbeebot.Sender

	switch {
	case update.Message != nil:
		msg = update.Message
		chat = &update.Message.Chat
		user = update.Message.From

	case update.EditedMessage != nil:
		msg = update.EditedMessage
		chat = &update.EditedMessage.Chat
		user = update.EditedMessage.From

	case update.ChannelPost != nil:
		msg = update.ChannelPost
		chat = &update.ChannelPost.Chat

	case update.EditedChannelPost != nil:
		msg = update.EditedChannelPost
		chat = &update.EditedChannelPost.Chat

	case update.BusinessConnection != nil:
		user = &update.BusinessConnection.User

	case update.BusinessMessage != nil:
		msg = update.BusinessMessage
		chat = &update.BusinessMessage.Chat
		user = update.BusinessMessage.From

	case update.EditedBusinessMessage != nil:
		msg = update.EditedBusinessMessage
		chat = &update.EditedBusinessMessage.Chat
		user = update.EditedBusinessMessage.From

	case update.DeletedBusinessMessages != nil:
		chat = &update.DeletedBusinessMessages.Chat

	case update.MessageReaction != nil:
		user = update.MessageReaction.User
		chat = &update.MessageReaction.Chat
		sender = update.MessageReaction.GetSender()

	case update.MessageReactionCount != nil:
		chat = &update.MessageReactionCount.Chat

	case update.InlineQuery != nil:
		user = &update.InlineQuery.From

	case update.ChosenInlineResult != nil:
		user = &update.ChosenInlineResult.From

	case update.CallbackQuery != nil:
		user = &update.CallbackQuery.From

		if update.CallbackQuery.Message != nil {
			switch m := update.CallbackQuery.Message.(type) {
			case winbeebot.Message:
				msg = &m
			case winbeebot.InaccessibleMessage:
				// Note: This conversion means that EffectiveMessage may not contain all Message fields
				msg = m.ToMessage()
			}

			tmpC := update.CallbackQuery.Message.GetChat()
			chat = &tmpC

			// Note: the sender is the sender of the CallbackQuery; not the sender of the CallbackQuery.Message.
			sender = &winbeebot.Sender{User: user, ChatId: chat.Id}
		}

	case update.ShippingQuery != nil:
		user = &update.ShippingQuery.From

	case update.PreCheckoutQuery != nil:
		user = &update.PreCheckoutQuery.From

	case update.Poll != nil:
		// no data

	case update.PollAnswer != nil:
		user = update.PollAnswer.User
		sender = update.PollAnswer.GetSender()

	case update.MyChatMember != nil:
		user = &update.MyChatMember.From
		chat = &update.MyChatMember.Chat

	case update.ChatMember != nil:
		user = &update.ChatMember.From
		chat = &update.ChatMember.Chat

	case update.ChatJoinRequest != nil:
		user = &update.ChatJoinRequest.From
		chat = &update.ChatJoinRequest.Chat

	case update.ChatBoost != nil:
		chat = &update.ChatBoost.Chat
		user = update.ChatBoost.Boost.Source.MergeChatBoostSource().User

	case update.RemovedChatBoost != nil:
		chat = &update.RemovedChatBoost.Chat
		user = update.RemovedChatBoost.Source.MergeChatBoostSource().User
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	if sender == nil {
		if msg != nil {
			sender = msg.GetSender()
		} else if user != nil {
			sender = &winbeebot.Sender{User: user}
			if chat != nil {
				sender.ChatId = chat.Id
			}
		}
	}

	return &Context{
		Update:           update,
		Data:             data,
		EffectiveMessage: msg,
		EffectiveChat:    chat,
		EffectiveUser:    user,
		EffectiveSender:  sender,
	}
}

// Args gets the list of whitespace-separated arguments of the message text.
func (c *Context) Args() []string {
	if c.EffectiveMessage == nil {
		return nil
	}

	return strings.Fields(c.EffectiveMessage.GetText())
}
