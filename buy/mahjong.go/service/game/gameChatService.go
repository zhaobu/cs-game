package game

import (
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// RoomChat 房间聊天
func RoomChat(userId int, chatId int16, memberId uint8, content string) *core.Error {
	// 判断用户是否已连接
	user, room, err := getUserRoom(userId)
	if err != nil {
		return nil
	}

	// 敏感词过滤
	if len(content) > 0 {
		replacedContent, err := core.FilterManager.Filter().Replace(content, '*')
		if err != nil {
			core.Logger.Warn("[RoomChat]敏感词过滤失败,userId:%v, content:%v, replacedContent:%v", userId, content, replacedContent)
		} else {
			content = replacedContent
		}
	}

	pushPacket := RoomChatPush(userId, chatId, memberId, content)

	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 推送消息给用户
	// 若ios用户离线，则推送一条消息给APNs，再转发给用户
	for _, u := range room.GetUsers() {
		if u.UserId == userId {
			// 跳过用户自己
			continue
		}
		tUser, err := UserMap.GetUser(u.UserId)
		if err == nil {
			tUser.AppendMessage(pushPacket)
		}

		// 文字聊天，推送push
		if len(content) > 0 {
			u.SendPush(int(chatId), user.UserId, user.Info.Nickname, content)
		}
	}
	core.Logger.Info("[RoomChat]userId:%d,roomId:%d,chatId:%d.", userId, room.RoomId, chatId)
	return nil
}

// RoomNotice 房间通知
func RoomNotice(userId int, chatId int16, memberId uint8, content string) *core.Error {
	_, room, err := getUserRoom(userId)
	if err != nil {
		return nil
	}

	// 推送消息给用户
	if room.GetUsersLen() > 0 {
		pushPacket := RoomChatPush(userId, chatId, memberId, content)
		room.Mux.Lock()
		defer room.Mux.Unlock()
		if chatId == config.CHAT_ID_VOICE_ID {
			// 语音消息也给用户自己发
			room.SendMessageToRoomUser(pushPacket, 0)
			// core.Logger.Debug("[RoomNotice]收到语音消息,userId:%v,content:%v", userId, content)
		} else {
			room.SendMessageToRoomUser(pushPacket, userId)
		}
		room.Ob.sendMessage(pushPacket, 0)
	}

	// 量太大，不记录
	//core.Logger.Info("[RoomNotice]userId:%d,roomId:%d,chatId:%d.", userId, room.RoomId, chatId)

	return nil
}
