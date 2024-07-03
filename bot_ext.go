package winbeebot

import(
	"time"
	"strconv"
	"encoding/json"
	"fmt"
)




/**
 * 禁言群员
 */
func (b *Bot)BanMember(gid int64, uid int64, sec int64) {
	if sec <= 0 {
		sec = 9999999999999
	}
	chatuserconfig := ChatMemberConfig{ChatID: gid, UserID: uid}
	boo := false
	restricconfig := RestrictChatMemberConfig{
		ChatMemberConfig: chatuserconfig,
		UntilDate:        time.Now().Unix() + sec,
		Permissions: &ChatPermissions{
			CanSendMessages:       boo,
			CanSendAudios:  boo,
			CanSendPhotos: boo,
			CanSendVideos:  boo,
			CanSendOtherMessages:  boo,
			CanAddWebPagePreviews: boo,
		},
	}
	_, _ = b.RequestTgbotapi(restricconfig)
}

func (b *Bot)UnBanMember(gid int64, uid int64, sec int64) {
	if sec <= 0 {
		sec = 9999999999999
	}
	chatuserconfig := ChatMemberConfig{ChatID: gid, UserID: uid}
	boo := true
	restricconfig := RestrictChatMemberConfig{
		ChatMemberConfig: chatuserconfig,
		UntilDate:        time.Now().Unix() + sec,
		Permissions: &ChatPermissions{
			CanSendMessages:       boo,
			CanSendAudios:  boo,
			CanSendPhotos: boo,
			CanSendVideos:  boo,
			CanSendOtherMessages:  boo,
			CanAddWebPagePreviews: boo,
		},
	}
	_, _ = b.RequestTgbotapi(restricconfig)
}

func (b *Bot)KickMember(gid int64, uid int64, revokeMsg bool) {
	// cmconf := ChatMemberConfig{ChatID: gid, UserID: uid}
	a := BanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: gid,
			UserID: uid,
		},
		UntilDate:      99999999999,
		RevokeMessages: revokeMsg,
	}
	_, _ = b.RequestTgbotapi(a)
}
/**
 * 返回群组的所有管理员, 用来进行一次性@
 */
func (b *Bot)GetChatAdmins(gid int64) string {
	admins, _ := b.GetChatAdministratorsTgbotapi(gid)
	list := ""
	for _, admin := range admins {
		user := admin.User
		if user.IsBot {
			continue
		}
		list += "[" + user.String() + "](tg://user?id=" + strconv.FormatInt(admin.User.Id, 10) + ")\r\n"
	}
	return list
}


func  (b *Bot)IsGroupAdmin(gid int64, uid int64) bool {
	admins, _ := b.GetChatAdministratorsTgbotapi(gid)
	for _, admin := range admins {
		if admin.User.Id == uid {
			return true
		}
		
	}
	return false
}
func  (b *Bot)IsSuperAdmin(uid int64) bool {
	for _, admin := range b.Admins {
		if admin == uid {
			return true
		}
	}
	return false
}
func (b *Bot) GetInviteLink(chatid int64,name string) (string,error) {
	
	c:=ChatInviteLinkConfig{
		ChatConfig: ChatConfig{
			ChatID: chatid,
			SuperGroupUsername: name,
		},
	}
	resp, err := b.RequestTgbotapi(c)
	if err != nil {
		return "", err
	}

	var inviteLink string
	err = json.Unmarshal(resp.Result, &inviteLink)

	return inviteLink, err
}

func (b *Bot) SendAndDelete(msg Chattable, second int) {
	m, err := b.Send(msg)
	if err == nil {
		go b.DeletMessageDelay(m.Chat.Id, int(m.MessageId), second)
	}
}
func (b *Bot) CheckIsMember(groupId, userid int64) bool {
		c :=GetChatMemberConfig{
			ChatConfigWithUser: ChatConfigWithUser{
				ChatID: groupId,
				UserID: userid,
			},
		}
	resp, err := b.RequestTgbotapi(c)
	if err != nil {
		b.Log.Sugar().Error("检查用户是否为群成员出错", err)
		return false
	}

	var m ChatMemberTgbotapi
	err = json.Unmarshal(resp.Result, &m)

	if err == nil && (m.Status == "creator" || m.Status == "administrator" || m.Status == "member" || m.IsMember) {
		return true
	} else if err != nil {
		b.Log.Sugar().Error("检查用户是否为群成员出错", err)
		return false
	}
	return false
}
func (b *Bot) SendText(chatId int64, txt string) (Message, error) {
	msg := NewMessage(chatId, txt)
	msg.ParseMode = ModeMarkdownV2
	return b.Send(msg)
}
func (b *Bot) SendHtml(chatId int64, txt string) (Message, error) {
	msg := NewMessage(chatId, txt)
	msg.ParseMode = ModeHTML
	msg.DisableWebPagePreview = true
	return b.Send(msg)
}
func (b *Bot) SendHtmlAndDelete(chatId int64, txt string,second int) {
	msg := NewMessage(chatId, txt)
	msg.ParseMode = ModeHTML
	msg.DisableWebPagePreview = true
	m, err := b.Send(msg)
	if err != nil {
		return
	}
	b.DeletMessageDelay(chatId, int(m.MessageId), second)
}
func (b *Bot) SendTextAndDelete(chatId int64, txt string, second int) {
	msg := NewMessage(chatId, txt)
	msg.ParseMode = ModeMarkdownV2
	m, err := b.Send(msg)
	if err != nil {
		return
	}
	b.DeletMessageDelay(chatId, int(m.MessageId), second)
}
func (b *Bot) SendTextToAdmins(txt string) error {
	for _, id := range b.Admins {
		msg := NewMessage(id, txt)
		msg.ParseMode = ModeMarkdownV2
		msg.DisableWebPagePreview = true
		_, err := b.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// 踢出群员
func (b *Bot) KickGroupMember(groupId int64, uid int64) error {
	kickConfig := KickChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: groupId,
			UserID: uid,
		},
		RevokeMessages: true,
	}
	_, err := b.RequestTgbotapi(kickConfig)
	return err
}

