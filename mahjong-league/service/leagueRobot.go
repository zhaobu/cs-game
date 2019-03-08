package service

import (
	"fmt"
	"mahjong-league/config"
	"mahjong-league/core"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/protocal"
	"mahjong-league/robot"

	"github.com/fwhappy/util"
)

// 自动将机器人加入房间
func autoApply(raceInfo *model.Race, robotCnt, interval int) {
	defer util.RecoverPanic()

	for i := 0; i < robotCnt; i++ {
		if raceInfo.IsFull() {
			break
		}
		robot.JoinSleep(interval)
		if raceInfo.Status != config.RACE_STATUS_SIGNUP {
			break
		}
		go robotApply(raceInfo)
	}
}

// 机器人报名
func robotApply(raceInfo *model.Race) {
	defer util.RecoverPanic()

	// 锁比赛，纺织并发
	model.RaceList.Mux.Lock()
	defer model.RaceList.Mux.Unlock()

	if raceInfo.IsFull() {
		core.Logger.Warn("[robotApply]机器人加入失败, raceId:%v, reason:比赛人数已满。", raceInfo.Id)
		return
	}

	leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)
	if leagueInfo == nil {
		core.Logger.Warn("[robotApply]机器人加入失败,raceId:%v, reason:比赛信息未找到", raceInfo.Id)
		return
	}

	if !leagueInfo.IsOpen(0) {
		core.Logger.Warn("[robotApply]机器人加入失败,raceId:%v, reason:比赛未开放", raceInfo.Id)
		return
	}

	// 获取比赛的报名用户列表
	raceUsers := model.GetRaceUsers(raceInfo.Id)
	if raceUsers == nil {
		core.Logger.Warn("[robotApply]机器人加入失败,raceId:%v, reason:raceUsers==nil", raceInfo.Id)
		return
	}

	robots := robot.Fetch(1)
	if len(robots) == 0 {
		core.Logger.Warn("[robotApply]机器人加入失败,raceId:%v, reason:获取机器人失败", raceInfo.Id)
		return
	}
	// 生成用户报名信息
	_, err := raceUsers.New(robots[0], raceInfo.Id, 0)
	if err != nil {
		robot.Remove(robots[0])
		core.Logger.Warn("[robotApply]机器人加入失败,raceId:%v, reason:添加比赛用户失败, err:%v", raceInfo.Id, err.Error())
		return
	}
	// 更新&广播人数更新消息
	raceInfo.SignupUserCount++
	hall.BrodcastMessage(LeagueRaceSignupCountPush(raceInfo.LeagueId, raceInfo.Id, raceInfo.SignupUserCount))

	core.Logger.Info("[robotApply]机器人加入成功,raceId:%v,leagueId:%v", raceInfo.Id, raceInfo.LeagueId)

	// 非定时赛已满，通知排赛
	if leagueInfo.StartCondition == config.LEAGUE_START_CONDITION_TEMP && raceInfo.IsFull() {
		RacePlan(raceInfo)
	}
}

// RobotApply 机器人报名
func RobotApply(userId int, impacket *protocal.ImPacket) *ierror.Error {
	// 用户未登录，什么都不做
	u, _ := hall.UserSet.Get(userId)
	if u == nil {
		return ierror.NewError(-10200, userId)
	}

	request := fbsCommon.GetRootAsLeagueRobotApplyRequest(impacket.GetBody(), 0)
	// 读取比赛模板
	leagueId := int(request.LeagueId())
	// 读取比赛信息
	raceId := int64(request.RaceId())
	if leagueId == 0 {
		return ierror.NewError(-10101, "league.RobotApply", "leagueId", leagueId)
	}
	if raceId == 0 {
		return ierror.NewError(-10101, "league.RobotApply", "leagueId", raceId)
	}
	leagueInfo := model.LeagueList.Get(leagueId)
	if leagueInfo == nil {
		return ierror.NewError(-10500, leagueId, userId)
	}
	raceInfo := model.RaceList.GetSignup(leagueId)
	if raceInfo == nil || raceInfo.Id != raceId {
		return ierror.NewError(-10504, leagueId, raceId, userId, "没有正在报名中的比赛")
	}

	// 判断报名人数是否已满
	if raceInfo.IsFull() {
		return ierror.NewError(-10502, leagueId, raceInfo.Id, userId)
	}

	// 获取比赛的报名用户列表
	raceUsers := model.GetRaceUsers(raceInfo.Id)
	if raceUsers == nil {
		return ierror.NewError(-10504, leagueId, raceInfo.Id, userId, "未找到报名中的用户列表")
	}

	// 生成用户报名信息
	raceUserInfo, err := raceUsers.New(userId, raceInfo.Id, leagueInfo.Price)
	if err != nil {
		return ierror.NewError(-10504, leagueId, raceInfo.Id, userId, err.Error())
	}

	// 更新&广播人数更新消息
	raceInfo.SignupUserCount++
	hall.BrodcastMessage(LeagueRaceSignupCountPush(leagueId, raceInfo.Id, raceInfo.SignupUserCount))

	// 通知用户报名成功
	u.SendMessage(LeagueApplyResponse(impacket.GetMessageNumber(), raceInfo, raceUserInfo, leagueInfo))

	core.Logger.Info("[league.RobotApplyRequest]leagueId:%v, userId:%v, raceId:%v, Signup/min/max:%v/%v/%v", leagueId, userId, raceInfo.Id, raceInfo.SignupUserCount, raceInfo.RequireUserMin, raceInfo.RequireUserCount)

	// 非定时赛已满，通知排赛
	if leagueInfo.StartCondition == config.LEAGUE_START_CONDITION_TEMP && raceInfo.IsFull() {
		RacePlan(raceInfo)
	}

	return nil
}

// 获取机器人房间列表的cachekey
func getHallRobotRoomListCacheKey(remote string) string {
	return fmt.Sprintf(config.CACHE_KEY_HALL_ROBOT_ROOM_LIST)
}

// AddHallRobotRoom 添加一个房间到redis中的机器人房间队列
func AddHallRobotRoom(remote string, data string) {
	core.RedisDo(core.RedisClient3, "lpush", getHallRobotRoomListCacheKey(remote), data)
}
