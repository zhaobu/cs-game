package game

import (
	"encoding/json"
	"net"

	"github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/model"

	fbsCommon "mahjong.go/fbs/Common"
	clubModel "mahjong.go/model/club"
	userModel "mahjong.go/model/user"
	configService "mahjong.go/service/config"
	hallService "mahjong.go/service/hall"
	raceService "mahjong.go/service/race"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

// 获取房间信息
func GetRoomInfo(conn *net.TCPConn, roomId int64) *core.Error {
	// 读取房间信息
	room, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	userList := make([]map[string]interface{}, 0, room.GetUsersLen())

	for _, u := range room.GetUsers() {
		userInfo := map[string]interface{}{"userId": u.UserId, "avatar": u.Info.Avatar, "nickname": u.Info.Nickname, "index": u.Index}
		if !configService.IsRobot(u.UserId) {
			userInfo["avatar"] = userService.GetUserAvatar(u.UserId)
		}
		userList = append(userList, userInfo)
	}

	data["setting"] = room.setting.GetSetting()
	data["users"] = userList
	data["gType"] = room.MType
	data["number"] = room.Number
	data["round"] = room.TRound
	data["clubId"] = room.ClubId

	conn.Write(SystemResponse(data).Serialize())

	core.Logger.Debug("[GetRoomInfo][H5]roomId:%d", roomId)

	return nil
}

// 获取游戏明细
func GetGameDetail(conn *net.TCPConn, roomId int64) *core.Error {
	// 读取房间信息
	room, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	// 游戏类型
	data["gType"] = room.MType
	// 游戏局数
	data["totalRound"] = room.TRound
	// 当前局数
	data["round"] = room.Round
	// 游戏设置
	data["setting"] = room.setting
	// 庄家
	data["dealer"] = room.Dealer
	// 连庄数
	data["dealCount"] = room.DealCount
	// 准备状态
	data["ready"] = room.Ready
	// 已准备用户列表
	data["readyList"] = room.ReadyList
	// 房间解散状态
	data["dismiss"] = room.DismissTime
	// 已同意解散用户
	data["dismissOp"] = room.GetDismissUsers()
	// 需回应用户列表
	waitUsers := []int{}
	// 游戏中的数据
	if room.MI != nil {
		// 冲锋鸡用户id
		data["chikenChargeBam1"] = room.MI.getChikenChargeBam1()
		// 冲锋乌骨鸡用户id
		data["chikenChargeDot8"] = room.MI.getChikenChargeDot8()
		// 责任鸡用户id
		data["chikenResponsibility"] = room.MI.getChikenResponsibility()
		// 前后鸡牌面
		data["chikenFB"] = room.MI.getChikenFBTile()

		// 用户信息
		userList := make([]map[string]interface{}, 0, room.GetUsersLen())
		mUsers := room.MI.getUsers()
		for _, u := range room.GetUsers() {
			// 基础数据
			userInfo := map[string]interface{}{"userId": u.UserId, "avatar": u.Info.Avatar, "nickname": u.Info.Nickname}
			if !configService.IsRobot(u.UserId) {
				userInfo["avatar"] = userService.GetUserAvatar(u.UserId)
			}
			// 用户手牌
			userInfo["handTiles"] = mUsers[u.UserId].HandTileList.ToSlice()
			// 用户明牌
			showTiles := make([]map[string]interface{}, 0)
			for _, showCard := range mUsers[u.UserId].ShowCardList.GetAll() {
				showTileInfo := make(map[string]interface{})
				showTileInfo["sType"] = showCard.GetOpCode()  // 明牌类型：吃、碰、杠、胡
				showTileInfo["free"] = showCard.IsFree()      // 是否收费
				showTileInfo["target"] = showCard.GetTarget() // 对象，碰了谁的， 吃了谁的
				showTileInfo["tiles"] = showCard.GetTiles()   // 牌
				showTiles = append(showTiles, showTileInfo)
			}
			userInfo["showTiles"] = showTiles
			// 用户弃牌
			userInfo["disCardTiles"] = mUsers[u.UserId].HandTileList.ToSlice()
			// 用户可进行的操作
			waitInfo := room.MI.getWait(u.UserId)
			userInfo["opList"] = waitInfo
			if waitInfo != nil && len(waitInfo.OpList) > 0 && waitInfo.Reply == nil {
				waitUsers = append(waitUsers, u.UserId)
			}
			userList = append(userList, userInfo)
		}
		data["userList"] = userList
		data["waitUsers"] = waitUsers
	}

	conn.Write(SystemResponse(data).Serialize())

	core.Logger.Debug("[GetGameDetail]roomId:%d", roomId)

	return nil
}

// 输出游戏统计数据
func Stat(conn *net.TCPConn, v int) {
	if v == 0 {
		Cre = false
		Coe = false
		Rne = false
	} else {
		Cre = true
		Coe = true
		Rne = true
	}
	statInfo := make(map[string]interface{})
	statInfo["cre"] = Cre
	statInfo["coe"] = Coe
	statInfo["rne"] = Rne

	conn.Write(SystemResponse(statInfo).Serialize())

	core.Logger.Debug("[GetGameDetail]%#v", statInfo)
}

// 将用户直接塞入房间
func H5JoinRoom(conn *net.TCPConn, js *simplejson.Json) *core.Error {
	// 解析用户数据
	token, err := js.Get("user").Get("token").String()
	core.Logger.Debug("token:%v", token)
	if err != nil || len(token) == 0 {
		return core.NewError(-211, js.Get("user"))
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
	// 读取ip
	ip, err := js.Get("user").Get("ip").String()
	if err != nil {
		ip = ""
	}
	// 读取房间号
	number, err := js.Get("number").String()
	if err != nil {
		return core.NewError(-319)
	}

	// 验证token, 并从token中解析中用户id和
	tokenInfo, _ := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	userId := int(tokenInfo[0].(float64))
	version := tokenInfo[3].(string)
	if userId <= 0 {
		return core.NewError(-212, token)
	}

	// 判断用户是否已经在房间内
	if roomId := LoadRoomId(userId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}
	/*
		if roomId := userService.GetRoomId(userId); roomId > 0 {
			core.Logger.Debug("userId:%v, roomId:%v", userId, roomId)
			// 这里可能有多条线的问题，所以需要判断房间是否真的存在
			if roomRemote := roomService.GetRoomRemote(roomId); roomRemote != "" {
				core.Logger.Debug("roomId:%v, roomRemote:%v", roomId, roomRemote)
				if hallService.IsRoomExists(roomRemote, roomId) {
					core.Logger.Debug("IsRoomExists, roomId:%v", roomId)
					return core.NewError(-202, roomService.GetRoomNumberById(roomId))
				}
			}
		}*/

	// 读取房间编号对应的房间id
	roomId := roomService.GetRoomIdByNumber(number)
	if roomId == 0 {
		return core.NewError(-302, number)
	}
	room, coreErr := RoomMap.GetRoom(roomId)
	if coreErr != nil {
		return coreErr
	}
	// 获取用户的raceId
	raceId := raceService.GetUserRace(userId)
	if raceId > 0 {
		// 如果用户在比赛中
		// 读取raceInfo
		raceInfo := model.GetRace(raceId)
		if raceInfo != nil {
			if raceInfo.Status == 1 || // 进行中
				raceInfo.Status == 2 { // 结算中
				// 比赛正在进行中，不能加入房间
				return core.NewError(-337)
			} else if raceInfo.Status == 0 {
				// 报名中的比赛，如果是定时塞，且不在两分钟内，才可以加入
				if raceInfo.StartTime > 0 && (raceInfo.StartTime-util.GetTime() > 120) {
					// do nothing
				} else {
					return core.NewError(-338)
				}
			}
		}
	}

	// 只有自主组建的房间，才能通过H5加入
	if !room.EnableJoin() {
		return core.NewError(-320)
	}
	if room.IsClub() && room.ClubId > 0 && !clubModel.IsClubUser(room.ClubId, userId) {
		return core.NewError(-334, userId, room.ClubId)
	}
	var clubUser *config.ClubUser
	if room.IsClub() {
		clubUser = clubModel.GetClubUser(room.ClubId, userId)
		// 判断用户是否在俱乐部中
		if clubUser.ClubId == 0 {
			return core.NewError(-334, userId, room.ClubId)
		}
		// 判断用户俱乐部积分是否足够
		if room.IsClubMatch() {
			if clubUser.Score < room.ClubScore {
				return core.NewError(-335, userId, room.ClubId, room.ClubScore, clubUser.Score)
			}
		}
	}

	// 获取用户信息
	userData := userService.GetUser(userId)
	if userData.UserId == 0 {
		return core.NewError(-213, userId)
	}
	// 获取用户扩展数据
	userInfoList := userService.GetUserInfoList(userId)

	// 填充基本数据
	info := &UserInfo{}
	info.Nickname = userData.Nickname
	info.Avatar = userData.IconUrl
	if !configService.IsRobot(userId) {
		info.Avatar = userService.GetUserAvatar(userId)
	}
	info.Gender = userService.GetGender(userInfoList)
	// 填充位置信息
	info.Ip = ip
	info.Area = userService.GetCity(userInfoList)
	// 填充地区、经纬度
	info.Longitude = longitude
	info.Latitude = latitude
	// 填充用户历史积分
	info.Score = userService.GetScore(userInfoList)
	info.ScoreMatch = userService.GetScoreMatch(userInfoList)
	info.ScoreRandom = userService.GetScoreRandom(userInfoList)
	if room.IsClubMatch() {
		info.ScoreClub = clubUser.Score
	}
	info.ScoreCoin = userService.GetScoreCoin(userInfoList)
	// 填充用户钻石
	info.Money = userModel.GetMoney(userData)
	// 填充设备信息
	info.Device = device
	info.DeviceToken = deviceToken
	// 版本信息
	info.Version = version
	// 用户头像框
	info.AvatarBox = userService.GetUserAvatarBox(userId)
	// 用户会员等级
	info.MemberLevel, info.MemberAddExp = userService.GetUserMemberLevel(userId)

	// 新建大厅用户
	user := NewUser(userId, nil)
	user.Info = info

	core.Logger.Debug("[JoinRoom][h5]用户扩展信息, userId:%d, roomId:%v, info:%#v", userId, user.RoomId, info)

	// 检查客户端版本
	if !room.checkJoinVersion(user.Info.Version) {
		core.Logger.Debug("[JoinRoom][h5]用户版本号过低,userId:%v, roomId:%v, version:%v", userId, user.RoomId, user.Info.Version)
		return core.NewError(-321)
	}

	// 防止并发，加锁
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 判断房间是否已结束
	if room.CheckStatus(config.ROOM_STATUS_COMPLETED) {
		return core.NewError(-302, number)
	}

	// 判断房间是否已满
	if room.IsFull() {
		return core.NewError(-303, number)
	}

	// 将用户加入房间
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 更新房间列表
	// RoomMap[room.RoomId] = room
	RoomMap.SetRoom(room)

	// 给其他成员发送有人加入的push
	if room.GetUsersLen() > 1 {
		roomUserInfo := room.GetUser(userId)
		pushPacket := JoinRoomPush(roomUserInfo)
		room.SendMessageToRoomUser(pushPacket, userId)

		// 发送push
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)
	}

	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if room.IsFull() {
		// 房间开始
		room.enter()
	}

	conn.Write(SystemResponse(nil).Serialize())

	// 如果是俱乐部房间，推送一个新建俱乐部房间的消息到俱乐部服务
	if room.IsClub() {
		CPool.appendMessage(ClubG2CJoinRoomPush(room.ClubId, room.RoomId, room.GetUser(userId)))
	}

	// 删除用户的未读房间结果标志
	userService.RemoveRoomResultUnread(userId)

	// 如果当前用户在线，需要将用户踢下线，防止用户数据异常
	if UserMap.IsUserExists(userId) {
		Kick(userId)
		core.Logger.Debug("[H5JoinRoom]用户在线状态下通过h5加入了房间，将用户连接断开，重新进行连接,userId:%v,roomId:%v,number:%v", userId, room.RoomId, room.Number)
	}

	core.Logger.Info("[JoinRoom][h5]userId:%d, roomId:%d,number:%v", userId, roomId, room.Number)

	return nil
}

// H5CreateRoom 麻将馆主创建房间
func H5CreateRoom(conn *net.TCPConn, js *simplejson.Json) *core.Error {
	// 解析参数
	userId, _ := js.Get("userId").Int()
	clubId, _ := js.Get("clubId").Int()
	gType, _ := js.Get("gType").Int()
	round, _ := js.Get("round").Int()
	settingInterface, _ := js.Get("setting").Array()
	// 参数校验
	if userId == 0 {
		return core.NewError(-6, "system.H5CreateRoom", "userId", userId)
	}
	if clubId == 0 {
		return core.NewError(-6, "system.H5CreateRoom", "clubId", clubId)
	}
	// 读取游戏类型
	if _, isExists := fbsCommon.EnumNamesGameType[gType]; !isExists {
		return core.NewError(-500, gType)
	}
	// 读取局数
	if _, roundExists := config.MahjongRoundPrice[round]; !roundExists {
		return core.NewError(-501, round)
	}
	// 扩展玩法
	setting := []int{}
	for _, i := range settingInterface {
		v, _ := i.(json.Number)
		val, _ := v.Int64()
		setting = append(setting, int(val))
	}
	// 读取麻将馆信息
	club := clubModel.GetClub(clubId)
	if club.Id == 0 {
		return core.NewError(-600, clubId)
	}
	// 创建房间的类型
	var cType = config.ROOM_TYPE_CLUB
	if club.EnableOut == 1 {
		cType = config.ROOM_TYPE_CLUB_MATCH
	}
	// 判断金额是否足够
	payType := config.ROOM_PAY_TYPE_CLUB
	price := configService.GetGamePrice(gType, config.ROOM_TYPE_CREATE, round)
	if price > 0 && club.Fund < price {
		return core.NewError(-204, price)
	}
	// 创建房间
	room := Create(userId, gType, round, cType, setting, config.ROOM_CREATE_MODE_CLUB)
	// 判断setting合理性
	if err := room.verifySetting(); err != nil {
		return err
	}
	// 保存房间号与房间id的对应关系，这里因为并发的原因，可能会失败
	if !roomService.SaveRoom(room.RoomId, room.Number, GetRemoteAddr()) {
		return core.NewError(-301, room.Number)
	}
	// 记录房间的俱乐部id和付费类型
	room.PayType = payType
	room.PayPrice = price
	room.ClubId = clubId
	room.ClubCapitorUserId = club.ManageUser
	if room.IsClubMatch() {
		room.ClubScore = club.Score
	}
	go listenRoomTimeout(room.RoomId)
	hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)
	// 加入房间列表
	RoomMap.SetRoom(room)
	// 推送俱乐部消息到俱乐部服务
	CPool.appendMessage(ClubG2CReloadRoomPush(room.ClubId, room))
	// 写入消息推送队列
	roomService.AppendClubCreateRoomPush(room.ClubId, room.RoomId, room.Number, room.Creator, room.setting.GetSetting())
	// 返回一条response
	data := make(map[string]interface{})
	data["roomId"] = room.RoomId
	data["number"] = room.Number
	conn.Write(SystemResponse(data).Serialize())
	core.Logger.Info("[CreateRoom][h5]userId:%d, roomId:%d, number:%s, clubId:%v, remote:%v", userId, room.RoomId, room.Number, clubId, GetRemoteAddr())
	return nil
}

