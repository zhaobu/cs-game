package service

import (
	"fmt"
	"io/ioutil"
	"mahjong-league/config"
	"mahjong-league/core"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/protocal"
	"mahjong-league/robot"
	"mahjong-league/user"
	"net/http"

	"github.com/fwhappy/util"
)

// PushAfterHandshake 在用户连接后，主动给用户推送数据
// 包括大厅列表、已参加的、正在进行的比赛
func PushAfterHandshake(u *user.User) {
	// 主动推送队列
	openList := model.LeagueList.GetOpenList()
	// 按版本过滤
	for k, league := range openList {
		if !league.CheckVersion(u.Version) {
			delete(openList, k)
		}
	}
	signuplist := model.RaceList.GetSignupList()
	u.SendMessage(LeagueListPush(openList, signuplist))
	core.Logger.Debug("[PushAfterHandshake]推送大厅列表,userId:%v, leagueCnt:%v, signupCnt:%v", u.ID, len(openList), len(signuplist))
	// 如果用户有比赛，主动推送比赛信息
	hasRace := false
	// 判断用户是否已报名
	if loadRaceId := model.GetUserRace(u.ID); loadRaceId > 0 {
		if loadRaceInfo := model.RaceList.Get(loadRaceId); loadRaceInfo != nil {
			raceUserInfo := model.GetRaceUserInfo(loadRaceId, u.ID)
			if raceUserInfo != nil {
				hasRace = true
				// 推送正在用户的当前比赛信息
				u.SendMessage(LeagueRacePush(loadRaceInfo, raceUserInfo, model.LeagueList.Get(loadRaceInfo.LeagueId)))
				core.Logger.Debug("[PushAfterHandshake]推送正在进行的比赛信息,userI:%v, raceId:%v, leagueId:%v", u.ID, loadRaceId, loadRaceInfo.LeagueId)
			} else {
				// 删除用户的已报名比赛
				model.DelUserRace(u.ID)
				core.Logger.Warn("[PushAfterHandshake]用户数据错误，内存中认为用户已报名比赛，但是报名列表中却没有这个用户, userId:%v, raceId:%v", u.ID, loadRaceId)
			}
		} else {
			// 删除用户的已报名比赛
			model.DelUserRace(u.ID)
			core.Logger.Warn("[PushAfterHandshake]内存中存在用户的报名比赛，但是比赛未找到, userId:%v, loadedRaceId:%v", u.ID, loadRaceId)
		}
	}

	// 如果用户没有比赛，那么看一下有没有最后结果
	if !hasRace {
		// 推送一个空的比赛
		u.SendMessage(LeagueRaceNilPush())

		if impacket := hall.GetLastRaceResult(u.ID); impacket != nil {
			core.Logger.Debug("[PushAfterHandshake]推送比赛结果:%v", u.ID)
			u.SendMessage(impacket)
		}
	}
}

// LegueListRequest 客户端请求获取大厅列表
func LegueListRequest(userId int, impacket *protocal.ImPacket) {
	u, online := hall.UserSet.Get(userId)
	if !online {
		core.Logger.Error("[LegueListRequest]user not online, userId:%v", userId)
		// return nil, ierror.NewError(-202)
		return
	}

	core.Logger.Info("[LegueListRequest]userId:%v", userId)

	PushAfterHandshake(u)

	// // 开放中的比赛列表
	// openList := model.LeagueList.GetOpenList()
	// // 读取报名中的比赛列表
	// signupRaceList := model.RaceList.GetSignupList()

	// responsePacket := LeagueListResponse(impacket.GetMessageNumber(), openList, signupRaceList)
	// hall.PrivateMessage(userId, responsePacket)

	// core.Logger.Info("[LegueListRequest]userId:%v", userId)
}

