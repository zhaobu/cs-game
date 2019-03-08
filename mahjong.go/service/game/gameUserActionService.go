package game

import (
	"net"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/model"
	"mahjong.go/rank"

	clubModel "mahjong.go/model/club"
	userModel "mahjong.go/model/user"
	configService "mahjong.go/service/config"
	rankService "mahjong.go/service/rank"
	userService "mahjong.go/service/user"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 握手
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func HandShake(conn *net.TCPConn, impacket *protocal.ImPacket) (*User, *core.Error) {
	// 从json数据中读出token
	js, _ := simplejson.NewJson(impacket.GetMessage())
	token, err := js.Get("user").Get("token").String()
	core.Logger.Debug("token:%#v", token)
	if err != nil || len(token) == 0 {
		return nil, core.NewError(-211, js.Get("user"))
	}
	// 读取用户经纬度
	latitude, err := js.Get("user").Get("lat").Float64()
	if err != nil {
		latitude = float64(-1)
	}
	longitude, err := js.Get("user").Get("lng").Float64()
	if err != nil {
		longitude = float64(-1)
	}
	// 读取设备信息
	device, err := js.Get("user").Get("device").String()
	if err != nil {
		device = ""
	}
	deviceToken, err := js.Get("user").Get("device_token").String()
	if err != nil {
		deviceToken = ""
	}
	// 是否无需心跳
	noHeartBeat, _ := js.Get("user").Get("no_heartbeat").Int()
	// 用户来源
	from, _ := js.Get("user").Get("from").String()

	// 验证token, 并从token中解析中用户id
	tokenInfo, _ := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	userId := int(tokenInfo[0].(float64))
	version := tokenInfo[3].(string)
	if userId <= 0 {
		return nil, core.NewError(-212, token)
	}
	core.Logger.Debug("token parse, token:%v, userId:%d", token, userId)

	var clientRoomId int64
	var round, seq int
	clientRoomId, _ = js.Get("user").Get("room_id").Int64()
	round, _ = js.Get("user").Get("round").Int()
	seq, _ = js.Get("user").Get("step").Int()
	loadedRoomId := LoadCurrentRemoteRoomId(userId)
	if clientRoomId > 0 && clientRoomId != loadedRoomId {
		core.Logger.Error("[handshake]error, clientRoomId:%v, loadedRoomId:%v", clientRoomId, loadedRoomId)
		return nil, core.NewError(-1002)
	}

	// 如果是重复登录，需要关闭之前的连接
	loginedUser, _ := UserMap.GetUser(userId)
	if loginedUser != nil && loginedUser.Conn != conn {
		RepeatLogin(loginedUser, conn)
	}

	// 获取用户信息
	userData := userService.GetUser(userId)
	if userData.UserId == 0 {
		return nil, core.NewError(-213, userId)
	}
	// 获取用户扩展数据
	userInfoList := userService.GetUserInfoList(userId)

	// 新建大厅用户
	user := NewUser(userId, conn)
	user.ConnectTime = time.Now().UnixNano()

	// 填充基本数据
	info := &UserInfo{}
	info.LastHeartBeatTime = util.GetTime()
	info.Nickname = userData.Nickname
	info.Avatar = userData.IconUrl
	if !configService.IsRobot(userId) {
		info.Avatar = userService.GetUserAvatar(userId)
	}
	info.Gender = userService.GetGender(userInfoList)
	// 填充位置信息

	info.Ip = util.GetIP(conn)
	info.Area = userService.GetCity(userInfoList)
	// 填充地区、经纬度
	info.Longitude = longitude
	info.Latitude = latitude
	// 填充用户历史积分
	info.Score = userService.GetScore(userInfoList)
	info.ScoreMatch = userService.GetScoreMatch(userInfoList)
	info.ScoreRandom = userService.GetScoreRandom(userInfoList)
	info.ScoreCoin = userService.GetScoreCoin(userInfoList)
	// 用户城市
	info.RankCity = userService.GetGameCity(userInfoList)
	// 填充用户排位赛数据
	season := model.GetSeason()
	if season != nil && season.IsOpen() {
		seasonUser := rankService.GetSeasonUser(userId, season.Id)
		if seasonUser != nil && seasonUser.UserId > 0 {
			info.ScoreRank = rank.FormatSLevel(seasonUser.GradeId, seasonUser.GradeLevel, seasonUser.StarNum)
			info.RankCity = seasonUser.LastCity
			info.RankExp = seasonUser.Exp
			info.RankConsume = rankService.GetConsumeCard(season.Id, seasonUser.GradeId)
		}
	}
	// 填充用户钻石
	info.Money = userModel.GetMoney(userData)
	// 填充设备信息
	info.Device = device
	info.DeviceToken = deviceToken
	// 版本信息
	info.Version = version
	// 载入用户已加入的房间id
	user.RoomId = loadedRoomId

	// 记录握手时的重连参数
	user.handshakeRoomId = clientRoomId
	user.handshakeRound = round
	user.handshakeSeq = seq

	// 是否无需心跳的用户
	user.NoHeartbeat = noHeartBeat
	// 用户来源
	user.From = from

	// 用户头像框
	info.AvatarBox = userService.GetUserAvatarBox(userId)
	// 用户会员等级
	info.MemberLevel, info.MemberAddExp = userService.GetUserMemberLevel(userId)

	core.Logger.Debug("用户扩展信息, userId:%d, info:%#v", userId, info)

	// 添加至用户map
	user.Info = info
	UserMap.SetUser(user)

	// 构造心跳返回参数
	responseMap := make(map[string]interface{})
	responseMap["heartbeat"] = config.HEART_BEAT_SECOND
	responseMap["roomid"] = 0
	responseMap["roomnum"] = ""

	var room *Room
	if user.RoomId > 0 {
		room, _ = RoomMap.GetRoom(user.RoomId)
		responseMap["roomid"] = room.RoomId
		responseMap["roomnum"] = room.Number
	}
	// 握手成功, 发送消息通知客户端
	// 此消息直接发送，不由消息队列处理
	user.MQ.Send(HandShakeResponse(responseMap))

	// 处理用户在房间内的逻辑
	if user.RoomId > 0 {
		// 更新用户的俱乐部积分
		if room.IsClubMatch() {
			clubUser := clubModel.GetClubUser(room.ClubId, user.UserId)
			user.Info.ScoreClub = clubUser.Score
		} else if room.IsLeague() && room.RaceInfo != nil {
			user.Info.ScoreLeague = model.GetRaceUserScore(room.RaceInfo.Id, userId)
		}
		// 重写用户在房间中的参数
		if room.IsLeague() && configService.IsRobot(user.UserId) {
			// 机器人不更新
			ru := room.GetUser(user.UserId)
			if ru != nil {
				user.Info.Avatar = ru.Info.Avatar
				user.Info.Nickname = ru.Info.Nickname
			}
		}

		room.UpdateUser(user)

	}

	core.Logger.Info("[handShake]userId:%d,roomId:%v, remote:%v, user remote:%v", userId, user.RoomId, GetRemoteAddr(), conn.RemoteAddr().String())
	return user, nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 握手成功
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func HandShakeAck(userId int) {
	user, err := UserMap.GetUser(userId)
	if err != nil {
		// 用户不在线
		return
	}

	core.Logger.Info("[HandShakeAckAction]userId:%d, roomId:%v", userId, user.RoomId)

	if user.RoomId > 0 {
		// 用户在房间中，需要通知用户进行重连
		room, _ := RoomMap.GetRoom(user.RoomId)
		if room != nil {
			if room.canRestoreSection(user, user.handshakeRoomId, user.handshakeRound, user.handshakeSeq) {
				room.restoreSection(user, user.handshakeSeq)
			} else {
				room.restoreIntact(user)
			}

			// 如果用户在房间内，推送用户下线的消息
			if !user.EnableRestoreDone() {
				if !room.Ob.hasUser(userId) {
					onlinePush := UserOnlinePush(userId, 1)
					room.SendMessageToRoomUser(onlinePush, userId)
					room.Ob.sendMessage(onlinePush, 0)

					// 记录重连日志
					restoreLog(userId, room.RoomId, room.Number, room.Round, config.RESTORE_LOG_TYPE_RECONNECT)
				}
			}
		}
	}

	// 开启消息发送队列
	// 如果用户支持restore done的话，需要在restore done 中开启
	if !user.EnableRestoreDone() || user.RoomId == 0 {
		go user.MQ.Start()
	}

	// 设置用户消息状态，可接受消息
	user.messageStatus = true

	// 开启心跳检测
	if user.NoHeartbeat == 0 {
		go user.ListenHeartBeat()
	}
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 退出
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func Kick(userId int) (bool, *core.Error) {
	user, err := UserMap.GetUser(userId)
	if err != nil {
		// 用户不在线，不做任何操作
		return false, err
	}
	// 从列表删除
	err = UserMap.DelUser(user.UserId)
	if err != nil {
		// 用户不在线，不做任何操作
		return false, err
	}

	user.KickOnce.Do(func() {
		// 关闭用户消息队列
		user.MQ.Close()

		// 关闭用户连接
		user.Conn.Close()

		// 如果用户在房间内，推送用户下线的消息
		if user.RoomId > 0 {
			if room, err := RoomMap.GetRoom(user.RoomId); err == nil {
				if !room.Ob.hasUser(userId) {
					onlinePush := UserOnlinePush(userId, 0)
					room.SendMessageToRoomUser(onlinePush, userId)
					room.Ob.sendMessage(onlinePush, 0)

					// 记录用户的断线日志
					restoreLog(user.UserId, room.RoomId, room.Number, room.Round, config.RESTORE_LOG_TYPE_KICK)
				}
			}
		}

		core.Logger.Info("[kick]userId:%d, roomId:%v, connectTime:%v, device:%v, deviceToken:%v, remote:%v.", userId, user.RoomId, util.FormatUnixTime(user.ConnectTime), user.Info.Device, user.Info.DeviceToken, user.Conn.RemoteAddr().String())
	})
	return true, err
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 心跳
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func HeartBeat(userId int) {
	// 更新心跳时间
	user, err := UserMap.GetUser(userId)
	if err == nil {
		// 更新心跳时间
		user.Info.LastHeartBeatTime = util.GetTime()

		// 返回一个心跳消息
		// 这里必须要直接推消息给用户，不能由推送队列发送，原因有二：
		// 1、消息队列可能有堆积，会导致心跳超时，被客户端断开连接
		// 2、在用户重连完成之前，消息队列是不开的, 如果重连时间过长，会导致心跳超时
		user.MQ.Send(HeartBeatResponse())
	}
	// 消息量太大，注释掉
	if HeartBeatLogFlag {
		core.Logger.Debug("用户心跳,userId:%v, time:%v, timestamp:%v.", userId, util.GetTime(), util.GetTimestamp())
	}
}
