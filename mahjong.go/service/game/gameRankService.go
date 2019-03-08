package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/model"
	"mahjong.go/rank"
	"mahjong.go/robot"
	configService "mahjong.go/service/config"
	friendService "mahjong.go/service/friend"
	hallService "mahjong.go/service/hall"
	rankService "mahjong.go/service/rank"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* RankRoomRequest 加入排位赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func RankRoomRequest(userId int, impacket *protocal.ImPacket) *core.Error {
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if roomId := LoadRoomId(user.UserId); roomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(roomId))
	}

	if !Rne {
		return nil
	}

	// 判断当前赛季是否开启
	season := model.GetSeason()
	// season = &model.Season{
	// 	Id:      2,
	// 	GType:   5,
	// 	Rounds:  2,
	// 	IsFree:  0,
	// 	Status:  1,
	// 	Setting: []int{1, 1, 0, 0, 0, 1, 0, 0, 0, 2, 108, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	// }
	if season == nil || !season.IsOpen() {
		return core.NewError(-10600, userId)
	}

	// 读取用户的赛季数据
	seasonUser := rankService.GetSeasonUser(userId, season.Id)
	core.Logger.Debug("seasonUser:%#v", seasonUser)
	if seasonUser == nil || seasonUser.UserId == 0 {
		return core.NewError(-10601, userId)
	}

	// 判断用户的游戏次数
	// 如果用户达到雀王级别，并且排名在前三，则限制每日游戏次数
	core.Logger.Debug("[RankRoom]userId:%v, seasonId:%v, gradeId:%v", userId, seasonUser.SeasonId, seasonUser.GradeId)
	if seasonUser.GradeId >= 5 {
		// userRank := rankService.GetProvinceRank(season.Id, userId)
		// if userRank < 3 {
		playedTimes := rankService.GetDayPlayTimes(userId)
		core.Logger.Debug("[RankRoom]userId:%v, seasonId:%v, gradeId:%v, playedTimes", userId, seasonUser.SeasonId, seasonUser.GradeId, playedTimes)
		if playedTimes >= 20 {
			return core.NewError(-10603)
		}
		// }
	}

	// 这里记录下用户的 赛季数据到用户信息
	user.Info.ScoreRank = rank.FormatSLevel(seasonUser.GradeId, seasonUser.GradeLevel, seasonUser.StarNum)
	// 记录用户所在的排位城市
	user.Info.RankCity = seasonUser.LastCity
	// 用户的排位赛经验
	user.Info.RankExp = seasonUser.Exp
	core.Logger.Debug("user.Info.ScoreRank:%#v, exp:%v, ", user.Info.ScoreRank, user.Info.RankExp)

	// 判断用户的参赛卡是否足够
	var consume int
	if season.IsFree == 0 {
		consume = rankService.GetConsumeCard(season.Id, seasonUser.GradeId)
		// 记录下用户的消耗，在实际扣费时，直接读次字段，无需重新计算
		user.Info.RankConsume = consume
	}
	if consume > 0 {
		userInfoList := userService.GetUserInfoList(userId)
		if userService.GetRankCards(userInfoList) < consume {
			// 参赛卡不足
			return core.NewError(-10602, userId, consume, userService.GetRankCards(userInfoList))
		}
	} else {
		core.Logger.Warn("[RankRoomRequest]consume is 0, seasonId:%v, gradeId:%v, userId:%v", season.Id, seasonUser.GradeId, userId)
	}

	// 防并发，全局加锁，如果并发量过大，可以改成queue锁
	RankLock.Lock()
	defer RankLock.Unlock()

	// 匹配房间
	room, queue := findRoom(season, seasonUser)
	if room == nil {
		// 这里加一个判断，机器人不能自己创建房间
		if configService.IsRobot(userId) {
			return core.NewError(-322)
		}
		// 未匹配到房间，则自动创建一个房间
		room = Create(userId, season.GType, season.Rounds, config.ROOM_TYPE_RANK, season.Setting, config.ROOM_CREATE_MODE_USER)
		room.PayPrice = consume
		room.SeasonId = season.Id
		room.GradeId = seasonUser.GradeId

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
			robotGameInfo.GradeId = room.GradeId
			robotGameInfo.AILevel = getRankRobotAILevel(room.GradeId)
			hallService.AddHallRobotRoom(GetRemoteAddr(), robotGameInfo.String())
		}
		RoomMap.SetRoom(room)
	}

	// 将用户加入房间
	user.RoomId = room.RoomId
	room.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

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

	core.Logger.Info("[RankRoom]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)

	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if room.IsFull() {
		// 从队列移除
		queue.Rooms.Delete(room.RoomId)
		// 房间开始
		room.enter()
	}

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 机器人加入排位赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func RankRoomRobotRequest(userId int, impacket *protocal.ImPacket) *core.Error {
	request := fbsCommon.GetRootAsRankRoomRobotRequest(impacket.GetBody(), 0)
	roomId := int64(request.RoomId())
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	r, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}

	// 防并发，全局加锁，如果并发量过大，可以改成queue锁
	RankLock.Lock()
	defer RankLock.Unlock()

	// 房间已满
	if r.IsFull() {
		return core.NewError(-303, r.Number)
	}

	// 如果用户是机器人，则重新生成一下机器人的星
	robotGradeId := r.GradeId
	// 机器人总要比用户低一个等级
	if robotGradeId > 1 {
		robotGradeId -= 1
	}
	robotGradeLevel := util.RandIntn(rank.GradeList[robotGradeId].GradeLevelMax) + 1
	robotStar := util.RandIntn(rank.GradeList[robotGradeId].StarMax) + 1
	user.Info.ScoreRank = rank.FormatSLevel(robotGradeId, robotGradeLevel, robotStar)
	rankSInfo := rank.RankStarList[user.Info.ScoreRank]
	// TODO
	if rankSInfo == nil {
		core.Logger.Error("[RankRoomRobotRequest]获取机器人段位失败, robotGradeId:%v, robotGradeLevel:%v, robotStar:%v, ScoreRank:%v",
			robotGradeId, robotGradeLevel, robotStar, user.Info.ScoreRank)
	}
	if robotStar == 0 {
		user.Info.RankExp = rankSInfo.MaxExp
	} else {
		user.Info.RankExp = util.RandIntn(rankSInfo.MaxExp-rankSInfo.MinExp+1) + rankSInfo.MinExp
	}
	core.Logger.Debug("[RankRoom]计算机器人排位赛等级, roomId:%v, gradeId:%v, gradeLevel:%v, star:%v, scoreRank:%v, rankExp:%v",
		r.RoomId, robotGradeId, robotGradeLevel, robotStar, user.Info.ScoreRank, user.Info.RankExp)

	// 写入机器人的排位赛等级
	rankService.SetSeasonUserFromCache(userId, robotGradeId, robotGradeLevel, robotStar, 0, user.Info.RankExp)

	// 将用户加入房间
	user.RoomId = r.RoomId
	r.AddUser(user)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, r.RoomId)

	// 给用户回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(r, impacket.GetMessageNumber()))

	// 给其他成员发送有人加入的push
	if r.GetUsersLen() > 1 {
		roomUserInfo := r.GetUser(userId)
		responsePacket := JoinRoomPush(roomUserInfo)
		r.SendMessageToRoomUser(responsePacket, userId)

		// 发送push
		r.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_JOIN, userId)
	}
	// 如果房间人满，则直接开始游戏，并且从队列中移除房间
	if r.IsFull() {
		// 从队列移除
		q := RankRoomQueueMap.Get(r.GradeId)
		q.Rooms.Delete(r.RoomId)
		// 房间开始
		r.enter()
	}
	core.Logger.Info("[RankRoomRobot]userId:%d, roomId:%d, number:%s, gradeId:%v", userId, r.RoomId, r.Number, r.GradeId)

	return nil
}