// ApplyRequest 用户申请报名
func ApplyRequest(userId int, impacket *protocal.ImPacket) *ierror.Error {
	request := fbsCommon.GetRootAsLeagueApplyRequest(impacket.GetBody(), 0)
	// 用户未登录，什么都不做
	u, _ := hall.UserSet.Get(userId)
	if u == nil {
		return ierror.NewError(-10200, userId)
	}

	// 读取比赛模板
	leagueId := int(request.LeagueId())
	if leagueId == 0 {
		return ierror.NewError(-10101, "league.apply", "leagueId", leagueId)
	}
	leagueInfo := model.LeagueList.Get(leagueId)
	if leagueInfo == nil {
		return ierror.NewError(-10500, leagueId, userId)
	}

	// 判断用户是否已报名
	if loadRaceId := model.GetUserRace(userId); loadRaceId > 0 {
		if loadRaceInfo := model.RaceList.Get(loadRaceId); loadRaceInfo != nil {
			return ierror.NewError(-10503, loadRaceInfo.LeagueId, loadRaceId, userId)
		}
		// 删除用户的已报名比赛
		model.DelUserRace(userId)
		core.Logger.Warn("[league.ApplyRequest]内存中存在用户的报名比赛，但是比赛已经找不到了，修正用户数据，允许用户报名, userId:%v, loadedRaceId:%v", userId, loadRaceId)
	}

	// 判断比赛是否允许报名
	t := util.GetTime()
	signupTime, _, startTime := leagueInfo.CalcLeagueRaceTime()
	if t < signupTime || (startTime > 0 && t > startTime) {
		return ierror.NewError(-10501, leagueId, userId, util.FormatUnixTime(signupTime), util.FormatUnixTime(startTime))
	}

	// 读取比赛数据
	model.RaceList.Mux.Lock()
	defer model.RaceList.Mux.Unlock()
	raceInfo, _ := model.RaceList.GetSignupNS(leagueId)
	// raceInfo, newFlag := model.RaceList.GetSignupNS(leagueId)
	if raceInfo == nil {
		return ierror.NewError(-1)
	}

	// 判断报名人数是否已满
	if raceInfo.IsFull() {
		return ierror.NewError(-10502, leagueId, raceInfo.Id, userId)
	}

	core.Logger.Debug("[ApplyRequest]准备报名比赛, userId:%v, leagueId:%v, raceId:%v", userId, leagueId, raceInfo.Id)

	// 扣费
	var consume map[int]int
	var chargeErr *ierror.Error
	if leagueInfo.Price > 0 {
		consume, chargeErr = ChargeEntity(userId, leagueInfo.PriceEntityId, leagueInfo.Price, config.MONEY_CHANGE_TYPE_XF, config.MONEY_CONSUME_TYPE_LEAGUE)
		if chargeErr != nil {
			return chargeErr
		}
	}

	// 获取比赛的报名用户列表
	raceUsers := model.GetRaceUsers(raceInfo.Id)
	if raceUsers == nil {
		// 新生成一个列表
		raceUsers = model.AddRaceUsers(raceInfo.Id, raceInfo.LeagueId)
	}

	// 生成用户报名信息
	raceUserInfo, err := raceUsers.New(userId, raceInfo.Id, leagueInfo.Price)
	if err != nil {
		return ierror.NewError(-10504, leagueId, raceInfo.Id, userId, err.Error())
	}
	// 记录用户消耗
	raceUserInfo.Consume = consume

	// 更新&广播人数更新消息
	// 如果是支持虚拟人数，则广播的是虚拟人数
	raceInfo.SignupUserCount++
	if leagueInfo.EnableSimulationUserCount() {
		leagueInfo.SimulationUserCount++
		hall.BrodcastMessage(LeagueRaceSignupCountPush(leagueId, raceInfo.Id, leagueInfo.SimulationUserCount))
	} else {
		hall.BrodcastMessage(LeagueRaceSignupCountPush(leagueId, raceInfo.Id, raceInfo.SignupUserCount))
	}

	// 通知用户报名成功
	u.SendMessage(LeagueApplyResponse(impacket.GetMessageNumber(), raceInfo, raceUserInfo, leagueInfo))

	core.Logger.Info("[league.ApplyRequest]leagueId:%v, userId:%v, raceId:%v, Signup/min/max:%v/%v/%v, consume:%v",
		leagueId, userId, raceInfo.Id, raceInfo.SignupUserCount, raceInfo.RequireUserMin, raceInfo.RequireUserCount, consume)

	// 非定时赛已满，通知排赛
	// 如果支持虚拟人数，则虚拟人数满了，也开赛
	if leagueInfo.EnableSimulationUserCount() {
		if leagueInfo.SimulationUserCount >= leagueInfo.RequireUserCount {
			leagueInfo.SimulationUserCount = 0
			RacePlan(raceInfo)
			return nil
		}
	} else {
		if leagueInfo.StartCondition == config.LEAGUE_START_CONDITION_TEMP && raceInfo.IsFull() {
			RacePlan(raceInfo)
			return nil
		}
	}

	// 是否支持机器人自动加入
	if !raceInfo.RobotJoinStarted && raceInfo.SignupUserCount == 1 && leagueInfo.EnableAutoApply() {
		raceInfo.RobotJoinStarted = true
		robotCnt := leagueInfo.RequireUserMin - 1
		go autoApply(raceInfo, robotCnt, leagueInfo.RobotJoinInterval)
	}

	return nil
}

