package service

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/protocal"
	"mahjong-league/response"
	"mahjong-league/user"
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

// HandShake 用户握手逻辑
func HandShake(conn *net.TCPConn, impacket *protocal.ImPacket) (*user.User, *ierror.Error) {
	// 从json数据中读出token
	js, _ := simplejson.NewJson(impacket.GetMessage())
	token, err := js.Get("user").Get("token").String()
	if err != nil || len(token) == 0 {
		return nil, ierror.NewError(-10101, "user.HandShake", "token", token)
	}
	// 是否无需心跳
	noHeartBeat, err := js.Get("user").Get("no_heartbeat").Int()
	// 用户来源
	from, err := js.Get("user").Get("from").String()
	core.Logger.Debug("[HandShake]token:%#v", token)

	// 验证token, 并从token中解析中用户id
	tokenInfo, _ := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	id := int(tokenInfo[0].(float64))
	version := tokenInfo[3].(string)
	if id <= 0 {
		return nil, ierror.NewError(-10101, "user.HandShake", "token=>userID", id)
	}
	core.Logger.Debug("[HandShake]token parse, token:%v, userId:%d", token, id)

	// 判断用户是否已在线
	if connectedUser, online := hall.UserSet.Get(id); online {
		core.Logger.Debug("[HandShake]repeat login, id:%v, old remote:%v, new remote:%v", id, connectedUser.Conn.RemoteAddr().String(), conn.RemoteAddr().String())
		hall.KickConn(connectedUser.Conn)
	}

	u := user.NewUser(id, conn)
	u.HandshakeTime = util.GetTime()
	u.HeartBeatTime = util.GetTime()
	u.NoHeartbeat = noHeartBeat
	u.From = from
	u.Version = version
	hall.UserSet.Add(u)

	// 删除handshake超时监听
	c, ok := hall.WaitConnectionSet.Load(conn.RemoteAddr().String())
	if ok {
		select {
		case c.(chan int) <- 2:
		default:
			core.Logger.Warn("[handshake]wait connection channel 删除失败,userID:%v, remote:%v", id, conn.RemoteAddr().String())
		}
	}

	// 存储连接=>id对照关系
	hall.ConnectionSet.Store(u.Conn.RemoteAddr().String(), u.ID)

	// 返回用户当前的房间
	var roomId int64
	raceId := model.GetUserRace(u.ID)
	if raceId > 0 {
		raceUserInfo := model.GetRaceUserInfo(raceId, u.ID)
		roomId = raceUserInfo.RoomId
	}

	// 通知用户握手成功
	jsResponse := simplejson.New()
	jsResponse.Set("heartbeat", config.HEART_BEAT_SECOND)
	jsResponse.Set("roomid", roomId)
	u.WriteMessage(response.JSONSuccess(protocal.PACKAGE_TYPE_HANDSHAKE, jsResponse))

	return u, nil
}

// HandShakeAck 握手成功
func HandShakeAck(id int, impacket *protocal.ImPacket) (*user.User, *ierror.Error) {
	u, online := hall.UserSet.Get(id)
	if !online {
		return nil, ierror.NewError(-202)
	}

	u.HandshakeAckTime = util.GetTime()

	// 开启消息推送
	go u.Mq.Start()

	// 开启心跳检测
	if u.NoHeartbeat == 0 {
		go listenHeartBeat(u)
	}

	core.Logger.Info("[HandShakeAck]id:%v", id)

	// 推送比赛列表和比赛信息
	// 延迟发送
	time.Sleep(time.Millisecond * 10)
	PushAfterHandshake(u)
	return u, nil
}

// HeartBeat 用户心跳
func HeartBeat(id int) {
	if u, online := hall.UserSet.Get(id); online {
		u.HeartBeatTime = util.GetTime()
		// 回应心跳
		u.WriteMessage(response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT))
	} else {
		core.Logger.Warn("[HeartBeat]user not online,id:%v", id)
	}
}

// 监听用户心跳
func listenHeartBeat(u *user.User) {
	// 捕获异常
	defer util.RecoverPanic()

	// 读取心跳间隔
	heartBeatInterval := int64(config.HEART_BEAT_SECOND)

	for {
		time.Sleep(time.Second * time.Duration(heartBeatInterval))

		user, online := hall.UserSet.Get(u.ID)
		if !online {
			core.Logger.Debug("用户已下线，停止心跳监测, id:%v", u.ID)
			break
		}

		if user.HandshakeTime != u.HandshakeTime {
			// 用户已经被顶号或者重新登录
			core.Logger.Debug("用户已重新登录，停止心跳监测, id:%v", u.ID)
			break
		}

		if util.GetTime()-user.HeartBeatTime > 2*heartBeatInterval {
			core.Logger.Debug("用户心跳停止，踢下线, id:%v", u.ID)
			hall.KickConn(u.Conn)
			break
		}
	}
}
