package service

import (
	"mahjong-connection/club"
	"mahjong-connection/core"
	"mahjong-connection/hall"
	"mahjong-connection/ierror"
	"mahjong-connection/model"
	"mahjong-connection/protocal"
	"time"

	"github.com/fwhappy/util"
)

// 开启比赛场连接
func openClub(u *model.User) {
	if u.Club == nil {
		u.Club = club.NewClub(u.UserId, u.Version)
	} else if u.Club.ConnStatus == club.CONN_STATUS_CLOSED {
		u.Club.ConnStatus = club.CONN_STATUS_UNCONNECTED
		u.Club.Queue = make(chan *protocal.ImPacket, 1024)
	} else {
		core.Logger.Error("[openClub]俱乐部连接已开启过了, userId:%v", u.UserId)
		return
	}
	go listenClubConnect(u)
	core.Logger.Info("[openClub]userId:%v", u.UserId)
}

// 关闭比赛场连接
func closeClub(u *model.User) {
	u.Club.Close()
	core.Logger.Info("[closeClub]userId:%v", u.UserId)
}

// 监听比赛大厅的连接
func listenClubConnect(u *model.User) {
	defer util.RecoverPanic()
	defer func() {
		u.Club.ListenStatus = false
		core.Logger.Info("[listenClubConnect]连接监控退出, userId:%v", u.UserId)
	}()
	if u.Club.ListenStatus {
		core.Logger.Warn("[listenClubConnect]监听已开启过了, userId:%v", u.UserId)
		return
	}

	for {
		// 判断用户是否已重新连接
		user := hall.UserMap.Load(u.UserId)
		if user == nil || user.ClientConn == nil || user.ClientConn != u.ClientConn {
			core.Logger.Info("[listenClubConnect]用户已下线或已重新登陆，退出监听")
			break
		}

		if u.Club.NeedReconnect() {
			core.Logger.Info("[listenClubConnect]检测到用户需要去连接俱乐部, userId:%v", u.UserId)
			suc, err := u.Club.Connect()
			if suc {
				// 开启消息监听
				go listenClubMessageReceive(u)
				// 发送握手消息
				u.Club.Handshake(u.Token)
			} else {
				core.Logger.Error("[listenClubConnect]连接俱乐部失败，等候下一次连接, userId:%v, err:%v", u.UserId, err.Error())
			}
		}

		// 每3秒重复检测一次
		time.Sleep(3 * time.Second)
	}
}

// 接受消息
func listenClubMessageReceive(u *model.User) {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[listenClubMessageReceive]退出, userId:%v", u.UserId)
	}()
	core.Logger.Info("[listenClubMessageReceive]已开启, userId:%v", u.UserId)
	for {
		// 读取包内容
		p, err := protocal.ReadPacket(u.Club.Conn)
		// 检查解析错误
		if err != nil {
			core.Logger.Info("[listenClubMessageReceive]disconnected, userId:%v, err:%v", u.UserId, err.Error())
			break
		}

		// 转发消息给用户
		packageId := p.GetPackage()
		switch packageId {
		case protocal.PACKAGE_TYPE_HANDSHAKE:
			u.Club.HandshakeAck()
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK:
		case protocal.PACKAGE_TYPE_HEARTBEAT:
			core.Logger.Warn("[listenClubMessageReceive]异常的收到了心跳回应, userId:%v", u.UserId)
		case protocal.PACKAGE_TYPE_KICK:
			core.Logger.Info("[listenClubMessageReceive]received kick, userId:%v", u.UserId)
			closeClub(u)
		case protocal.PACKAGE_TYPE_DATA:
			// 转发给用户
			u.SendMessageToClient(p)
		default:
			core.Logger.Warn("[listenClubMessageReceive]错误的消息id, userId:%v, packageId:%v", u.UserId, packageId)
		}
	}
}

// ClubMessageRedirect 转发请求到联赛
func ClubMessageRedirect(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		core.Logger.Warn("[ClubMessageRedirect]user not online,id:%v", id)
		return ierror.NewError(-201, id)
	}

	// 如果用户未处于连接状态，帮用户进行重连
	if u.Club == nil ||
		u.Club.ConnStatus == club.CONN_STATUS_UNCONNECTED ||
		u.Club.ConnStatus == club.CONN_STATUS_CLOSED {
		openClub(u)
		// 连接失败，通知客户端
	}

	u.Club.AppendMessage(p)

	core.Logger.Info("[ClubMessageRedirect]转发消息,packageId:%v, messageId:%v", p.GetPackage(), p.GetMessageId())
	return nil
}