// RacePlan 更新比赛状态、通知拍塞
func RacePlan(raceInfo *model.Race) {
	raceInfo.Status = config.RACE_STATUS_PLAN
	if _, err := raceInfo.Update(nil, "status"); err != nil {
		core.Logger.Warn("[ApplyRequest]更新比赛状态为排赛状态失败, raceId:%v", raceInfo.Id)
	}
	// 写入至排赛队列
	model.LeagueRacePlanChannel <- raceInfo.Id
	// 再次更新比赛比赛人数
	hall.BrodcastMessage(LeagueRaceSignupCountPush(raceInfo.LeagueId, 0, 0))
}

// CancelRequest 用户取消报名
func CancelRequest(userId int, impacket *protocal.ImPacket) *ierror.Error {
	// 用户未登录，什么都不做
	u, _ := hall.UserSet.Get(userId)
	if u == nil {
		return ierror.NewError(-10200, userId)
	}

	// 判断用户是否已报名
	raceId := model.GetUserRace(userId)
	var raceInfo *model.Race
	if raceId > 0 {
		if raceInfo = model.RaceList.Get(raceId); raceInfo == nil {
			// 删除用户的已报名比赛
			model.DelUserRace(userId)
			core.Logger.Warn("[league.CancelRequest]用户比赛数据未找到, userId:%v, loadedRaceId:%v", userId, raceId)
		}
	}
	// 判断户未报名
	if raceId == 0 || raceInfo == nil {
		return ierror.NewError(-10508, raceId, userId)
	}

	t := util.GetTime()
	// 比赛已开始，不允许取消
	if raceInfo.Status != config.RACE_STATUS_SIGNUP {
		return ierror.NewError(-10506, raceId, userId, util.FormatUnixTime(raceInfo.StartTime))
	}
	// 比赛快开始，不允许取消
	if raceInfo.GiveupTime > 0 && t >= raceInfo.GiveupTime {
		return ierror.NewError(-10507, raceId, userId, util.FormatUnixTime(raceInfo.GiveupTime))
	}

	// 查看用户是否已报名参赛
	raceUsers := model.GetRaceUsers(raceId)
	var raceUserInfo *model.RaceUser
	if raceUsers != nil {
		raceUserInfo = raceUsers.Get(userId)
		if raceUserInfo != nil {
			// 删除报名信息
			err := raceUsers.Del(userId, raceUserInfo.Id)
			ierror.MustNil(err)
		} else {
			core.Logger.Warn("[league.CancelRequest]raceUserInfo==nil, userId:%v, raceId:%v", userId, raceId)
		}
	} else {
		core.Logger.Warn("[league.CancelRequest]raceUsers==nil, userId:%v, raceId:%v", userId, raceId)
	}

	// 删除用户已报名的比赛id
	model.DelUserRace(userId)

	// 退费
	var chargeEntities map[int]int
	var chargeErr *ierror.Error
	if raceUserInfo != nil && raceUserInfo.Consume != nil && len(raceUserInfo.Consume) > 0 {
		chargeEntities, chargeErr = ChargeReturnEntities(userId, raceUserInfo.Consume, config.MONEY_CHANGE_TYPE_TF, config.MONEY_TRANS_TYPE_LEAGUE)
		if chargeErr != nil {
			core.Logger.Error("[league.CancelRequest]退费失败,raceId:%v, userId:%v, consume:%v, err:%v", raceId, userId, raceUserInfo.Consume, chargeErr.Error())
		}
		core.Logger.Debug("[CancelRequest]userId:%v, raceId:%v, consume:%+v", userId, raceId, chargeEntities)
	} else {
		core.Logger.Warn("[league.CancelRequest]未找到用户付费信息, userId:%v, raceId:%v", userId, raceId)
	}
	leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)

	// 更新&广播比赛参与人数更新消息
	raceInfo.SignupUserCount--
	if leagueInfo.EnableSimulationUserCount() {
		leagueInfo.SimulationUserCount--
		hall.BrodcastMessage(LeagueRaceSignupCountPush(leagueInfo.Id, raceInfo.Id, leagueInfo.SimulationUserCount))
	} else {
		hall.BrodcastMessage(LeagueRaceSignupCountPush(leagueInfo.Id, raceInfo.Id, raceInfo.SignupUserCount))
	}

	// 通知用户取消报名成功
	u.SendMessage(LeagueCancelResponse(impacket.GetMessageNumber(), nil))

	core.Logger.Info("[league.CancelRequest]leagueId:%v, userId:%v, raceId:%v, Signup/min/max:%v/%v/%v, charge entities:%+v", raceInfo.LeagueId, userId, raceInfo.Id, raceInfo.SignupUserCount, raceInfo.RequireUserMin, raceInfo.RequireUserCount, chargeEntities)

	return nil
}

