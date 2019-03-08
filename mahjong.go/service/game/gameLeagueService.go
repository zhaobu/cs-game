package game

import (
	"math/rand"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/model"
	configService "mahjong.go/service/config"
	hallService "mahjong.go/service/hall"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

// 收到来自联赛的排赛请求
func l2sPlanPush(impacket *protocal.ImPacket) {
	core.Logger.Debug("[l2sPlanPush]收到消息")
	raceRoomId, raceId, roomId, err := handleLeaguePlan(impacket)
	core.Logger.Debug("[l2sPlanPush]收到消息,raceRoomId:%v, raceId:%v, roomId:%v", raceRoomId, raceId, roomId)
	if err != nil {
		core.Logger.Error("[l2sPlanPush]排赛失败, raceRoomId:%v, raceId:%v, roomId:%v, err:%v", raceRoomId, raceId, roomId, err.Error())
	}
	LPool.appendMessage(LeagueS2LPlanPush(err, roomId, raceRoomId, raceId))
}

// 收到来自联赛的排名更新消息
func l2sRankRefresh(impacket *protocal.ImPacket) {
	push := fbsCommon.GetRootAsLeagueL2SRankRefreshPush(impacket.GetBody(), 0)
	raceId := push.RaceId()
	playRoomCount := int(push.RoomCnt())
	core.Logger.Debug("[l2sRankRefresh]收到比赛排名更新通知, raceId:%v", raceId)

	// 获取race对应的房间列表
	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms == nil {
		core.Logger.Warn("[l2sRankRefresh]raceRooms = nil, raceId:%v", raceId)
		return
	}

	// 获取所有的房间用户
	for _, userId := range raceRooms.GetRoomUserIds() {
		// 读取用户排名, 推送用户排名发生变化, 不在线的用户不发
		if u, _ := UserMap.GetUser(userId); u != nil {
			rank := model.GetRaceUserRank(raceId, userId)
			u.AppendMessage(LeagueGameRankPush(rank, 0, playRoomCount))
			core.Logger.Debug("[l2sRankRefresh]通知用户排名发生变化,raceId:%v, userId:%v, rank:%v", raceId, userId, rank)
		}
	}
	/*
		raceRooms.Mux.Lock()
		for roomId, users := range raceRooms.Data {
			for _, userId := range users {
				// 读取用户排名, 推送用户排名发生变化, 不在线的用户不发
				if u, _ := UserMap.GetUser(userId); u != nil {
					rank := model.GetRaceUserRank(raceId, userId)
					u.AppendMessage(LeagueGameRankPush(rank, 0, playRoomCount))
					core.Logger.Debug("[l2sRankRefresh]通知用户排名发生变化,raceId:%v, roomId:%v, userId:%v, rank:%v", raceId, roomId, userId, rank)
				}
			}
		}
		raceRooms.Mux.Unlock()
	*/
}

// handleLeaguePlan 排赛
// return raceRoomId raceId roomId
func handleLeaguePlan(impacket *protocal.ImPacket) (int64, int64, int64, *core.Error) {
	request := fbsCommon.GetRootAsLeagueL2SPlanPush(impacket.GetBody(), 0)
	raceRoomId := request.RaceRoomId()
	core.Logger.Debug("[handleLeaguePlan]收到消息,raceRoomId:%v", raceRoomId)

	// 读取比赛房间信息
	var raceRoom *model.RaceRoom
	if raceRoomId > 0 {
		raceRoom = model.GetRaceRoom(raceRoomId)
	}
	if raceRoom == nil {
		return raceRoomId, 0, 0, core.NewError(-10500, raceRoomId)
	}
	// 通知排赛成功
	if raceRoom.RoomId > 0 {
		core.Logger.Warn("[LeaguePlan]比赛已排赛，无需重新排赛, raceRoomId:%v, raceId:%v, roomId:%v",
			raceRoomId, raceRoom.RaceId, raceRoom.RoomId)
		return raceRoomId, raceRoom.RaceId, raceRoom.RoomId, nil
	}
	// 读取比赛信息
	raceInfo := model.GetRace(raceRoom.RaceId)
	// 读取比赛模板
	leagueInfo := model.GetLeague(raceInfo.LeagueId)
	// 是否最后一轮
	isLastRound := raceInfo.Round == len(leagueInfo.GetRounds())

	// 读取房间设置
	users := raceRoom.GetUsers()
	// 读取房间用户
	setting := leagueInfo.GetSetting()

	// 新建一个房间
	room := Create(users[0], leagueInfo.GameType, leagueInfo.GetGameRound(raceRoom.Round), config.ROOM_TYPE_LEAGUE, setting, config.ROOM_CREATE_MODE_SYSTEM)
	// 保存房间号与房间id的对应关系，这里因为并发的原因，可能会失败
	if !roomService.SaveRoom(room.RoomId, room.Number, GetRemoteAddr()) {
		return raceRoomId, raceRoom.RaceId, 0, core.NewError(-10501, raceRoom.Id, raceRoom.RaceId)
	}

	go listenRoomTimeout(room.RoomId)
	hallService.AddHallRoom(GetRemoteAddr(), room.RoomId)

	// 将用户加入房间
	for _, userId := range users {
		// 新建一个信息
		user := NewUser(userId, nil)
		info := &UserInfo{}
		// 读取用户信息
		userData := userService.GetUser(userId)
		// 获取用户扩展数据
		userInfoList := userService.GetUserInfoList(userId)

		// 如果是机器人，改变机器人的头像
		// 最后一轮还原原本的头像
		if configService.IsRobot(userId) && !isLastRound {
			offsetUserId := leagueRoomRobotOffset(room.RoomId, userId)
			offsetUserData := userService.GetUser(offsetUserId)
			info.Nickname = offsetUserData.Nickname
			info.Avatar = offsetUserData.IconUrl
		} else {
			info.Nickname = userData.Nickname
			info.Avatar = userData.IconUrl
		}
		info.Gender = userService.GetGender(userInfoList)
		info.Area = userService.GetCity(userInfoList)
		info.ScoreLeague = model.GetRaceUserScore(raceInfo.Id, userId)
		// 用户头像框
		info.AvatarBox = userService.GetUserAvatarBox(userId)
		// 用户会员等级
		info.MemberLevel, info.MemberAddExp = userService.GetUserMemberLevel(userId)
		user.Info = info
		room.AddUser(user)

		// 将用户房间数据存入cache
		userService.SetRoomId(userId, room.RoomId)
	}

	// 更新房间数据
	room.RaceInfo = raceInfo
	room.RaceRoom = raceRoom
	room.LeagueInfo = leagueInfo

	// 加入房间列表
	RoomMap.SetRoom(room)

	// 记录本服务器的比赛房间
	raceRooms := model.GetRaceRoomsNS(raceInfo.Id)
	raceRooms.Set(room.RoomId, users)

	// 更新联赛信息
	raceRoom.RoomId = room.RoomId
	raceRoom.ServerRemote = GetRemoteAddr()
	raceRoom.UpdateTime = util.GetTime()
	core.GetWriter().Update(raceRoom, "room_id", "update_time", "server_remote")
	core.Logger.Info("[LeaguePlan]leagueId:%v, raceId:%v, raceRoomId:%v, roomId:%v, users:%v", leagueInfo.Id, raceRoom.RaceId, raceRoom.Id, room.RoomId, users)

	go room.enter()

	return raceRoomId, raceRoom.RaceId, room.RoomId, nil
}

// 单局结算完成
func leagueGameRoundFinish(r *Room, scores []int) {
	for i := 0; i < len(scores); i += 2 {
		// 更新用户积分
		model.UpdateRaceUserScoreAndRank(r.RaceInfo.Id, scores[i], scores[i+1])
	}
	// 推送单局完成的消息
	LPool.appendMessage(LeagueS2LRoundFinishPush(r.RaceRoom.Id, r.RaceRoom.RaceId, scores))
	core.Logger.Info("[leagueGameRoundFinish]raceId:%v, raceRoomId:%v, roomId:%v, scores:%v", r.RaceInfo.Id, r.RaceRoom.Id, r.RoomId, scores)
}

// 房间结束
func leagueGameFinish(r *Room, code int) {
	// 更新房间状态
	if code == config.DISMISS_ROOM_CODE_FINISH {
		r.RaceRoom.Status = config.RACE_ROOM_STATUS_FINISH
	} else {
		r.RaceRoom.Status = config.RACE_ROOM_STATUS_DISMISS
	}
	r.RaceRoom.UpdateTime = util.GetTime()
	core.GetWriter().Update(r.RaceRoom, "status", "update_time")
	//从比赛对应的本服房间列表中删除
	if raceRooms := model.GetRaceRooms(r.RaceInfo.Id); raceRooms != nil {
		raceRooms.Del(r.RoomId)
		core.Logger.Debug("[leagueGameFinish]从本服比赛房间列表中删除, raceId:%v, roomId:%v", r.RaceInfo.Id, r.RoomId)
	}
	// 推送房间完成的消息
	LPool.appendMessage(LeagueS2LGameFinishPush(r.RaceRoom.Id, r.RaceRoom.RaceId, code))
	core.Logger.Info("[leagueGameFinish]raceId:%v, raceRoomId:%v, roomId:%v, code:%v", r.RaceInfo.Id, r.RaceRoom.Id, r.RoomId, code)
}

// 机器人头像偏移
func leagueRoomRobotOffset(roomId int64, userId int) int {
	r := rand.New(rand.NewSource(roomId))
	offset := r.Intn(100) + 1
	if userId >= 999800 {
		userId -= offset
	} else {
		userId += offset
	}
	return userId
}