// bot主动退群
func (b *Bot) LeaveGroup(groupId int64) error {
	leaveConfig := LeaveChatConfig{
		ChatID: groupId,
	}
	_, err := b.RequestTgbotapi(leaveConfig)
	return err
}

func (b *Bot) SendHtmlToAdmins(txt string) error {
	for _, id := range b.Admins {
		msg := NewMessage(id, txt)
		msg.ParseMode = ModeHTML
		_, err := b.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) DeletMessage(chatID int64, messageID int) (Message, error) {
	codem := NewDeleteMessage(chatID, int64(messageID))
	return b.Send(codem)
}
func (b *Bot) DeletMessageDelay(chatID int64, messageID int, seconds int) {
	codem := NewDeleteMessage(chatID, int64(messageID))
	if seconds > 0 {
		go func(aa DeleteMessageConfig) {
			time.Sleep(time.Duration(seconds) * time.Second)
			b.RequestTgbotapi(aa)
		}(codem)
	} else if seconds == 0 {
		b.RequestTgbotapi(codem)
	} else {
		return
	}
}

func (b *Bot) PinChatMessageTgbotapi(chatID int64, messageID int) (*APIResponse, error){
	pinConfig := PinChatMessageConfig{
		ChatID:              chatID,
		MessageID:           messageID,
		DisableNotification: false,
	}
	return b.RequestTgbotapi(pinConfig)
}


func (b *Bot) GetChatMembersCount(chatID int64) (int, error) {
	c := ChatMemberCountConfig{
		ChatConfig: ChatConfig{
			ChatID: chatID,
		},
	}
	resp,err :=b.RequestTgbotapi(c)
	if err != nil {
		return -1, err
	}

	var count int
	err = json.Unmarshal(resp.Result, &count)

	return count, err
}

func (b *Bot)GetGroupLink(chatId int64) string {
	chat, err := b.GetChatTgbotapi(ChatInfoConfig{
		ChatConfig: ChatConfig{
			ChatID: chatId,
		},
	})
	if err == nil {
		if (chat.Type == "group" || chat.Type == "supergroup") && chat.Username != "" {
			groupLink := fmt.Sprintf("https://t.me/%s", chat.Username)
			return groupLink
		} else if chat.InviteLink != "" {
			return chat.InviteLink
		}
	}
	return ""
}
func (b *Bot)SelfLink() string {
	return "https://t.me/" + b.Username
}

// 获取群创建者
func (b *Bot)GetGroupCreator(gid int64) *User {
	admins, _ := b.GetChatAdministratorsTgbotapi(gid)
	for _, admin := range admins {
		if admin.IsCreator() {
			return admin.User
		}
	}
	return nil
}

// 获取用户id
func (b *Bot)GetUserIdFromUpdate(update Update) int64 {
	if update.Message != nil {
		return update.Message.From.Id
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.Id
	}
	if update.InlineQuery != nil {
		return update.InlineQuery.From.Id
	}
	if update.ChatJoinRequest != nil {
		upmsg := update.ChatJoinRequest
		return upmsg.From.Id
	}
	return 0
}

// 获取对话(群组)id
func (b *Bot)GetChatIdFromUpdate(update Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.Id
	}
	if update.CallbackQuery != nil {
		c := update.CallbackQuery.Message.GetChat()
		return c.Id
	}
	if update.InlineQuery != nil {
		return update.InlineQuery.From.Id
	}
	if update.ChatJoinRequest != nil {
		upmsg := update.ChatJoinRequest
		return upmsg.Chat.Id
	}
	return 0
}

// 获取群聊id
func GetGroupIdFromUpdate(update Update) int64 {
	if update.FromChat() != nil && update.FromChat().IsPrivate() {
		return -1
	}
	if update.Message != nil {
		return update.Message.Chat.Id
	}
	if update.CallbackQuery != nil {
		c := update.CallbackQuery.Message.GetChat()
		return c.Id
	}
	if update.InlineQuery != nil {
		return update.InlineQuery.From.Id
	}
	if update.ChatJoinRequest != nil {
		upmsg := update.ChatJoinRequest
		return upmsg.Chat.Id
	}
	return 0
}

