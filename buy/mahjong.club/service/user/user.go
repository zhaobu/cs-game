package user

import (
	"fmt"
	"net"
	"time"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	"mahjong.club/config"
	"mahjong.club/core"
	"mahjong.club/hall"
	"mahjong.club/ierror"
	"mahjong.club/response"
	"mahjong.club/user"

	simplejson "github.com/bitly/go-simplejson"
)

// HandShake 用户握手
func HandShake(conn *net.TCPConn, impacket *protocal.ImPacket) (*user.User, *ierror.Error) {
	// 从json数据中读出token
	js, _ := simplejson.NewJson(impacket.GetMessage())
	token, err := js.Get("user").Get("token").String()
	if err != nil || len(token) == 0 {
		return nil, ierror.NewError(-10101, "user.HandShake", "token")
	}
	core.Logger.Debug("[HandShake]token:%#v", token)

	// 验证token, 并从token中解析中用户id
	tokenInfo, _ := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	id := int(tokenInfo[0].(float64))
	if id <= 0 {
		return nil, ierror.NewError(-10101, "user.HandShake", "token=>userID")
	}
	core.Logger.Debug("[HandShake]token parse, token:%v, userId:%d", token, id)

	// 判断用户是否已在线
	if connectedUser, online := hall.UserSet.Get(id); online {
		core.Logger.Debug("[HandShake]repeat login, id:%v, old remote:%v, new remote:%v", id, connectedUser.Conn.RemoteAddr().String(), conn.RemoteAddr().String())
		KickUser(connectedUser)
	}

	// 是否需要监听用户心跳
	noHeartBeat, _ := js.Get("user").Get("no_heartbeat").Int()
	// 用户来源
	from, _ := js.Get("user").Get("from").String()

	u := user.NewUser(id)
	u.Conn = conn
	u.HandshakeTime = util.GetTime()
	u.HeartBeatTime = util.GetTime()
	u.NoHeartbeat = noHeartBeat
	u.Info.AvatarBox = GetUserAvatarBox(id)
	u.From = from

	// 读取用户头像框

	hall.UserSet.Add(u)
	core.Logger.Debug("[HandShake]id:%v, user:%#v", id, u)

	// 通知用户握手成功
	jsResponse := simplejson.New()
	jsResponse.Set("heartbeat", core.AppConfig["heartbeat_interval"])
	u.WriteMessage(response.JSONSuccess(protocal.PACKAGE_TYPE_HANDSHAKE, jsResponse))
	return u, nil
}

// HandShakeAck 握手成功
func HandShakeAck(id int, impacket *protocal.ImPacket) *ierror.Error {
	u, online := hall.UserSet.Get(id)
	if !online {
		return ierror.NewError(-202)
	}

	// 开启消息推送
	go u.SendMessage()

	// 开启心跳检测
	if u.NoHeartbeat == 0 {
		go listenHeartBeat(u)
	} else {
		core.Logger.Info("用户无需开启心跳检测, userId:%v, from:%v", u.ID, u.From)
	}

	core.Logger.Info("[HandShakeAck]id:%v", id)
	return nil
}

// HeartBeat 用户心跳
func HeartBeat(id int) {
	if u, online := hall.UserSet.Get(id); online {
		u.HeartBeatTime = util.GetTime()
		// 回应心跳
		u.WriteMessage(response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT))
		// 适时屏蔽
		// core.Logger.Debug("[HeartBeat]id:%v", id)
	} else {
		core.Logger.Warn("[HeartBeat]user not online,id:%v", id)
	}
}

// Logout 用户请求退出
func Logout(id int) *ierror.Error {
	if u, online := hall.UserSet.Get(id); online {
		// 踢除用户
		KickUser(u)
		return nil
	}
	return ierror.NewError(-202)
}

// KickUser 踢出用户
func KickUser(u *user.User) {
	u.QuitOnce.Do(func() {
		u.MStatus = false
		// 将用户从俱乐部移除
		for _, clubID := range u.Clubs {
			hall.RemoveClubUser(clubID, u)
		}

		// 将用户从大厅删除
		hall.UserSet.Del(u.ID)

		// 关闭用户连接
		u.Conn.Close()

		// 关闭用户消息队列
		close(u.Mq)

		core.Logger.Info("[kickUser]id:%v", u.ID)
	})
}

// 监听用户心跳
func listenHeartBeat(u *user.User) {
	// 捕获异常
	defer util.RecoverPanic()

	// 读取心跳间隔
	heartBeatInterval := core.GetAppConfig("heartbeat_interval").(int64)

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
			KickUser(u)
			break
		}
	}
}

// GetUserAvatarBox 获取用户头像框
func GetUserAvatarBox(userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_AVATAR_BOX, userId)
	v, _ := core.RedisDoInt(core.RedisClient0, "get", cacheKey)
	return v
}
