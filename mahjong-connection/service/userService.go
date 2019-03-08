package service

import (
	"mahjong-connection/config"
	"mahjong-connection/core"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/hall"
	"mahjong-connection/ierror"
	"mahjong-connection/model"
	"mahjong-connection/protocal"
	"mahjong-connection/response"
	"mahjong-connection/selectserver"
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

// ClientHandShake 客户端逻辑握手
func ClientHandShake(conn *net.TCPConn, p *protocal.ImPacket) (*model.User, *ierror.Error) {
	// 读取参数
	js, err := simplejson.NewJson(p.GetMessage())
	if !ierror.MustNil(err) {
		return nil, ierror.NewError(-10100, "ClientHandShake", err.Error())
	}
	// 解析出token
	token, err := js.Get("user").Get("token").String()
	if !ierror.MustNil(err) {
		return nil, ierror.NewError(-10101, "ClientHandShake", "token", err.Error())
	}
	// 解析出重连必要的三个参数
	// 客户端当前房间id
	clientRoomId, _ := js.Get("user").Get("room_id").Int64()
	// 客户端当前局
	round, _ := js.Get("user").Get("round").Int()
	// 客户端当前步骤
	seq, _ := js.Get("user").Get("step").Int()
	core.Logger.Debug("[ClientHandShake]参数, token:%v, clientRoomId:%v, round:%v, seq:%v", token, clientRoomId, round, seq)
	// 从token中解析出userId
	tokenInfo, _ := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	userId := int(tokenInfo[0].(float64))
	version := tokenInfo[3].(string)
	if userId <= 0 {
		return nil, ierror.NewError(-212, token)
	}
	core.Logger.Debug("token parse, token:%v, userId:%d, version:%v", token, userId, version)

	// 如果是重复登录，需要关闭之前的连接
	loginedUser := hall.UserMap.Load(userId)
	if loginedUser != nil && loginedUser.ClientConn != conn {
		repeatLogin(loginedUser, conn)
	}
	// 选服，获取用户当前的房间id和房间号
	result := selectserver.Reconnect(userId, version)

	// 当前时间
	t := util.GetTime()
	// 新建用户
	u := model.NewUser(userId, conn)
	u.Version = version
	u.Token = token
	u.ClientConnectTime = t
	u.ClientHandshakeTime = t
	if result.RoomId == clientRoomId {
		// 保存当前房间数据
		u.ClientRoomId = clientRoomId
		u.ClientRound = round
		u.ClientSeq = seq
	}
	// 保存用户
	hall.UserMap.Store(u)
	// 保存连接=>id的对应关系
	hall.ConnectionSet.Store(conn.RemoteAddr().String(), u.UserId)

	// 通知用户握手成功
	jsResponse := simplejson.New()
	jsResponse.Set("heartbeat", config.HEARTBEAT_INTERVAL)
	// jsResponse.Set("roomid", 0)
	// jsResponse.Set("raceid", 0)
	jsResponse.Set("roomid", result.RoomId)
	jsResponse.Set("raceid", result.RaceId)
	u.WriteMessageToClient(response.JSONSuccess(protocal.PACKAGE_TYPE_HANDSHAKE, jsResponse))

	core.Logger.Info("[ClientHandShake]userId:%d, server remote:%v, user remote:%v", userId, hall.RemoteAddr, conn.RemoteAddr().String())

	return nil, nil
}

// ClientHandShakeAck 客户端逻辑握手完成
func ClientHandShakeAck(id int, impacket *protocal.ImPacket) (*model.User, *ierror.Error) {
	u := hall.UserMap.Load(id)
	if u == nil {
		return nil, ierror.NewError(-201, id)
	}

	// 删除handshake超时监听
	hall.FinishListenHandShakeTimeout(u.ClientConn, 2)

	// 开启用户心跳
	u.ClientHeartBeatTime = util.GetTime()
	u.ClientHandshakeAckTime = util.GetTime()
	go listenHeartBeat(u)

	core.Logger.Info("[ClientHandShakeAck]userId:%v", id)

	// 检查用户当前的游戏状态
	// selectserver.Reconnect(u.UserId, u.Version)

	// 开启用户消息队列
	go u.ClientMQ.Start()

	// 检查用户当前状态
	result := selectserver.Reconnect(id, u.Version)
	if result != nil {
		// 用户在房间中
		if result.RoomId > 0 {
			core.Logger.Debug("[ClientHandShakeAck]openGame, id:%v", id)
			openGame(u, "RECONNCET_ROOM")
		}
		// if u.ClientRoomId > 0 && result.RoomId == 0 {
		// 	u.WriteMessageToClient(CloseGamePush(result.Code))
		// }

		// 用户在联赛大厅，自动启动
		core.Logger.Debug("[ClientHandShakeAck]openLeague, id:%v", id)
		if result.RaceId > 0 {
			openLeague(u)
		}
	}

	// 连接俱乐部服务器 暂时自动启动
	// openLeague(u)

	return u, nil
}

// ClientHeartBeat 客户端心跳
func ClientHeartBeat(id int) {
	u := hall.UserMap.Load(id)
	if u == nil {
		core.Logger.Warn("[ClientHeartBeat]user not online,id:%v", id)
		return
	}
	u.ClientHeartBeatTime = util.GetTime()
	u.WriteMessageToClient(response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT))

	// core.Logger.Info("[ClientHeartBeat]userId:%v", id)
}