func (b *Bot)RestrictUser(chatId int64, userId int64) error {
	bl := false
	rf := RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate: Forever(),
		Permissions: &ChatPermissions{
			CanSendMessages:       bl,
			CanSendAudios:  bl,
			CanSendPhotos: bl,
			CanSendVideos:  bl,
			CanSendOtherMessages:  bl,
			CanAddWebPagePreviews: bl,
		},
	}
	_, err := b.RequestTgbotapi(rf)
	return err
}



func (bot *Bot)RestrictUserADay(chatId int64, userId int64) error {
	b := false
	rf := RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate: ADay(),
		Permissions: &ChatPermissions{
			CanSendMessages:       b,
			CanSendAudios:  b,
			CanSendPhotos: b,
			CanSendVideos:  b,
			CanSendOtherMessages:  b,
			CanAddWebPagePreviews: b,
		},
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}
func (bot *Bot)RestrictUserByTime(chatId int64, userId int64, hour int) error {
	b := false
	rf := RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate: time.Now().Add(1 * time.Duration(hour) * time.Hour).Unix(),
		Permissions: &ChatPermissions{
			CanSendMessages:       b,
			CanSendAudios:  b,
			CanSendPhotos:  b,
			CanSendVideos:  b,
			CanSendOtherMessages:  b,
			CanAddWebPagePreviews: b,
		},
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}
func (bot *Bot)RestrictUseMedia(chatId int64, userId int64, minutes int) error {
	b := false
	rf := RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate: ToNMinutes(minutes),
		Permissions: &ChatPermissions{
			CanSendMessages:       b,
			CanSendAudios:  b,
			CanSendPhotos:  b,
			CanSendVideos:  b,
			CanSendOtherMessages:  b,
			CanAddWebPagePreviews: b,
		},
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}

func (bot *Bot)UnRestrictUser(chatId int64, userId int64) error {
	b := true
	rf := RestrictChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate: Forever(),
		Permissions: &ChatPermissions{
			CanSendMessages:       b,
			CanSendAudios:  b,
			CanSendPhotos:  b,
			CanSendVideos:  b,
			CanSendOtherMessages:  b,
			CanAddWebPagePreviews: b,
		},
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}

// 屏蔽用户时可以删除所有信息
func (bot *Bot)BanUser(chatId int64, userId int64) error {
	rf := BanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate:      Forever(),
		RevokeMessages: true,
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}
func (bot *Bot)BanUserForTime(chatId int64, userId int64, hour int) error {
	rf := BanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		UntilDate:      time.Now().Add(1 * time.Duration(hour) * time.Hour).Unix(),
		RevokeMessages: true,
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}
func (bot *Bot)UnBanUser(chatId int64, userId int64) error {
	rf := UnbanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		OnlyIfBanned: true,
	}
	_, err := bot.RequestTgbotapi(rf)
	return err
}

func Forever() int64 {
	return time.Now().Add(367 * 24 * time.Hour).Unix()
}
func ADay() int64 {
	return time.Now().Add(1 * 24 * time.Hour).Unix()
}
func ToNMinutes(n int) int64 {
	return time.Now().Add(time.Duration(n) * time.Minute).Unix()
}
func (b *Bot) SendCallBack(qid string, txt string) (Message, error) {
	msg := NewCallback(qid, txt)
	return b.Send(msg)
}
func (b *Bot) SendCallBackAlert(qid string, txt string) (Message, error) {
	msg := NewCallbackWithAlert(qid, txt)
	return b.Send(msg)
}

func (b *Bot) SetPrivateCommands(cmdName,desc []string) {
	if len(cmdName)!=len(desc){
	  panic("命令和参数不符合")
	}
	var cmds []BotCommand
	for i,v:= range cmdName{
	  cmd := BotCommand{
		Command: v,
		Description: desc[i],
	  }
	  cmds = append(cmds,cmd)
	}
	setCommands := NewSetMyCommandsWithScope(NewBotCommandScopeAllPrivateChats(), cmds...)
  
	if _, err := b.RequestTgbotapi(setCommands); err != nil {
	  panic("Unable to set commands")
	}
  }
  func (b *Bot) SetGroupCommands(cmdName,desc []string) {
	if len(cmdName)!=len(desc){
	  panic("命令和参数不符合")
	}
	var cmds []BotCommand
	for i,v:= range cmdName{
	  cmd := BotCommand{
		Command: v,
		Description: desc[i],
	  }
	  cmds = append(cmds,cmd)
	}
	setCommands := NewSetMyCommandsWithScope(NewBotCommandScopeAllGroupChats(), cmds...)
  
	if _, err := b.RequestTgbotapi(setCommands); err != nil {
	  panic("Unable to set commands")
	}
  }