// 匹配队列
func findRoom(season *model.Season, seasonUser *model.SeasonUser) (*Room, *RankRoomQueue) {
	// 匹配
	// 当前段位可以加入
	canJoinGrades := []int{seasonUser.GradeId}
	// 如果有前一个段位，则也加入匹配对类
	if seasonUser.GradeId > 1 {
		canJoinGrades = append(canJoinGrades, seasonUser.GradeId-1)
	}

	core.Logger.Debug("[findRankRoom]gradeId:%v, gradeLevel:%v, canJoinGrades:%v", seasonUser.GradeId, seasonUser.GradeLevel, canJoinGrades)

	var queue *RankRoomQueue
	var room *Room
	for _, gradeId := range canJoinGrades {
		core.Logger.Debug("[findRankRoom]from queue:%v", gradeId)
		q := RankRoomQueueMap.Get(gradeId)
		q.Rooms.Range(func(k, v interface{}) bool {
			r, err := RoomMap.GetRoom(k.(int64))
			if err != nil {
				core.Logger.Warn("[findRankRoom]roomId存在于队列中，但是在房间列表未找到，roomId: %v", k)
				q.Rooms.Delete(k.(int64))
				return true
			}
			if r.IsFull() || !r.CheckStatus(config.ROOM_STATUS_CREATING) {
				core.Logger.Warn("[findRankRoom]房间已满或者已开始,roomId:%v", k)
				q.Rooms.Delete(k.(int64))
				return true
			}
			// 检测有没有限制组队的人
			for _, otherId := range r.GetIndexUserIds() {
				togetherTimes := rankService.GetTogetherGameTimes(seasonUser.UserId, otherId)
				core.Logger.Debug("[together times]userId:%v, otherId:%v, times:%v", seasonUser.UserId, otherId, togetherTimes)
				if togetherTimes > 0 {
					return true
				}
			}
			core.Logger.Debug("[find rank room]gradeId:%v, roomId:%v, number:%v", gradeId, r.RoomId, r.Number)
			room = r
			queue = q
			return false
		})
		if room != nil {
			break
		}
	}
	// 如果没找到房间，则用户自己创建房间， 需要重新获取一次队列
	queue = RankRoomQueueMap.Get(canJoinGrades[0])

	return room, queue
}