// GiveupRequest 晋级赛过程中，用户放弃比赛
// 报名中的时候，不允许放弃
// 本轮比赛，本人已打完，其他人未打完的时候，才允许退赛
func GiveupRequest(userId int, impacket *protocal.ImPacket) *ierror.Error {
	// 用户未登录，什么都不做
	u, _ := hall.UserSet.Get(userId)
	if u == nil {
		return ierror.NewError(-10200, userId)
	}

	// check user race exists
	raceId := model.GetUserRace(userId)
	var raceInfo *model.Race
	if raceId > 0 {
		if raceInfo = model.RaceList.Get(raceId); raceInfo == nil {
			// 删除用户的已报名比赛
			model.DelUserRace(userId)
			core.Logger.Warn("[league.GiveupRequest]内存中存在用户的报名比赛，但是比赛数据找不到, userId:%v, loadedRaceId:%v", userId, raceId)
		}
	}
	// 用户未报名
	if raceId == 0 || raceInfo == nil {
		return ierror.NewError(-10508, raceId, userId)
	}

	// 查看用户是否已报名参赛
	raceUsers := model.GetRaceUsers(raceId)
	if raceUsers != nil {
		raceUserInfo := raceUsers.Get(userId)
		// 如果未找到用户报名信息，则认为用户操作正常的，直接返回
		if raceUserInfo != nil {
			// 进行中的，『允许放弃』的用户才可以退出比赛
			if raceInfo.Status != config.RACE_STATUS_PLAY ||
				raceUserInfo.GiveupStatus != config.RACE_USER_GIVEUP_STATUS_ALLOW {
				return ierror.NewError(-10510, raceId, userId, raceInfo.Status, raceUserInfo.GiveupStatus)
			}

			// 更新用户信息
			raceUserInfo.Status = config.RACE_USER_STATUS_GIVEUP
			raceUserInfo.GiveupTime = util.GetTime()
			core.GetWriter().Update(raceUserInfo, "status", "giveup_time")

			// 删除用户已报名的比赛id
			model.DelUserRace(userId)

			// 从比赛用户排名的集合中删除
			model.DelRaceUserRank(raceId, []int{userId})

			// 删除之前存储的比赛结果
			hall.LastRaceResultSet.Delete(userId)

			// 生成用户的最终排名
			rankInfo := &model.RaceRank{
				RaceId: raceId,
				UserId: userId,
				Rank:   0,
				Round:  raceUserInfo.Round,
				Score:  raceUserInfo.Score,
				Status: config.RACE_USER_STATUS_GIVEUP,
			}
			_, err := core.GetWriter().Insert(rankInfo)
			ierror.MustNil(err)
		} else {
			core.Logger.Warn("[league.GiveupRequest]raceUserInfo==nil, userId:%v, raceId:%v", userId, raceId)
		}
	} else {
		core.Logger.Warn("[league.GiveupRequest]raceUsers==nil, userId:%v, raceId:%v", userId, raceId)
	}

	// 推送用户退赛结果
	u.SendMessage(LeagueQuitResponse(impacket.GetMessageNumber(), raceInfo, model.LeagueList.Get(raceInfo.LeagueId), nil))

	// 广播排名变化到所有游戏服
	hall.BrodcastMessageToGameServers(LeagueL2SRankRefreshPush(raceId))

	core.Logger.Info("[GiveupRequest]用户退赛, raceId:%v, round:%v, userId:%v", raceInfo.Id, raceInfo.Round, userId)

	return nil
}

