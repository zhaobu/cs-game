package game

import (
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/robot"

	fbsCommon "mahjong.go/fbs/Common"
	clubModel "mahjong.go/model/club"
	coinService "mahjong.go/service/coin"
	configService "mahjong.go/service/config"
	hallService "mahjong.go/service/hall"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 加入比赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func MatchJoin(userId int, gType int, mNumber uint16) *core.Error {
	// 检测房间类型是否正确
	if !configService.CheckMatchRoomGameType(gType) {
		return core.NewError(-324, userId, gType)
	}

	// 读取游戏配置
	setting := configService.GetMatchRoomDefaultSetting(gType)
	if setting == nil {
		return core.NewError(-325, userId, gType)
	}
	// 老版本的配置，没有本鸡, 这里兼容一下
	if len(setting) == 11 {
		setting = append(setting, 0)
	}
	core.Logger.Debug("[MatchJoin]userId:%v,gType:%v,setting:%v", userId, gType, setting)

	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	// 读取比赛房间队列
	queue := MatchRoomQueueMap.Get(gType)

	// 判断金额是否足够
	price := configService.GetGamePrice(gType, config.ROOM_TYPE_MATCH, config.ROOM_MATCH_ROUND)
	if price > 0 && !userService.CheckMoneyEnough(userId, price, nil) {
		return core.NewError(-204, price)
	}

	// 加锁
	queue.Mux.Lock()
	defer queue.Mux.Unlock()

	// 找出是否有可加入的队伍
	var roomId int64
	if len(queue.Rooms) > 0 {
		for rId, _ := range queue.Rooms {
			// 如果room不存在，说明数据错误，修复数据
			room, err := RoomMap.GetRoom(rId)
			if err != nil {
				core.Logger.Warn("roomId存在于比赛队列中，但是在房间列表未找到，roomId: %d", roomId)
				delete(queue.Rooms, rId)
				continue
			}

			// 如果房间人数已满或者已开始，说明说句错误，需要修复
			if room.IsFull() || !room.CheckStatus(config.ROOM_STATUS_CREATING) {
				core.Logger.Warn("房间已满或者已开始，roomId: %d", roomId)
				delete(queue.Rooms, rId)
				continue
			}

			// 找到房间
			roomId = rId
			break
		}
	}
	core.Logger.Debug("find match room, userId:%v, roomId:%#v", userId, roomId)

	// 这里加一个判断，机器人不能自己创建房间
	if roomId == 0 && configService.IsRobot(userId) {
		return core.NewError(-322)
	}

	// 未找到可以加入的房间，创建一个房间
	var room *Room
	if roomId == 0 {
		room = Create(userId, gType, config.ROOM_MATCH_ROUND, config.ROOM_TYPE_MATCH, setting, config.ROOM_CREATE_MODE_USER)
		room.PayPrice = price

		// 保存房间号与房间id的对应关系，这里因为并发的原因，可能会失败
		if !roomService.SaveRoom(room.RoomId, room.Number, GetRemoteAddr()) {
			return core.NewError(-301, room.Number)
		}
		// 监听房间超时
		go listenRoomTimeout(room.RoomId)
		// 监听房间准备超时
		go listenRoomKickUser(room.RoomId)
		// 写入队列
		queue.Rooms[room.RoomId] = room.Number
		// 记录大厅房间列表
		hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)
		// 将这个随机房间的id，记入robot room list
		if room.EnableRobot() {
			robotGameInfo := robot.NewGameInfo(GetRemoteAddr(), room.RoomId, room.MType, room.CType, room.setting.GetSettingPlayerCnt(), room.TRound)
			hallService.AddHallRobotRoom(GetRemoteAddr(), robotGameInfo.String())
		}
	} else {
		room, _ = RoomMap.GetRoom(roomId)
	}

	// 这里需要加锁，防止并发退出
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 如果是机器人，这里重写一下机器人的随机积分，防止出错
	if configService.IsRobot(userId) {
		switch gType {
		case fbsCommon.GameTypeMAHJONG_MATCH_GZ_1:
			user.Info.ScoreMatch = util.RandIntn(200)
		case fbsCommon.GameTypeMAHJONG_MATCH_GZ_2:
			user.Info.ScoreMatch = 200 + util.RandIntn(300)
		case fbsCommon.GameTypeMAHJONG_MATCH_GZ_3:
			user.Info.ScoreMatch = 1000 + util.RandIntn(1000)
		case fbsCommon.GameTypeMAHJONG_MATCH_GZ_4:
			user.Info.ScoreMatch = 2000 + util.RandIntn(5000)
		default:
			break
		}
		core.Logger.Debug("[matchRoom]设置机器人积分,userId:%v, score:%v", userId, user.Info.ScoreMatch)
	}

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 因为并发问题，只有房间是新创建的，才将房间数据添加到RoomMap
	if roomId == 0 {
		RoomMap.SetRoom(room)
	}

	// 给用户回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(room, mNumber))

	// 给其他成员发送有人加入的push
	if room.GetUsersLen() > 1 {
		core.Logger.Debug("给其他人发用户加入push")
		roomUserInfo := room.GetUser(userId)
		responsePacket := JoinRoomPush(roomUserInfo)
		room.SendMessageToRoomUser(responsePacket, userId)

		// 发送push
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)
	}

	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if room.IsFull() {
		// 从队列移除
		delete(queue.Rooms, room.RoomId)

		// 房间开始
		room.enter()
	}

	core.Logger.Info("[MatchJoin]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 随机组队
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func RandomJoin(userId int, gType int, mNumber uint16) *core.Error {
	// 如果传入了错误的房间类型，自动定义成贵阳麻将
	if !configService.CheckRandomRoomGameType(gType) {
		gType = fbsCommon.GameTypeMAHJONG_GY
	}

	// 默认支持连庄、满堂鸡
	setting := getRandomSetting(userId, gType)
	core.Logger.Debug("[RandomJoin]random room setting,userId:%v,gType:%v,setting:%v", userId, gType, setting)

	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	// 读取组队队列
	queue := RandomRoomQueueMap.Get(gType)

	// 判断是否是正确的麻将类型
	if _, isExists := fbsCommon.EnumNamesGameType[gType]; !isExists {
		return core.NewError(-500, gType)
	}

	// 判断金额是否足够
	var price int
	if !userService.IsRandomPayed(userId) {
		price = configService.GetGamePrice(gType, config.ROOM_TYPE_RAND, config.ROOM_RANDOM_ROUND)
		if price > 0 && !userService.CheckMoneyEnough(userId, price, nil) {
			return core.NewError(-204, price)
		}
	}

	// 加锁
	queue.Mux.Lock()
	defer queue.Mux.Unlock()

	// 找出是否有可加入的队伍
	var roomId int64
	if len(queue.Rooms) > 0 {
		for rId, _ := range queue.Rooms {
			// 如果room不存在，说明数据错误，修复数据
			room, err := RoomMap.GetRoom(rId)
			if err != nil {
				core.Logger.Warn("[RandomJoin]roomId存在于随机队列中，但是在房间列表未找到，roomId: %d", rId)
				delete(queue.Rooms, rId)
				continue
			}

			// 如果房间人数已满或者已开始，说明说句错误，需要修复
			if room.IsFull() || !room.CheckStatus(config.ROOM_STATUS_CREATING) {
				core.Logger.Warn("[RandomJoin]房间已满或者已开始，roomId: %d", roomId)
				delete(queue.Rooms, rId)
				continue
			}

			// 找到房间
			roomId = rId
			break
		}
	}

	core.Logger.Debug("[RandomJoin]find room, userId:%v, roomId:%#v", userId, roomId)

	// 如果服务器设置了不允许加入其它人的随机房间
	// 则将真人的搜索到的随机房间设置为0
	// 让用户创建一个房间, 随后由机器人填入
	if !configService.IsRobot(userId) && core.AppConfig.EnableJoinRandomRoom != 1 {
		roomId = 0
	}

	// 这里加一个判断，机器人不能自己创建房间
	if roomId == 0 && configService.IsRobot(userId) {
		return core.NewError(-322)
	}

	// 未找到可以加入的房间，创建一个房间
	var room *Room
	if roomId == 0 {
		room = Create(userId, gType, config.ROOM_RANDOM_ROUND, config.ROOM_TYPE_RAND, setting, config.ROOM_CREATE_MODE_USER)
		room.PayPrice = price

		// 保存房间号与房间id的对应关系，这里因为并发的原因，可能会失败
		if !roomService.SaveRoom(room.RoomId, room.Number, GetRemoteAddr()) {
			return core.NewError(-301, room.Number)
		}
		// 监听房间超时
		go listenRoomTimeout(room.RoomId)
		// 监听房间准备超时
		go listenRoomKickUser(room.RoomId)
		// 写入队列
		queue.Rooms[room.RoomId] = room.Number
		// 记录大厅房间列表
		hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)
		// 将这个随机房间的id，记入robot room list
		if room.EnableRobot() {
			robotGameInfo := robot.NewGameInfo(GetRemoteAddr(), room.RoomId, room.MType, room.CType, room.setting.GetSettingPlayerCnt(), room.TRound)
			hallService.AddHallRobotRoom(GetRemoteAddr(), robotGameInfo.String())
		}
	} else {
		room, _ = RoomMap.GetRoom(roomId)
	}

	// 这里需要加锁，防止并发退出
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 如果是机器人，这里重写一下机器人的随机积分，防止出错
	if configService.IsRobot(userId) {
		user.Info.ScoreRandom = util.RandIntn(500)
		core.Logger.Debug("[randomRoom]设置机器人积分,userId:%v, score:%v", userId, user.Info.ScoreRandom)
	}

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 因为并发问题，只有房间是新创建的，才将房间数据添加到RoomMap
	if roomId == 0 {
		RoomMap.SetRoom(room)
	}

	// 给用户回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(room, mNumber))

	// 给其他成员发送有人加入的push
	if room.GetUsersLen() > 1 {
		roomUserInfo := room.GetUser(userId)
		responsePacket := JoinRoomPush(roomUserInfo)
		room.SendMessageToRoomUser(responsePacket, userId)

		// 发送push
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)
	}

	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if room.IsFull() {
		// 从队列移除
		delete(queue.Rooms, room.RoomId)

		// 房间开始
		room.enter()
	}

	core.Logger.Info("[RandomJoin]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 创建房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func CreateRoom(userId int, gType int, round int, setting []int, clubId int, mNumber uint16) *core.Error {
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	// 判断是否是正确的麻将类型
	if _, isExists := fbsCommon.EnumNamesGameType[gType]; !isExists {
		return core.NewError(-500, gType)
	}

	// 判断是否合理的局数
	_, roundExists := config.MahjongRoundPrice[round]
	if !roundExists {
		return core.NewError(-501, round)
	}

	if !Cre {
		return nil
	}

	// 创建房间的类型
	var cType = config.ROOM_TYPE_CREATE
	var club *config.Club
	var clubUser *config.ClubUser
	if round == 0 {
		// 无局数限制的，就认为是TV房间
		cType = config.ROOM_TYPE_TV
	} else if clubId > 0 {
		cType = config.ROOM_TYPE_CLUB
		// 判断俱乐部是否开启淘汰赛
		club = clubModel.GetClub(clubId)
		clubUser = clubModel.GetClubUser(clubId, userId)
		if club.EnableOut == 1 {
			cType = config.ROOM_TYPE_CLUB_MATCH
			// 判断用户的俱乐部积分是否足够
			if clubUser.Score < club.Score {
				return core.NewError(-335, userId, clubId, club.Score, clubUser.Score)
			}
			// 更新用户信息中的俱乐部淘汰赛积分
			user.Info.ScoreClub = clubUser.Score
		}
		// 是否设置了会员不允许创建房间
		if club.AllowCreateroom == 0 && clubUser.Type != 2 {
			return core.NewError(-336)
		}
	}

	// 电视端需要限制ip登录
	if configService.IsTVRoom(cType) && !IsOBIP(user.Info.Ip) {
		return core.NewError(-331, user.UserId)
	}

	// 判断金额是否足够
	price := configService.GetGamePrice(gType, config.ROOM_TYPE_CREATE, round)
	// 付费方式，默认是用户
	payType := config.ROOM_PAY_TYPE_USER
	if price > 0 {
		if configService.IsClubRoom(cType) || configService.IsClubMatchRoom(cType) {
			// 判断用户是否在俱乐部中
			if club.Id == 0 || clubUser.ClubId == 0 {
				return core.NewError(-333, userId, clubId)
			}
			// 如果俱乐部基金足够，使用俱乐部基金付房费
			// 如果俱乐部基金不足，则使用成员自己的钻石付房费
			// 俱乐部基金与用户钻石都不足时，不能创建房间
			if club.Fund >= price {
				payType = config.ROOM_PAY_TYPE_CLUB
			} else {
				if !userService.CheckMoneyEnough(userId, price, nil) {
					return core.NewError(-204, price)
				}
			}
		} else {
			if !userService.CheckMoneyEnough(userId, price, nil) {
				return core.NewError(-204, price)
			}
		}
	}

	// 创建房间
	room := Create(userId, gType, round, cType, setting, config.ROOM_CREATE_MODE_USER)
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
	if room.IsClub() {
		room.ClubId = clubId
		room.ClubCapitorUserId = club.ManageUser
	}
	if room.IsClubMatch() {
		room.ClubScore = club.Score
	}
	go listenRoomTimeout(room.RoomId)
	hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 加入房间列表
	RoomMap.SetRoom(room)

	// 创建成功
	user.AppendMessage(JoinRoomResponse(room, mNumber))

	// 如果是俱乐部房间
	if room.IsClub() {
		// 推送一个新建俱乐部房间的消息到俱乐部服务
		CPool.appendMessage(ClubG2CReloadRoomPush(room.ClubId, room))
		// 写入消息推送队列
		roomService.AppendClubCreateRoomPush(room.ClubId, room.RoomId, room.Number, room.Creator, room.setting.GetSetting())
	}

	core.Logger.Info("[createRoom]userId:%d, roomId:%d, number:%s, clubId:%v, remote:%v", userId, room.RoomId, room.Number, clubId, GetRemoteAddr())

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 加入房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func JoinRoom(userId int, number string, mNumber uint16) *core.Error {
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	// 读取房间编号对应的房间id
	roomId := roomService.GetRoomIdByNumber(number)
	if roomId == 0 {
		return core.NewError(-302, number)
	}

	// 判断房间是否已存在
	room, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}

	// 检查房间是否允许加入
	if !room.EnableJoin() {
		return core.NewError(-320)
	}

	// 电视端需要判断加入用户的ip
	if room.IsTV() && !IsOBIP(user.Info.Ip) {
		return core.NewError(-331, user.UserId)
	}

	// 检查客户端潘奔
	if !room.checkJoinVersion(user.Info.Version) {
		core.Logger.Debug("[JoinRoom]用户版本号过低,userId:%v, roomId:%v, version:%v", user.UserId, room.RoomId, user.Info.Version)
		return core.NewError(-321)
	}

	var clubUser *config.ClubUser
	if room.IsClub() {
		clubUser = clubModel.GetClubUser(room.ClubId, userId)
		if clubUser.ClubId == 0 {
			return core.NewError(-334, userId, room.ClubId)
		}
		// 判断用户俱乐部积分是否足够
		if room.IsClubMatch() {
			if clubUser.Score < room.ClubScore {
				return core.NewError(-335, userId, room.ClubId, room.ClubScore, clubUser.Score)
			}
			// 更新用户信息中的俱乐部淘汰赛积分
			user.Info.ScoreClub = clubUser.Score
		}
	}

	// 这里需要加锁，防止并发加入
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 判断房间是否已结束
	if room.CheckStatus(config.ROOM_STATUS_COMPLETED) {
		return core.NewError(-302, number)
	}

	// 判断房间人数是否已满
	if room.IsFull() {
		return core.NewError(-303, number)
	}

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 更新房间列表
	// RoomMap[room.RoomId] = room
	RoomMap.SetRoom(room)

	// 回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(room, mNumber))

	// 给其他成员发送有人加入的push
	if room.GetUsersLen() > 1 {
		roomUserInfo := room.GetUser(userId)
		pushPacket := JoinRoomPush(roomUserInfo)
		room.SendMessageToRoomUser(pushPacket, userId)

		// 发送push
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)

		// 发送消息给观察员
		room.Ob.sendMessage(pushPacket, 0)
	}

	// 如果是俱乐部房间，推送一个新建俱乐部房间的消息到俱乐部服务
	if room.IsClub() {
		CPool.appendMessage(ClubG2CJoinRoomPush(room.ClubId, room.RoomId, room.GetUser(userId)))
	}

	core.Logger.Info("[joinRoom]userId:%d,roomId:%d,number:%s", userId, room.RoomId, room.Number)

	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if room.IsFull() {
		// 房间开始
		room.enter()
	}

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* CoinJoin加入金币场
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func CoinJoin(userId int, impacket *protocal.ImPacket) *core.Error {
	request := fbsCommon.GetRootAsCoinRoomRequest(impacket.GetBody(), 0)
	coinType := int(request.CoinType())
	gType := int(request.GameType())
	lastRoomId := int64(request.LastRoomId())
	core.Logger.Debug("[CoinJoin]userId:%v, coinType:%v, gType:%v, lastRoomId:%v", userId, coinType, gType, lastRoomId)

	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	if !Coe {
		return nil
	}

	// 读取游戏类型配置
	var coinConfig *coinService.Config
	if coinType == 0 {
		// 新版本，自动选择金币类型
		for coinType = 6; coinType >= 1; coinType-- {
			coinConfig = coinService.GetConfig(coinType, gType)

			if coinConfig != nil && coinConfig.Status == 0 &&
				user.Info.ScoreCoin >= coinConfig.RequireLowCoin {
				break
			}
			coinConfig = nil
		}
		if coinConfig == nil || coinConfig.Status == 1 {
			return core.NewError(-700, userId, 0, gType)
		}
	} else {
		// 继续支持老版本
		coinConfig = coinService.GetConfig(coinType, gType)
		if coinConfig == nil || coinConfig.Status == 1 {
			return core.NewError(-700, userId, coinType, gType)
		}

		// 判断积分是否匹配
		if !configService.IsRobot(userId) {
			if coinConfig.RequireLowCoin > 0 && coinConfig.RequireLowCoin > user.Info.ScoreCoin {
				return core.NewError(-701, userId, coinConfig.RequireLowCoin, user.Info.ScoreCoin)
			}
			if coinConfig.RequireHighCoin > 0 && user.Info.ScoreCoin > coinConfig.RequireHighCoin {
				return core.NewError(-702, userId, coinConfig.RequireHighCoin, user.Info.ScoreCoin)
			}
		}
	}

	// 读取金币场的设置
	s := configService.GetCoinRoomDefaultSetting(gType)
	if s == nil {
		return core.NewError(-703, userId, gType)
	}

	// 读取游戏队列
	queue := CoinRoomQueueMap.Get(coinType, gType)
	// 找出是否有可加入的队列
	var roomId int64
	queue.Rooms.Range(func(k, v interface{}) bool {
		room, err := RoomMap.GetRoom(k.(int64))
		if err != nil {
			core.Logger.Warn("[CoinJoin]roomId存在于队列中，但是在房间列表未找到，roomId: %v", k)
			queue.Rooms.Delete(k.(int64))
			return true
		}
		if room.IsFull() || !room.CheckStatus(config.ROOM_STATUS_CREATING) {
			core.Logger.Warn("[CoinJoin]房间已满或者已开始,roomId:%v", k)
			queue.Rooms.Delete(k.(int64))
			return true
		}
		if lastRoomId > 0 {
			// 如果选择了不换桌，只找上次一起玩的人放在一起
			if lastRoomId != room.RelationRoomId {
				return true
			}
		} else {
			// 换桌的情况下，需要做防刷限制，已经一起玩过的玩家，不能继续匹配,
			if room.HasTogetherGameLog(userId) {
				return true
			}

			// 如果房间属于“再来一把”创建的，在规定的“锁定”秒内，不允许非上把同一桌的玩家加入
			if room.RelationRoomId > 0 && util.GetTime()-room.CreateTime < config.ROOM_COIN_ALLOW_OTHER_USER_INTERVAL {
				return true
			}
		}
		roomId = room.RoomId
		return false
	})
	core.Logger.Debug("[CoinJoin]find coin room, userId:%v, roomId:%v, lastRoomId:%v,", userId, roomId, lastRoomId)
	// 这里加一个判断，机器人不能自己创建房间
	if roomId == 0 && configService.IsRobot(userId) {
		return core.NewError(-322)
	}
	// 未找到可以加入的房间，创建一个房间
	var room *Room
	if roomId == 0 {
		room = Create(userId, gType, config.ROOM_COIN_ROUND, config.ROOM_TYPE_COIN, s, config.ROOM_CREATE_MODE_USER)
		room.CoinType = coinType
		room.PayPrice = coinConfig.ConsumeCoin
		room.setting.Multiple = coinConfig.BaseCoin
		room.RelationRoomId = lastRoomId
		room.CoinConfig = coinConfig

		// 保存房间号与房间id的对应关系，这里因为并发的原因，可能会失败
		if !roomService.SaveRoom(room.RoomId, room.Number, GetRemoteAddr()) {
			return core.NewError(-301, room.Number)
		}
		// 监听房间超时
		go listenRoomTimeout(room.RoomId)
		// 监听房间准备超时
		// go listenRoomKickUser(room.RoomId)
		// 写入队列
		queue.Rooms.Store(room.RoomId, room.Number)
		// 记录大厅房间列表
		hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)
		// 将这个随机房间的id，记入robot room list
		if room.EnableRobot() {
			robotGameInfo := robot.NewGameInfo(GetRemoteAddr(), room.RoomId, room.MType, room.CType, room.setting.GetSettingPlayerCnt(), room.TRound)
			robotGameInfo.CoinType = coinType
			hallService.AddHallRobotRoom(GetRemoteAddr(), robotGameInfo.String())
		}
	} else {
		room, _ = RoomMap.GetRoom(roomId)
	}

	// 这里需要加锁，防止并发退出
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 如果是机器人，这里重写一下机器人的随机积分，防止出错
	if configService.IsRobot(userId) {
		minCoin := 0
		maxCoin := 0

		// 计算真是玩家的平均数
		realUserCnt := 0
		realUserCoinTotal := 0
		average := 0
		room.Users.Range(func(k, v interface{}) bool {
			ru := v.(*RoomUser)
			if !configService.IsRobot(ru.UserId) {
				realUserCnt++
				realUserCoinTotal += ru.Info.ScoreCoin
			}
			return true
		})
		if realUserCnt > 0 {
			average = realUserCoinTotal / realUserCnt
			minCoin = average / 2
			maxCoin = (average * 3) / 2
		}
		if minCoin == 0 || minCoin < coinConfig.RequireLowCoin {
			minCoin = coinConfig.RequireLowCoin
		}
		if maxCoin == 0 || maxCoin > coinConfig.RequireHighCoin {
			maxCoin = coinConfig.RequireHighCoin
		}
		if maxCoin > config.COIN_UPPER_LIMIT {
			maxCoin = config.COIN_UPPER_LIMIT
		}
		if minCoin >= 10000000 {
			minCoin = 8000000
		}
		if maxCoin >= 15000000 {
			maxCoin = 15000000
		}
		// 取整，100的倍数
		user.Info.ScoreCoin = ((minCoin + util.RandIntn(maxCoin-minCoin)) / 5) * 5
		core.Logger.Debug("[coinRoom]设置机器人积分,userId:%v, score:%v, 真实用户个数:%v, 平均分:%v", userId, user.Info.ScoreCoin, realUserCnt, average)
	}

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 因为并发问题，只有房间是新创建的，才将房间数据添加到RoomMap
	if roomId == 0 {
		RoomMap.SetRoom(room)
	}

	// 给用户回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(room, impacket.GetMessageNumber()))

	// 给其他成员发送有人加入的push
	if room.GetUsersLen() > 1 {
		roomUserInfo := room.GetUser(userId)
		responsePacket := JoinRoomPush(roomUserInfo)
		room.SendMessageToRoomUser(responsePacket, userId)

		// 发送push
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)
	}
	// 增加金币�����实时在线人数
	coinService.IncrCoinUserCnt(gType, coinType)

	core.Logger.Info("[CoinJoin]userId:%d, coinType:%v, gType:%v, roomId:%d, number:%s", userId, coinType, gType, room.RoomId, room.Number)

	// 如果房间人满，则直接开始游戏，并且从队列中移除房���
	if room.IsFull() {
		// 从队列移除
		queue.Rooms.Delete(room.RoomId)
		// 房间开始
		room.enter()
	}

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 退出房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func QuitRoom(userId int) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		// 异常数据修复
		core.Logger.Warn("[QuitRoom]修复用户cache中的roomId, userId:%d, roomId: %d.", userId, user.RoomId)

		if user.RoomId > 0 {
			user.RoomId = int64(0)
			userService.DelRoomId(userId, room.RoomId)
		}

		// 异常的时候，需要回复一个正常的退出回应，让客户端退出游戏
		user.AppendMessage(QuitRoomPush(userId, 0, config.QUIT_ROOM_CODE_INITIATIVE))

		return nil
	}

	// 未满员的才可以退出
	if room.Round > 0 {
		return core.NewError(-304, userId, room.RoomId)
	}

	// 这里需要加锁，防止并发退出
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 将用户从房间移除
	userIndex := room.GetUser(userId).Index
	isDismiss, isSuccess := RemoveRoomUser(room, userId, config.DISMISS_ROOM_CODE_HOST_LEAVE, config.QUIT_ROOM_CODE_INITIATIVE)
	// 用户是否已不在房间中
	if !isSuccess {
		return nil
	}
	// 删除用户内存中的房间
	user.RoomId = int64(0)

	// 如果是俱乐部房间，推送一个新建俱乐部房间的消息到俱乐部服务
	if room.IsClub() && !isDismiss {
		CPool.appendMessage(ClubG2CQuitRoomPush(room.ClubId, room.RoomId, userId, userIndex))
	}

	core.Logger.Info("[QuitRoom]userId:%d, roomId:%d, number:%s, dismiss:%t", userId, room.RoomId, room.Number, isDismiss)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 申请解散房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func DismissApply(userId int) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 判断是否正在解散中
	if room.Dissmisser > 0 {
		// 如果房间正在解散中，则将用户的操作改为同意回应
		return DismissReply(userId, config.ROOM_DISMISS_ALLOW)
	}

	// 防并发加锁
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 判断房间是否正在创建中
	if room.CheckStatus(config.ROOM_STATUS_CREATING) {
		return core.NewError(-306, user.RoomId)
	}

	// 这里做一下容错，在提出申请时候，操作列表必须是空的
	room.DismissOp = &sync.Map{}

	// 更新房间状态，解散发起者默认已同意
	room.Dissmisser = userId
	room.DismissTime = util.GetTime()
	room.DismissOp.Store(userId, config.ROOM_DISMISS_ALLOW)
	// 开启通信channel
	room.DissmissChan = make(chan int)

	// 开启协程，监听房间状态，60秒之后，如果无人操作，则将房间解散
	go listenRoomDissmiss(room)

	// 发送push，通知需要回应解散房间
	responsePacket := DismissRoomPush(userId, config.ROOM_DISMISS_APPLY)
	room.SendMessageToRoomUser(responsePacket, userId)

	// 给离线用户推送APNs消息
	room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_DISMISS_APPLY, userId)

	core.Logger.Info("[DismissApply]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 解散操作回应
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func DismissReply(userId int, op int) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 防并发加锁
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 判断是否正在解散中
	if room.Dissmisser == 0 {
		return core.NewError(-309, user.RoomId)
	}

	// 判断用户是否已操作过
	if _, ok := room.DismissOp.Load(userId); ok {
		return core.NewError(-310, user.RoomId, userId)
	}

	// 发送push，通知回应操作
	responsePacket := DismissRoomPush(userId, op)
	room.SendMessageToRoomUser(responsePacket, 0)

	if op == config.ROOM_DISMISS_DENY {
		// 拒绝解散
		room.DissmissChan <- config.ROOM_DISMISS_DENY

		// 推送消息，拒绝解散房间
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_DISMISS_DENY, userId)
	} else {
		room.DismissOp.Store(userId, op)

		// 所有人都做出回应后，执行解散操作
		if util.SMapLen(room.DismissOp) == room.GetUsersLen() {
			room.DissmissChan <- config.ROOM_DISMISS_ALLOW
		}

		// 推送消息，同意解散房间
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_DISMISS_AGREE, userId)
	}

	core.Logger.Info("[DismissReply]userId:%d, roomId:%d, number:%s, agree:%d", userId, room.RoomId, room.Number, op)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 客户端准备
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func GameReady(userId int, impacket *protocal.ImPacket) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 防并发加锁
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 判断房间是否处于回应准备状态
	allReadyed, err := room.userOperationReady(userId)
	if err != nil {
		core.Logger.Error("[GameReady]roomId:%d,userId:%d,error:%s", room.RoomId, userId, err.Error())
		return nil
	}

	// 更新用户的经纬度
	request := fbsCommon.GetRootAsGameReadyNotify(impacket.GetBody(), 0)
	ru := room.GetUser(userId)
	ru.Longitude = request.Lng()
	ru.Latitude = request.Lat()
	core.Logger.Info("[GameReady]roomId:%d, userId:%d, lng:%v, lat:%v", user.RoomId, userId, ru.Longitude, ru.Latitude)

	if allReadyed {
		// 准备完成，执行初始化
		room.nextGame()
	}
	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 请求重连
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func GameRestore(userId int, impacket *protocal.ImPacket) *core.Error {
	// 解析参数
	request := fbsCommon.GetRootAsGameRestoreNotify(impacket.GetBody(), 0)
	roomId := int64(request.RoomId())
	round := int(request.Round())
	seq := int(request.Step())

	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 这里需要判断是否允许进行片断重连
	core.Logger.Info("[GameRestoreNotify]userId:%d, roomId:%v", userId, user.RoomId)

	if room.canRestoreSection(user, roomId, round, seq) {
		room.restoreSection(user, seq)
	} else {
		room.restoreIntact(user)
	}

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 客户端重连
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func GameRestoreDone(userId int) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	user.Mux.Lock()
	defer user.Mux.Unlock()

	if !user.MQ.WasStarted() {
		// 推送用户上线的消息
		// 观察员上线，不发消息
		if !room.Ob.hasUser(userId) {
			onlinePush := UserOnlinePush(userId, 1)
			room.SendMessageToRoomUser(onlinePush, userId)
			room.Ob.sendMessage(onlinePush, 0)
		}

		// 开启消息推送
		go user.MQ.Start()
	} else if user.MQ.WasPaused() {
		// 暂停结束
		user.MQ.Continue()
	} else {
		core.Logger.Warning("[GameRestoreDone]mq.status error, roomId:%v, userId:%v, status:%v", room.RoomId, userId, user.MQ.GetStatus())
	}

	// 记录重连日志
	restoreLog(userId, room.RoomId, room.Number, room.Round, config.RESTORE_LOG_TYPE_RECONNECT)

	core.Logger.Info("[GameRestoreDone]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)
	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* ��取用户距离
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func GetDistanceList(userId int, impacket *protocal.ImPacket) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// if len(room.UserDistances) > 0 {
	packet := GameUserDistanceResponse(room.UserDistances, impacket.GetMessageNumber())
	user.AppendMessage(packet)
	// }
	core.Logger.Info("[GetDistance]userId:%v, roomId:%v, number:%v", userId, room.RoomId, room.Number)
	return nil
}