// 观察用户的段位等级发生改变
// 如果用户在好友中的排名有提升，提要通知用户
func obUserSLevelModify(seasonId, userId, city, fromExp, toExp, fromLevel, toLevel int) {
	defer util.RecoverPanic()
	// 计算rankScore
	score := rank.FormatRankScore(toLevel, userId)
	var rankUp = rank.SLevelRev(toLevel) > rank.SLevelRev(fromLevel)
	// 更新前省内排行, 更新后省内排行
	var fromFriendRank, toFriendRank int
	// 超越了谁
	var beyondFriend int
	// 读取更新前在好友中的排行
	if rankUp {
		fromFriendRank = rankService.GetUserFriendRank(seasonId, userId)
	}
	// 更新好友排行
	rankService.UpdateFriendRank(seasonId, userId, score)
	// 读取更新后在好友中的排行
	// 读取更新后超越的好友
	if rankUp {
		toFriendRank = rankService.GetUserFriendRank(seasonId, userId)
	}
	// 通知用户排名发生变化
	if toFriendRank < fromFriendRank {
		beyondFriend = rankService.GetFriendRankUser(seasonId, userId, toFriendRank+1)
		rankService.SaveRankUpgrade(userId, beyondFriend, toFriendRank)
	}
	core.Logger.Info("[obUserSLevelModify]seasonId:%v, userId:%v, city:%v, fromLevel:%v, toLevel:%v, fromFriendRank:%v, toFriendRank:%v, beyondFriend:%v.",
		seasonId, userId, city, fromLevel, toLevel, fromFriendRank, toFriendRank, beyondFriend)

	// 更新省排行
	rankService.UpdateProvinceRank(seasonId, userId, score)
	// 更新城市排行
	rankService.UpdateCityRank(seasonId, userId, city, score)
	// 最后更新我在好友中的排行
	for _, friendUserId := range friendService.GetFriends(userId) {
		rankService.UpdateUserFriendRank(seasonId, userId, friendUserId, score)
	}
	// 更新段位人数
	fromGradeId, _, _ := rank.ExplainSLevel(fromLevel)
	toGradeId, _, _ := rank.ExplainSLevel(toLevel)
	if fromGradeId != toGradeId {
		rankService.UpdateGradeSignCnt(fromGradeId, -1)
		rankService.UpdateGradeSignCnt(toGradeId, 1)
	}

	// 如果省排名，进入前10，那么发送全服消息
	provinceRank := rankService.GetProvinceRank(seasonId, userId)
	rankService.SendProvinceRankMessage(userId, provinceRank+1)

	// 如果用户新升级到雀王或者雀神，发送全服消息
	if fromGradeId < toGradeId {
		rankService.SendGradeUpMessage(userId, toGradeId)
	}
}

func getRankRobotAILevel(gradeId int) int {
	var AILevel int
	switch gradeId {
	case 1:
		fallthrough
	case 2:
		AILevel = 1
	case 3:
		AILevel = 2
	case 4:
		AILevel = 3
	case 5:
		fallthrough
	case 6:
		AILevel = 4
	default:
	}
	return AILevel
}