// PlanResultPush 排赛结果
func PlanResultPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsLeagueS2LPlanPush(impacket.GetBody(), 0)
	raceRoomId := response.RaceRoomId()
	raceId := response.RaceId()
	roomId := response.RoomId()
	gameResult := new(fbsCommon.GameResult)
	response.S2cResult(gameResult)
	if gameResult.Code() != 0 {
		// TODO 排赛失败，需要通知重排
		core.Logger.Error("[PlanResultPush]排赛失败，需要通知排赛进程，重新排赛, raceId:%v, roomId:%v, raceRoomId:%v, code:%v, msg:%v", raceId, roomId, raceRoomId, gameResult.Code(), string(gameResult.Msg()))
		return
	}

	raceInfo := model.RaceList.Get(raceId)
	if raceInfo == nil || raceInfo.PlanRooms == nil {
		core.Logger.Error("[PlanResultPush]比赛排赛中的房间异常, raceInfo:%v, planRooms:%+v", raceInfo, raceInfo.PlanRooms)
		return
	}
	// 删除房间的等待排赛状态
	raceInfo.DelPlanWait(raceRoomId)

	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms == nil {
		core.Logger.Error("[PlanResultPush]raceRooms==nil,raceId:%v, raceRoomId:%v, roomId:%v", raceId, raceRoomId, roomId)
		return
	}
	raceRoom := raceRooms.Get(raceRoomId)
	if raceRoom == nil {
		core.Logger.Error("[PlanResultPush]raceRooms==nil,raceId:%v, raceRoomId:%v, roomId:%v", raceId, raceRoomId, roomId)
	}

	raceRoom.RoomId = int64(roomId)
	raceRoom.UpdateTime = util.GetTime()
	// 通知用户比赛开始
	push := LeagueGameStartPush(raceId, roomId)
	for _, userId := range raceRoom.GetUsers() {
		// 如果是机器人，通知重连
		// 如果是用户，通知比赛开始
		if !robot.IsRobotUser(userId) {
			hall.PrivateMessage(userId, push)
		} else {
			robotGameInfo := robot.NewGameInfo(raceRoom.ServerRemote, 0, 0, 0, 1)
			robotGameInfo.RobotId = userId
			AddHallRobotRoom(raceRoom.ServerRemote, robotGameInfo.String())
		}
		if raceUserInfo := model.GetRaceUserInfo(raceId, userId); raceUserInfo != nil {
			raceUserInfo.RoomId = raceRoom.RoomId
		}
	}
	// 开启比赛监听
	raceRoom.ActiveTime = util.GetTime()
	go listenRoomActive(raceRoom)
	core.Logger.Info("[PlanResultPush]raceId:%v, raceRoomId:%v, roomId:%v", raceId, raceRoomId, roomId)
}

// GameFinish 单局结束
func GameFinish(impacket *protocal.ImPacket) {
	// 通知客户端更新排名
	push := fbsCommon.GetRootAsLeagueS2LRoundFinishPush(impacket.GetBody(), 0)
	raceId := push.RaceId()
	raceRoomId := push.RaceRoomId()
	scores := make([]int, 0, push.ScoresLength())
	for i := 0; i < push.ScoresLength(); i++ {
		scores = append(scores, int(push.Scores(i)))
	}
	core.Logger.Debug("[RoundFinish]parse request data, raceId:%v, raceRoomId:%v, scores:%+v", raceId, raceRoomId, scores)

	// 处理比赛单局结束
	gameFinish(raceId, raceRoomId, scores)
}

