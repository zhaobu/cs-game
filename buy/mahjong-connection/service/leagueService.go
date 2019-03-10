package service

import (
	"mahjong-connection/core"
	"mahjong-connection/hall"
	"mahjong-connection/ierror"
	"mahjong-connection/league"
	"mahjong-connection/model"
	"mahjong-connection/protocal"
	"time"

	fbsCommon "mahjong-connection/fbs/Common"

	"github.com/fwhappy/util"
)

// 开启比赛场连接
func openLeague(u *model.User) {
	if u.League == nil {
		u.League = league.NewLeague(u.UserId, u.Version)
	} else if u.League.ConnStatus == league.CONN_STATUS_CLOSED {
		u.League.ConnStatus = league.CONN_STATUS_UNCONNECTED
		u.League.Queue = make(chan *protocal.ImPacket, 1024)
	} else {
		core.Logger.Error("[openLeague]比赛大厅连接已开启过了, userId:%v", u.UserId)
		return
	}
	go listenLeagueConnect(u)
	core.Logger.Info("[openLeague]userId:%v", u.UserId)
}

// 关闭比赛场连接
func closeLeague(u *model.User) {
	u.League.Close()
	core.Logger.Info("[closeLeague]userId:%v", u.UserId)
}

// 监听比赛大厅的连接
func listenLeagueConnect(u *model.User) {
	defer util.RecoverPanic()
	defer func() {
		u.League.ListenStatus = false
		core.Logger.Info("[listenLeagueConnect]连接监控退出, userId:%v", u.UserId)
	}()
	if u.League.ListenStatus {
		core.Logger.Warn("[listenLeagueConnect]监听已开启过了, userId:%v", u.UserId)
		return
	}

	for {
		// 判断用户是否已重新连接
		user := hall.UserMap.Load(u.UserId)
		if user == nil || user.ClientConn == nil || user.ClientConn != u.ClientConn {
			core.Logger.Info("[listenLeagueConnect]用户已下线或已重新登陆，退出监听")
			break
		}

		if u.League.NeedReconnect() {
			core.Logger.Info("[listenLeagueConnect]检测到用户需要去连接联赛大厅, userId:%v", u.UserId)
			suc, err := u.League.Connect()
			if suc {
				// 开启消息监听
				go listenLeagueMessageReceive(u)
				// 发送握手消息
				u.League.Handshake(u.Token)
			} else {
				core.Logger.Error("[listenLeagueConnect]连接联赛大厅失败，等候下一次连接, userId:%v, err:%v", u.UserId, err.Error())
			}
		}

		// 每3秒重复检测一次
		time.Sleep(3 * time.Second)
	}
}

// 接受消息
func listenLeagueMessageReceive(u *model.User) {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[listenLeagueMessageReceive]退出, userId:%v", u.UserId)
	}()
	core.Logger.Info("[listenLeagueMessageReceive]已开启, userId:%v", u.UserId)
	for {
		// 读取包内容
		p, err := protocal.ReadPacket(u.League.Conn)
		// 检查解析错误
		if err != nil {
			core.Logger.Info("[listenLeagueMessageReceive]disconnected, userId:%v, err:%v", u.UserId, err.Error())
			break
		}

		// 转发消息给用户
		packageId := p.GetPackage()
		switch packageId {
		case protocal.PACKAGE_TYPE_HANDSHAKE:
			u.League.HandshakeAck()
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK:
		case protocal.PACKAGE_TYPE_HEARTBEAT:
			core.Logger.Warn("[listenLeagueMessageReceive]异常的收到了心跳回应, userId:%v", u.UserId)
		case protocal.PACKAGE_TYPE_KICK:
			core.Logger.Info("[listenLeagueMessageReceive]received kick, userId:%v", u.UserId)
			closeLeague(u)
		case protocal.PACKAGE_TYPE_DATA:
			// 转发给用户
			u.SendMessageToClient(p)
		default:
			core.Logger.Warn("[listenLeagueMessageReceive]错误的消息id, userId:%v, packageId:%v", u.UserId, packageId)
		}
	}
}

// LeagueMessageRedirect 转发请求到联赛
func LeagueMessageRedirect(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		core.Logger.Warn("[LeagueMessageRedirect]user not online,id:%v", id)
		return ierror.NewError(-201, id)
	}

	// 如果用户未处于连接状态，帮用户进行重连
	openFlag := false
	if u.League == nil ||
		u.League.ConnStatus == league.CONN_STATUS_UNCONNECTED ||
		u.League.ConnStatus == league.CONN_STATUS_CLOSED {
		openLeague(u)
		openFlag = true
	}

	// 如果是新开的，且是拉取大厅列表的协议，就不放到队列了
	if openFlag && p.GetMessageId() == fbsCommon.CommandLeagueListRequest {
		// do nothing
	} else {
		u.League.AppendMessage(p)
	}

	core.Logger.Info("[LeagueMessageRedirect]转发消息,packageId:%v, messageId:%v", p.GetPackage(), p.GetMessageId())
	return nil
}
