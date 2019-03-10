package cli

import (
	"fmt"
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/model"
	"mahjong-league/robot"
	"mahjong-league/service"
	"os"
	"time"

	"github.com/fwhappy/util"
)

// 可报名比赛列表
var canSignLeagues map[int]int

// ListenLeagueListRefresh 监听league list的变化
func ListenLeagueListRefresh() {
	defer util.RecoverPanic()
	defer core.Logger.Info("[ListenLeagueListRefresh]completed")

	// 可报名比赛列表
	canSignLeagues = make(map[int]int, 0)

	core.Logger.Info("[ListenLeagueListRefresh]started")
	for {
		time.Sleep(time.Duration(config.MODEL_REFRESH_INTERVAL) * time.Second)
		handleLeagueListRefresh()
	}
}

// 处理LeagueList的修改与新增
// 读取新的列表，和当前的列表比较，如果有新增的，则添加，有下架的
func handleLeagueListRefresh() {
	// 获取当前数据库中的比赛列表
	leagueList := model.LoadLeagueListFromDB()
	for i := 0; i < len(leagueList); i++ {
		// db中的联赛信息
		newLeagueInfo := leagueList[i]
		// 内存中的联赛信息
		oldLeagueInfo := model.LeagueList.Get(newLeagueInfo.Id)

		// 开启监听
		openListenLeagueSimulationUser(&newLeagueInfo)

		// 旧的开放标志
		oldOpenFlag := oldLeagueInfo != nil && oldLeagueInfo.IsOpen(util.GetTime()-10)
		// 新比赛的开放标志
		newOpenFlag := newLeagueInfo.IsOpen(util.GetTime())
		// 旧的可以报名， 新比赛可以报名标志
		oldSignup, newSignup := false, false
		// 下架/关闭标志、需更新标志
		closeFlag, updateFlag := false, false
		// 计算比赛是否需要更新
		if oldLeagueInfo == nil || // 新增
			oldLeagueInfo.UpdateTime != newLeagueInfo.UpdateTime || // 修改
			oldOpenFlag != newOpenFlag { // 到了开放、关闭时间
			updateFlag = true
		}

		// 定时赛 如果刚刚还可以报名，但是现在不能报名了，则需要刷一次比赛列表
		if newLeagueInfo.StartCondition == config.LEAGUE_START_CONDITION_FIXED {
			// 旧的报名状态
			if _, exists := canSignLeagues[newLeagueInfo.Id]; exists {
				oldSignup = true
			}
			signupTime, _, _ := newLeagueInfo.CalcLeagueRaceTime()
			newSignup = util.GetTime() > signupTime
			if !oldSignup && newSignup {
				canSignLeagues[newLeagueInfo.Id] = newLeagueInfo.Id
				core.Logger.Info("[handleLeagueListRefresh]检测到开启报名的比赛:%v, 报名开始时间:%v", newLeagueInfo.Id, util.FormatUnixTime(signupTime))
			} else if oldSignup && !newSignup {
				delete(canSignLeagues, newLeagueInfo.Id)
				updateFlag = true
				core.Logger.Info("[handleLeagueListRefresh]检测到比赛报名结束:%v, 报名开始时间:%v", newLeagueInfo.Id, util.FormatUnixTime(signupTime))
			}
		}

		// 计算比赛是否新关闭，新关闭的比赛，需要将报名中的比赛解散
		if oldOpenFlag && !newOpenFlag {
			closeFlag = true
		}
		// core.Logger.Trace("[handleLeagueListRefresh]当前比赛状态, leagueId:%v, updateFlag:%v, oldOpenFlag:%v, newOpenFlag:%v, closeFlag:%v", newLeagueInfo.Id, updateFlag, oldOpenFlag, newOpenFlag, closeFlag)

		// 如果比赛更新了，需要更新内存中的比赛信息，并重新拉取比赛奖励
		if updateFlag {
			model.LeagueList.Set(&newLeagueInfo)
			core.Logger.Debug("[handleLeagueListRefresh]更新league_lists数据, leagueId:%v", newLeagueInfo.Id)
			// 更新LeagueRewards数据
			model.LeagueRewardsList.ReloadLeagueRewards(newLeagueInfo.Id)
			core.Logger.Debug("[handleLeagueListRefresh]更新league_rewards数据, leagueId:%v", newLeagueInfo.Id)
		}

		raceInfo := model.RaceList.GetSignup(newLeagueInfo.Id)
		if closeFlag {
			// 有比赛关闭或下架
			// 推送比赛关闭的消息
			hall.BrodcastMessage(service.LeagueListRemovePush(newLeagueInfo.Id))
			// 处理比赛结束, 关闭报名中的比赛
			go closeRace(raceInfo, &newLeagueInfo)
		} else if updateFlag {
			// 推送有比赛更新的消息
			if !oldOpenFlag && !newOpenFlag {
				// 一直处于下架状态， 不做任何更新
			} else {
				hall.BrodcastMessageWithVersion(service.LeagueListReloadPush(&newLeagueInfo, raceInfo), newLeagueInfo.RequireVerMin)
			}
		}
	}
	// core.Logger.Debug("[handleLeagueListRefresh]同步比赛列表完成")

	// 如果非正式环境，每次都重新加载配置文件
	if core.GetAppConfig().Env != "product" {
		_path := fmt.Sprintf("/opt/app/mahjong/src/mahjong.go/conf/%v/robot.toml", core.GetAppConfig().Env)
		_, err := os.Stat(_path)
		if err != nil && os.IsNotExist(err) {
			// core.Logger.Debug("测试文件未找到 path:%v", _path)
		} else {
			core.LoadRobotConfig(_path)
			// core.Logger.Info("加载机器人配置完成, config:%#v", core.RobotCfg)
		}
	}
}