// 广播比赛排名变化给大厅的用户
func refreshHallRank(raceId int64, playingRoomCount int) {
	ranks := model.GetRaceUserRanks(raceId)
	rank := 0
	for i := 0; i < len(ranks); i += 2 {
		userId := ranks[i]
		raceUserInfo := model.GetRaceUserInfo(raceId, ranks[i])
		if raceUserInfo == nil {
			core.Logger.Debug("[refreshHallRank]用户排名存在，但是raceUserInfo中找不到, raceId:%v, userId:%v", raceId, userId)
			continue
		}
		// 只有未淘汰的用户，才进行排名更新
		if raceUserInfo.Status == config.RACE_USER_STATUS_SIGNUP {
			// 更新内存中的用户的排名
			if playingRoomCount < 1 {
				if raceInfo := model.RaceList.Get(raceId); raceInfo != nil {
					if leagueInfo := model.LeagueList.Get(raceInfo.LeagueId); leagueInfo != nil {
						if leagueInfo.SkipRank > 0 {
							playingRoomCount = 1
						}
					}
				}
			}
			rank++
			raceUserInfo.Rank = rank
			// 推给客户端的房间数，最小是1
			hall.PrivateMessage(userId, LeagueGameRankPush(rank, raceUserInfo.Score, playingRoomCount))
			core.Logger.Debug("[refreshHallRank]更新比赛用户排名, raceId:%v, userId:%v, rank:%v", raceId, userId, rank)
		} else {
			core.Logger.Debug("[refreshHallRank]用户已退赛或者被淘汰, 不更新排名, raceId:%v, userId:%v, 最终rank:%v", raceId, userId, raceUserInfo.Rank)
		}
	}
}

// RoomFinish 游戏结束
func RoomFinish(impacket *protocal.ImPacket) {
	push := fbsCommon.GetRootAsLeagueS2LGameFinishPush(impacket.GetBody(), 0)
	raceId := push.RaceId()
	raceRoomId := push.RaceRoomId()
	code := int(push.Code())
	core.Logger.Debug("[RoomFinish]parse request data, raceId:%v, raceRoomId:%v, code:%v", raceId, raceRoomId, code)

	roomFinish(raceId, raceRoomId, code)
}

// RoomActive 收到房间活跃通知
func RoomActive(impacket *protocal.ImPacket) {
	// 更新房间的活跃时间
	push := fbsCommon.GetRootAsLeagueS2LGameActivePush(impacket.GetBody(), 0)
	raceId := push.RaceId()
	raceRoomId := push.RaceRoomId()
	raceRoom := model.GetRaceRoom(raceId, raceRoomId)
	if raceRoom == nil {
		core.Logger.Warn("[RoomActive]raceRoom==nil, raceId:%v, raceRoomId:%v", raceId, raceRoomId)
	}
	raceRoom.ActiveTime = util.GetTime()
	core.Logger.Info("[RoomActive]raceId:%v, raceRoomId:%v", raceId, raceRoomId)
}

// RaceResultReceived 客户端收到比赛结果的反馈
func RaceResultReceived(userId int, impacket *protocal.ImPacket) {
	// 更新房间的活跃时间
	push := fbsCommon.GetRootAsLeagueRaceResultRecievedNotify(impacket.GetBody(), 0)
	raceId := push.RaceId()
	// 删除比赛结果
	hall.LastRaceResultSet.Delete(userId)
	core.Logger.Info("[RaceResultReceived]userId:%v, raceId:%v", userId, raceId)
}

// 通知info进行发奖
func noticeInfoRaceRewardsCompleted(raceId int64) {
	url := core.GetAppConfig().RewardsNotifyUrl + fmt.Sprintf("?race_id=%v", raceId)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[推送发奖通知]het.Get, raceId:%v, url:%v, error:%v", raceId, url, err.Error())
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			core.Logger.Error("[��送发奖通知]het.Get read body, raceId:%v, url:%v, error:%v", raceId, url, err.Error())
		} else {
			core.Logger.Info("[推送发奖通知]success, raceId:%v, url:%v, result:%v", raceId, url, string(body))
		}
	}
}