// H5DismissRoom 麻将馆主解散房间
func H5DismissRoom(conn *net.TCPConn, js *simplejson.Json) *core.Error {
	// 解析参数
	userId, _ := js.Get("userId").Int()
	clubId, _ := js.Get("clubId").Int()
	roomId, _ := js.Get("roomId").Int64()

	/*
		if userId == 0 {
			return core.NewError(-6, "system.H5DismissRoom", "userId", userId)
		}
		if clubId == 0 {
			return core.NewError(-6, "system.H5DismissRoom", "clubId", clubId)
		}
	*/
	if roomId == 0 {
		return core.NewError(-6, "system.H5DismissRoom", "roomId", clubId)
	}
	// 读取房间信息
	room, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}
	/*
		if room.ClubId != clubId || !room.IsClubMode() {
			return core.NewError(-704, roomId)
		}
	*/
	room.Mux.Lock()
	defer room.Mux.Unlock()
	/*
		if !room.CheckStatus(config.ROOM_STATUS_CREATING) {
			return core.NewError(-705, roomId)
		}
	*/
	dismissRoom(room, config.DISMISS_ROOM_CODE_CLUB_DISMISS)
	// 返回一条response
	data := make(map[string]interface{})
	conn.Write(SystemResponse(data).Serialize())
	core.Logger.Info("[DismissRoom][h5]userId:%d, roomId:%d, number:%s, clubId:%v, remote:%v", userId, room.RoomId, room.Number, clubId, GetRemoteAddr())
	return nil
}

// HeartBeatFlag 设置心跳日志是否开启
func HeartBeatFlag(conn *net.TCPConn, js *simplejson.Json) *core.Error {
	f, _ := js.Get("f").Int()
	if f > 0 {
		HeartBeatLogFlag = true
	} else {
		HeartBeatLogFlag = false
	}
	core.Logger.Info("[HeartBeatFlag]f:%v", f)
	return nil
}
