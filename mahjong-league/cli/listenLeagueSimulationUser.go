package cli

import (
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/model"
	"mahjong-league/robot"
	"mahjong-league/service"
	"sync"
	"time"

	"github.com/fwhappy/util"
)

var simulationList *sync.Map

func openListenLeagueSimulationUser(l *model.League) {
	if simulationList == nil {
		simulationList = &sync.Map{}
	}

	if _, ok := simulationList.Load(l.Id); !ok {
		simulationList.Store(l.Id, util.GetTime())
		go listenLeagueSimulationUser(l)
	}
}

func listenLeagueSimulationUser(l *model.League) {
	core.Logger.Info("[listenLeagueSimulationUser]开启联赛的模拟参与人数计算,leagueId:%v", l.Id)
	defer func() {
		simulationList.Delete(l.Id)
		core.Logger.Info("[listenLeagueSimulationUser]关闭联赛的模拟参与人数计算,leagueId:%v", l.Id)
	}()
	for {
		// 如果联赛已经下架，则关闭
		league := model.LeagueList.Get(l.Id)
		if league == nil {
			return
		}

		// 未开放
		if !league.IsOpen(0) {
			time.Sleep(time.Second)
			continue
		}

		// 不支持虚拟人数
		if !league.EnableSimulationUserCount() {
			time.Sleep(time.Second)
			continue
		}

		// 执行sleep
		robot.JoinSleep(league.RobotJoinInterval)

		// 将报名人数加1
		league.SimulationUserCount++
		// core.Logger.Debug("[listenLeagueSimulationUser]联赛的模拟人数+1, leagueId:%v, simulation count:%v", league.Id, league.SimulationUserCount)

		// 获取报名中的比赛id
		var raceID int64
		raceInfo := model.RaceList.GetSignup(league.Id)
		if raceInfo != nil {
			raceID = raceInfo.Id
		}

		// 广播报名人数
		hall.BrodcastMessage(service.LeagueRaceSignupCountPush(league.Id, raceID, league.SimulationUserCount))

		// 满了就从头开始
		if league.SimulationUserCount >= league.RequireUserCount {
			core.Logger.Info("[listenLeagueSimulationUser]联赛的模拟人数已达上限, leagueId:%v, raceId:%v", league.Id, raceID)
			// 如果有真实玩家，则去安排排赛
			if raceID > 0 {
				raceUsers := model.GetRaceUsers(raceID)
				if raceUsers != nil {
					// 是否有真实玩家
					hasRealUser := false
					raceUsers.Mux.RLock()
					for _, raceUser := range raceUsers.Users {
						if !robot.IsRobotUser(raceUser.UserId) {
							hasRealUser = true
							break
						}
					}
					raceUsers.Mux.RUnlock()
					core.Logger.Debug("hasRealUser:%v", hasRealUser)

					// 有真实玩家，安排排赛
					if hasRealUser {
						core.Logger.Info("[listenLeagueSimulationUser]有真实玩家，开始排赛, leagueId:%v, raceId:%v", league.Id, raceID)
						service.RacePlan(raceInfo)
					} else {
						core.Logger.Info("[listenLeagueSimulationUser]无真实玩家，开始下一轮模拟, leagueId:%v, raceId:%v", league.Id, raceID)
					}
				}
			} else {
				core.Logger.Info("[listenLeagueSimulationUser]无真实玩家，开始下一轮模拟, leagueId:%v", league.Id)
			}

			league.SimulationUserCount = 0
			// 一个很短的时间，显示人满
			time.Sleep(100 * time.Millisecond)
			hall.BrodcastMessage(service.LeagueRaceSignupCountPush(league.Id, 0, 0))
		}
	}
}