// 当比赛开始时人未满或者下架时，解散比赛
func closeRace(raceInfo *model.Race, leagueInfo *model.League) {
	defer util.RecoverPanic()

	if raceInfo == nil {
		return
	}

	core.Logger.Debug("[closeRace]收到解散比赛请求, raceId:%v, leagueId:%v", raceInfo.Id, leagueInfo.Id)

	// 非报名状态的不能解散
	if raceInfo.Status != config.RACE_STATUS_SIGNUP {
		core.Logger.Error("[closeRace]比赛状态错误, 不能解散, raceId:%v, leagueId:%v, status:%v",
			raceInfo.Id, leagueInfo.Id, raceInfo.Status)
		return
	}

	// 更新比赛表状态为已解散
	raceInfo.Status = config.RACE_STATUS_DISMISS
	_, err := raceInfo.Update(nil, "status")
	core.Logger.Debug("[closeRace]更新比赛状态结果, raceId:%v, leagueId:%v, result:%v", raceInfo.Id, leagueInfo.Id, err)

	// 更新用户状态为已结束
	raceUsers := model.GetRaceUsers(raceInfo.Id)
	raceUsers.ToDismiss(raceInfo.Id)
	core.Logger.Debug("[closeRace]更新比赛用户状态、删除用户比赛对应关系,raceId:%v, leagueId:%v", raceInfo.Id, leagueInfo.Id)

	// 从比赛成员表删除、退费
	impacket := service.LeagueRaceCancelPush(raceInfo.Id)
	for _, raceUser := range raceUsers.Users {
		// 机器人无需退费
		if robot.IsRobot(raceUser.UserId) {
			robot.Remove(raceUser.UserId)
			continue
		}

		// 退费
		if raceUser.Consume != nil && len(raceUser.Consume) > 0 {
			m, err := service.ChargeReturnEntities(raceUser.UserId, raceUser.Consume, config.MONEY_CHANGE_TYPE_TF, config.MONEY_TRANS_TYPE_LEAGUE)
			if err != nil {
				core.Logger.Error("[closeRace]退费失败,raceId:%v, userId:%v, consume:%v, err:%v", raceInfo.Id, raceUser.UserId, raceUser.Consume, err.Error())
			}
			core.Logger.Debug("[closeRace]退费, raceId:%v, leagueId:%v, userId:%v, price:%v",
				raceInfo.Id, leagueInfo.Id, raceUser.UserId, m)
		} else {
			core.Logger.Warn("[closeRace]未找到用户付费信息, userId:%v, raceId:%v", raceUser.UserId, raceInfo.Id)
		}
		hall.PrivateMessage(raceUser.UserId, impacket)
	}

	core.Logger.Info("[closeRace]解散比赛完成, raceId:%v, leagueId:%v", raceInfo.Id, leagueInfo.Id)
}
