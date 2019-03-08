package service

import (
	"io/ioutil"
	"mahjong-connection/core"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/hall"
	"mahjong-connection/model"
	"mahjong-connection/protocal"
	"net/http"
	"strings"

	"github.com/fwhappy/util"
)

// PrivateMessage 推送消息给单人
func PrivateMessage(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Logger.Error("[web.PrivateMessage]read body error:%v", err.Error())
		return
	}
	// 解析消息
	p := protocal.NewPacket(b)
	push := fbsCommon.GetRootAsGatewayS2CPrivateMessagePush(p.GetBody(), 0)
	message := push.Message(new(fbsCommon.Message))
	// 消息接收者
	userId := int(push.UserId())

	// 将消息转发给用户
	if userId == 0 {
		core.Logger.Error("[web.PrivateMessage]userId = 0")
		return
	}

	// 读取用户
	u := hall.UserMap.Load(userId)
	if u == nil {
		// 用户不在线
		core.Logger.Trace("[web.PrivateMessage]用户不在线, userId:%v", userId)
		return
	}

	// 给用户发送消息
	u.SendMessageToClient(p)
	core.Logger.Trace("[web.PrivateMessage]userId:%v, senderId:%v, messageId:%v, content:%v", userId, message.SenderId(), message.MessageId(), string(message.Content()))
}

// BroadcastMessage 推送消息给所有人
func BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Logger.Error("[web.BroadcastMessage]read body error:%v", err.Error())
		return
	}
	// 解析消息
	p := protocal.NewPacket(b)
	push := fbsCommon.GetRootAsGatewayS2CPrivateMessagePush(p.GetBody(), 0)
	message := push.Message(new(fbsCommon.Message))

	core.Logger.Trace("[web.BroadcastMessage]准备发送, senderId:%v, messageId:%v, content:%v", message.SenderId(), message.MessageId(), string(message.Content()))
	go func() {
		defer util.RecoverPanic()
		defer core.Logger.Trace("[web.BroadcastMessage]发送完成, senderId:%v, messageId:%v, content:%v", message.SenderId(), message.MessageId(), string(message.Content()))
		hall.UserMap.LoadUsers().Range(func(k, v interface{}) bool {
			u := v.(*model.User)
			// 低版本不支持这个消息
			if strings.Compare(u.Version, "3.3.0") > 0 {
				core.Logger.Debug("[web.BroadcastMessage]发给用户, userId:%v, senderId:%v, messageId:%v, content:%v", u.UserId, message.SenderId(), message.MessageId(), string(message.Content()))
				u.SendMessageToClient(p)
			}
			return true
		})
	}()
}