func listenHeartBeat(u *model.User) {
	for {
		time.Sleep(time.Millisecond * time.Duration(config.HEARTBEAT_INTERVAL))
		user := hall.UserMap.Load(u.UserId)
		if user == nil {
			core.Logger.Debug("[退出心跳]用户已下线, id:%v", u.UserId)
			break
		}

		if user.ClientHandshakeTime != u.ClientHandshakeTime {
			// 用户已经被顶号或者重新登录
			core.Logger.Debug("[退出心跳]用户已重新登录, id:%v", u.UserId)
			break
		}

		if user.ClientHeartBeatTime > 0 {
			if (util.GetTime()-user.ClientHeartBeatTime)*1000 > 2*config.HEARTBEAT_INTERVAL {
				core.Logger.Debug("[退出心跳]用户心跳停止, id:%v", u.UserId)
				// 踢用户下线
				go func() {
					defer util.RecoverPanic()
					hall.KickConn(u.ClientConn)
				}()
				break
			}
		}
	}
}

// 用户重复登录，需将用户踢下线
func repeatLogin(loginedUser *model.User, newConn *net.TCPConn) {
	core.Logger.Warn("[RepeatLogin]userId:%d, new remote:%s", loginedUser.UserId, newConn.RemoteAddr().String())

	// 发送踢下线的协议
	impacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_KICK, nil)
	loginedUser.ClientMQ.Send(impacket)

	// 踢下线
	hall.KickConn(loginedUser.ClientConn)
}

// GameActivate 用户唤醒
func GameActivate(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		return ierror.NewError(-201, id)
	}
	request := fbsCommon.GetRootAsGameActivateRequest(p.GetBody(), 0)
	selectType := string(request.RoomNum())

	result := selectserver.SelectByType(id, u.Version, selectType)

	// 返回选服接口
	responseP := GameActivateResponse(result, p.GetMessageNumber())
	u.WriteMessageToClient(responseP)

	core.Logger.Debug("[GameActivate]userId:%v, result:%v", id, result)

	// 如果用户在房间内，发送 game restore notify
	if result.RoomId > 0 {
		if u.Game != nil && u.Game.IsConnected() {
			core.Logger.Debug("[GameActivate]用户已连接, 发送GameRestoreNotify, userId:%v", id)
			nofity := GameRestoreNotify()
			u.Game.AppendMessage(nofity)
		} else {
			u.ClientRoomId = 0
			u.ClientRound = 0
			u.ClientSeq = 0
			core.Logger.Debug("[GameActivate]用户未连接, openGame, userId:%v", id)
			openGame(u, "RECONNCET_ROOM")
		}
	}

	return nil
}

// CloseClub 客户端关闭俱乐部连接的请求
func CloseClub(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		return ierror.NewError(-201, id)
	}

	if u.Club != nil {
		u.Club.Close()
	} else {
		core.Logger.Warn("[CloseClub]u.Club is nil, userId:%v", id)
	}

	u.WriteMessageToClient(CloseClubPush())

	// 发送客户端俱乐部连接关闭的推送
	return nil
}

// CloseLeague 客户端关闭联赛连接的请求
func CloseLeague(id int, p *protocal.ImPacket) *ierror.Error {
	u := hall.UserMap.Load(id)
	if u == nil {
		return ierror.NewError(-201, id)
	}

	if u.League != nil {
		u.League.Close()
	} else {
		core.Logger.Warn("[CloseLeague]u.League is nil, userId:%v", id)
	}

	u.WriteMessageToClient(CloseLeaguePush())

	// 发送客户端俱乐部连接关闭的推送
	return nil
}
