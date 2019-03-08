package service

import (
	"mahjong-connection/core"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/game"
	"mahjong-connection/hall"
	"mahjong-connection/ierror"
	"mahjong-connection/model"
	"mahjong-connection/protocal"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

func openGame(u *model.User, selectType string) {
	if u.Game == nil {
		u.Game = game.NewGame(u.UserId, u.Version, selectType)
	} else if u.Game.IsClosed() {
		u.Game.SelectType = selectType
		u.Game.ConnStatus = game.CONN_STATUS_UNCONNECTED
		u.Game.Queue = make(chan *protocal.ImPacket, 1024)
	} else {
		core.Logger.Error("[openGame]游戏服连接已开启过了, userId:%v", u.UserId)
		return
	}
	core.Logger.Info("[openGame]userId:%v, version:%v, selectType:%v", u.UserId, u.Version, selectType)
	listenGameConnect(u)
}

// 监听比赛大厅的连接
func listenGameConnect(u *model.User) {
	core.Logger.Info("[listenGameConnect]用户需要去连接游戏服, userId:%v", u.UserId)
	code, err := u.Game.Connect()
	if code == 0 {
		// 开启消息监听
		go listenGameMessageReceive(u)
		// 发送握手消息
		u.Game.Handshake(u.Token, u.ClientRoomId, u.ClientRound, u.ClientSeq)
	} else {
		u.Game.ConnStatus = game.CONN_STATUS_CLOSED
		core.Logger.Error("[listenGameConnect]连接游戏服失败，等候下一次连接, userId:%v, err:%v", u.UserId, err.Error())
		// TODO 通知客户端
		err := ierror.NewError(code)
		u.WriteMessageToClient(CloseGamePush(code, err.Error()))
	}
}

// 关闭比赛场连接
func closeGame(u *model.User) {
	u.Game.Close()
	core.Logger.Info("[closeGame]userId:%v", u.UserId)
}

// 接受消息
func listenGameMessageReceive(u *model.User) {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[listenGameMessageReceive]退出, userId:%v", u.UserId)
		closeGame(u)
	}()
	core.Logger.Info("[listenGameMessageReceive]已开启, userId:%v", u.UserId)
	for {
		// 读取包内容
		p, err := protocal.ReadPacket(u.Game.Conn)
		// 检查解析错误
		if err != nil {
			core.Logger.Info("[listenGameMessageReceive]disconnected, userId:%v, err:%v", u.UserId, err.Error())
			break
		}

		// 转发消息给用户
		packageId := p.GetPackage()
		switch packageId {
		case protocal.PACKAGE_TYPE_HANDSHAKE:
			// 解析握手参数
			js, _ := simplejson.NewJson(p.GetMessage())
			code, _ := js.Get("code").Int()
			msg, _ := js.Get("message").String()
			if code != 0 {
				u.WriteMessageToClient(CloseGamePush(code, msg))
				core.Logger.Error("[listenGameMessageReceive]HandShake error, userId:%v, code:%v, message:%v", u.UserId, code, msg)
			} else {
				u.Game.HandshakeAck()
			}
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK:
		case protocal.PACKAGE_TYPE_HEARTBEAT:
			core.Logger.Warn("[listenGameMessageReceive]异常的收到了心跳回应, userId:%v", u.UserId)
		case protocal.PACKAGE_TYPE_KICK:
			core.Logger.Info("[listenGameMessageReceive]received kick, userId:%v", u.UserId)
			closeGame(u)
		case protocal.PACKAGE_TYPE_DATA:
			// 转发给用户
			u.SendMessageToClient(p)
			// core.Logger.Info("[listenGameMessageReceive]转发游戏消息给用户,userId:%v, version:%v, packageId:%v, messageId:%v, seq:%v", u.UserId, u.Version, p.GetPackage(), p.GetMessageId(), p.GetMessageIndex())
			// 如果是退出房间，则关闭连接
			if p.GetMessageId() == fbsCommon.CommandCloseRoomPush {
				closeGame(u)
			}
		default:
			core.Logger.Warn("[listenGameMessageReceive]错误的消息id, userId:%v, packageId:%v", u.UserId, packageId)
		}
	}
}

// GameMessageRedirect 转发请求到游戏服
func GameMessageRedirect(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		core.Logger.Warn("[GameMessageRedirect]user not online,id:%v", id)
		return ierror.NewError(-201, id)
	}

	// 部分操作前，需要断开重新连接
	if u.Game != nil && u.Game.IsConnected() {
		switch p.GetMessageId() {
		case fbsCommon.CommandCreateRoomRequest:
			fallthrough
		case fbsCommon.CommandJoinRoomRequest:
			fallthrough
		case fbsCommon.CommandRandomRoomRequest:
			fallthrough
		case fbsCommon.CommandMatchRoomRequest:
			fallthrough
		case fbsCommon.CommandCoinRoomRequest:
			fallthrough
		case fbsCommon.CommandRankRoomRequest:
			core.Logger.Info("导致进房间的操作，需要重新进行选服、连接")
			closeGame(u)
		default:
		}
	}

	// 处理用户连接
	if u.Game == nil || u.Game.IsClosed() {
		u.ClientRoomId = 0
		u.ClientRound = 0
		u.ClientSeq = 0
		selectType := getSelectType(p.GetMessageId())
		if selectType == "JOIN_ROOM" {
			// 加入房间
			request := fbsCommon.GetRootAsJoinRoomRequest(p.GetBody(), 0)
			selectType = string(request.Number())
		}
		core.Logger.Debug("[GameMessageRedirect]openGame, id:%v", id)
		openGame(u, selectType)
	}

	u.Game.AppendMessage(p)

	core.Logger.Info("[GameMessageRedirect]转发消息,userId:%v, version:%v, packageId:%v, messageId:%v, seq:%v", id, u.Version, p.GetPackage(), p.GetMessageId(), p.GetMessageIndex())

	return nil
}

// 根据消息id， 返回待连接类型
func getSelectType(messageId uint16) string {
	switch messageId {
	case fbsCommon.CommandCreateRoomRequest:
		return "CREATE_ROOM"
	case fbsCommon.CommandRandomRoomRequest:
		return "RANDOM_ROOM"
	case fbsCommon.CommandMatchRoomRequest:
		return "KING_ROOM"
	case fbsCommon.CommandCoinRoomRequest:
		return "COIN_ROOM"
	case fbsCommon.CommandRankRoomRequest:
		return "RANK"
	case fbsCommon.CommandJoinRoomRequest:
		return "JOIN_ROOM"
	default:
		return "RECONNCET_ROOM"
	}
}
