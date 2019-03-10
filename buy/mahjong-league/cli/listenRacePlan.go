package cli

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/gateway"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/robot"
	"mahjong-league/service"
	"mahjong-league/ss"
	"net"
	"sync"
	"time"

	fbsCommon "mahjong-league/fbs/Common"

	"github.com/fwhappy/util"
)

// 监听非定时塞排赛
func ListenFullRace() {
	core.Logger.Info("[listenFullRace] started...")
	defer util.RecoverPanic()
	defer core.Logger.Error("[listenFullRace]exited")

	for raceId := range model.LeagueRacePlanChannel {
		core.Logger.Info("[listenFullRace]raceId:%v", raceId)
		go handleRacePlan(raceId)
	}
}

// ListenFixRace 监听定时赛排赛
func ListenFixRace() {
	defer util.RecoverPanic()
	for {
		t := util.GetTime()
		model.RaceList.Mux.Lock()
		for _, raceInfo := range model.RaceList.Data {
			leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)
			if leagueInfo == nil {
				core.Logger.Error("[ListenFixRace]leagueInfo未找到, raceId:%v, leagueId:%v", raceInfo.Id, raceInfo.LeagueId)
				continue
			}
			if raceInfo.Status == config.RACE_STATUS_SIGNUP && // 报名中
				leagueInfo.StartCondition == config.LEAGUE_START_CONDITION_FIXED && // 定时塞
				raceInfo.StartTime >= 0 {
				if t >= raceInfo.StartTime { // 已经达到开始时间
					handleFixRacePlan(raceInfo, leagueInfo)
				} else if raceInfo.StartTime-t == 120 { // 踢出在线用户
					handleFixRaceKick(raceInfo, leagueInfo)
				}

			}
		}
		model.RaceList.Mux.Unlock()
		// 每秒执行一次
		time.Sleep(time.Second)
	}
}

// 处理定时赛的排赛
func handleFixRacePlan(raceInfo *model.Race, leagueInfo *model.League) {
	// 检查报名人数是否已够
	if raceInfo.IsEnough() {
		// 报名人数足够，安排排赛
		// 更新比赛状态
		raceInfo.Status = config.RACE_STATUS_PLAN
		if _, err := raceInfo.Update(nil, "status"); err != nil {
			core.Logger.Warn("[handleFixRacePlan]更新比赛状态为排赛状态失败, raceId:%v", raceInfo.Id)
		}
		core.Logger.Debug("[handleFixRacePlan]检测到定时赛开始, leagueId:%v, raceId:%v", raceInfo.LeagueId, raceInfo.Id)
		go handleRacePlan(raceInfo.Id)
	} else {
		// 报名人数不足，解散比赛
		core.Logger.Debug("[handleFixRacePlan]定时赛报名人数不足, 结算比赛, leagueId:%v, raceId:%v", raceInfo.LeagueId, raceInfo.Id)
		go closeRace(raceInfo, leagueInfo)
	}

	// 通知客户端刷新比赛列表
	hall.BrodcastMessage(service.LeagueListReloadPush(leagueInfo, nil))
}

// 排赛处理
func handleRacePlan(raceId int64) {
	defer util.RecoverPanic()
	core.Logger.Info("[handleRacePlan]start, raceId:%v", raceId)

	// 读取比赛信息
	raceInfo := model.RaceList.Get(raceId)
	if raceInfo == nil {
		core.Logger.Error("[handleRacePlan]比赛信息未找到, raceId:%v", raceId)
		return
	}

	// 读取比赛模板
	leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)
	if leagueInfo == nil {
		core.Logger.Error("[handleRacePlan]比赛模板信息未找到, leagueId:%v", raceInfo.LeagueId)
		return
	}

	// 读取待排赛用户列表
	raceUsers := model.GetRaceUsers(raceId)
	if raceUsers == nil {
		core.Logger.Error("[handleRacePlan]参与比赛的用户未找到, leagueId:%v, raceId:%v", leagueInfo.Id, raceId)
		return
	}

	// 所有比赛的用户id
	raceUserIds := make([]int, 0)
	// 已排赛用户id
	planedUserIds := make([]int, 0)
	// 需排赛用户列表
	needPlanUserIds := make([]int, 0)
	// 需要机器人数量
	robotCnt := 0
	// 已排赛房间列表
	planedRaceRoomIds := make([]int64, 0)
	// 需要推送排赛的房间列表
	needPlanRaceRoomIds := make(map[int64]bool, 0)

	// 找出已经参与分组的用户、找出已经成功创建了房间的用户
	// 此项主要用在服务重启时的数据恢复
	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms == nil {
		raceRooms = model.NewRaceRooms(raceId)
	} else {
		for _, raceRoom := range raceRooms.Data {
			if raceRoom.Round == raceInfo.Round || raceRoom.Status == config.RACE_ROOM_STATUS_NORMAL {
				ids := raceRoom.GetUsers()
				planedUserIds = append(planedUserIds, ids...)
				if raceRoom.RoomId == 0 {
					// needPlanRaceRoomIds = append(needPlanRaceRoomIds, raceRoom.Id)
					needPlanRaceRoomIds[raceRoom.Id] = true
				} else {
					planedRaceRoomIds = append(planedRaceRoomIds, raceRoom.Id)
				}
			}
		}
	}
	// 遍历所有比赛列表，找到所有需要排赛的用户
	planUsers := make(map[int]*model.RaceUser)
	raceUsers.Mux.Lock()
	for _, ru := range raceUsers.Users {
		// 跳过已被淘汰和状态不合的用户
		if ru.Round != raceInfo.Round || ru.Status != config.RACE_USER_STATUS_SIGNUP {
			continue
		}
		// 记录所有可排赛的用户
		raceUserIds = append(raceUserIds, ru.UserId)
		// 跳过已排赛的用户
		if util.IntInSlice(ru.UserId, planedUserIds) {
			continue
		}
		// 记录待排赛的用户
		planUsers[ru.UserId] = ru
		needPlanUserIds = append(needPlanUserIds, ru.UserId)
		// 标记用户为"禁止退赛"状态
		ru.GiveupStatus = config.RACE_USER_GIVEUP_STATUS_FORBID
	}
	raceUsers.Mux.Unlock()

	core.Logger.Debug("[handleRacePlan]raceId:%v, raceUserIds:%v", raceId, raceUserIds)
	core.Logger.Debug("[handleRacePlan]raceId:%v, planedUserIds:%v", raceId, planedUserIds)
	core.Logger.Debug("[handleRacePlan]raceId:%v, needPlanUserIds:%v", raceId, needPlanUserIds)
	core.Logger.Debug("[handleRacePlan]raceId:%v, needPlanRaceRoomIds:%+v", raceId, needPlanRaceRoomIds)
	core.Logger.Debug("[handleRacePlan]raceId:%v, planedRaceRoomIds:%v", raceId, planedRaceRoomIds)
	core.Logger.Info("[handleRacePlan]数据分析, raceId:%v, 总用户数:%v, 已排赛用户数:%v, 需排赛用户数:%v, 需排赛房间数:%v, 已排赛房间数:%v",
		raceId, len(raceUserIds), len(planedUserIds), len(needPlanUserIds), len(needPlanRaceRoomIds), len(planedRaceRoomIds))

	if len(raceUserIds) == 0 {
		core.Logger.Error("[handleRacePlan]未找到可以晋级的用户, leagueId:%v, raceId:%v", leagueInfo.Id, raceId)
		return
	}
	// 给未参与分组

	// 若人数不足，则补机器人
	if raceInfo.Round == 1 {
		// 如果支持氛围人数，则直接给补到氛围人数
		if leagueInfo.EnableSimulationUserCount() {
			if raceInfo.SignupUserCount < leagueInfo.RequireUserCount {
				robotCnt = leagueInfo.RequireUserCount - raceInfo.SignupUserCount
			}
		} else {
			if remainder := len(raceUserIds) % leagueInfo.GetRoomUserCount(); remainder > 0 {
				robotCnt = leagueInfo.GetRoomUserCount() - remainder
			}
		}
		raceInfo.SignupUserCount += robotCnt
		core.Logger.Debug("[handleRacePlan]第一轮比赛凑机器人, raceId:%v, 报名玩家数:%v, 补机器人数:%v, 几人局:%v, ", raceId, len(needPlanUserIds), robotCnt, leagueInfo.GetRoomUserCount())
	} else {
		// 非第一轮比赛，凑够晋级人数
		needUserCount := leagueInfo.GetRoundRequireUserCount(raceInfo.Round - 1)
		core.Logger.Debug("[handleRacePlan]raceId:%v, 本轮需要玩家数:%v, 未淘汰玩家数:%v", raceId, needUserCount, len(needPlanUserIds))
		robotCnt = needUserCount - len(raceUserIds)
	}

	if robotCnt > 0 {
		core.Logger.Debug("[handleRacePlan]raceId:%v, 人数不足, 需补充机器人数:%v", raceId, robotCnt)
		for _, robotId := range robot.Fetch(robotCnt) {
			// 插入到raceUser表
			// 生成用户报名信息
			raceUserInfo, _ := raceUsers.New(robotId, raceInfo.Id, leagueInfo.Price)
			raceUserInfo.Round = raceInfo.Round
			raceUserIds = append(raceUserIds, robotId)
			needPlanUserIds = append(needPlanUserIds, robotId)
			planUsers[robotId] = raceUserInfo
		}
		// 补机器人
		core.Logger.Debug("[handleRacePlan]raceId:%v, 补充机器人完成, 共机器人数:%v", raceId, robotCnt)
		core.Logger.Debug("[handleRacePlan]raceId:%v, raceUserIds:%v", raceId, raceUserIds)
		core.Logger.Debug("[handleRacePlan]raceId:%v, needPlanUserIds:%v", raceId, needPlanUserIds)
	}

	// 如果是第一轮，会给机器人加分
	if raceInfo.Round == 1 {
		robotFWScores := leagueInfo.GetRobotFWScores()
		if len(robotFWScores) > 0 {
			raceUsers.Mux.Lock()
			index := 0
			for _, raceUserInfo := range raceUsers.Users {
				if !robot.IsRobotUser(raceUserInfo.UserId) {
					continue
				}
				raceUserInfo.FwScores, _ = util.InterfaceToJsonString(robotFWScores[index])
				core.GetWriter().Update(raceUserInfo, "fw_score")
				core.Logger.Debug("[handleRacePlan]计算机器人的氛围积分, leagueId:%v, raceId:%v, robotId:%v, score:%v",
					leagueInfo.Id, raceInfo.Id, raceUserInfo.Id, raceUserInfo.FwScores)
				if index >= len(robotFWScores)-1 {
					break
				}
				index++
			}
			raceUsers.Mux.Unlock()
		}
	}

	// 更新比赛状态、给用户分组
	// 初始化一个orm连接，用于后面使用事务
	o := core.GetWriter()
	// 更新比赛开始时间
	raceInfo.StartTime = util.GetTime()
	o.Update(raceInfo, "start_time")
	// 分组
	// 打乱排序
	needPlanUserIds = util.ShuffleSliceInt(needPlanUserIds)
	roomUserCount := leagueInfo.GetRoomUserCount()
	for i := 0; i < len(needPlanUserIds)/roomUserCount; i++ {
		userIds := make([]int, roomUserCount)
		copy(userIds, needPlanUserIds[i*roomUserCount:(i+1)*roomUserCount])
		// 更新比赛用户信息
		for _, userId := range userIds {
			raceUser := planUsers[userId]
			raceUser.Round = raceInfo.Round
		}
		raceRoom := raceRooms.New(raceInfo, userIds, o)
		needPlanRaceRoomIds[raceRoom.Id] = true
		needPlanRaceRoomIds[raceRoom.Id] = true
	}
	model.RaceRoomsList.Store(raceId, raceRooms)
	core.Logger.Debug("[handleRacePlan]raceId:%v, 用户分组完成, 待排赛房间列表:%+v", raceId, needPlanRaceRoomIds)

	// 初始化比赛排名, 仅第一局
	if raceInfo.Round == 1 {
		model.InitRaceUserRank(raceId, raceUserIds)
	}

	// 开始排赛
	if len(needPlanRaceRoomIds) > 0 {
		raceInfo.PlanWait = make(chan int, 1024)
		raceInfo.PlanRooms = needPlanRaceRoomIds
		raceInfo.PlanRoomsMux = &sync.RWMutex{}

		times := 1
		for {
			if raceInfo.GetPlanWaitLen() == 0 {
				break
			}
			core.Logger.Debug("[handleRacePlan]raceId:%v, 第%v次分组, 剩余房间个数:%v", raceId, times, raceInfo.GetPlanWaitLen())

			// 选服，拉取可用的游戏服
			var conn *net.TCPConn
			var serverRemote string
			for {
				serverRemote = ss.Select()
				core.Logger.Debug("选服结果, remote:%v", serverRemote)
				if serverRemote == "" {
					core.Logger.Warn("[handleRacePlan]选服失败, 6秒后重新选服")
				} else {
					if connInterface, ok := hall.GameServers.Load(serverRemote); ok {
						conn = connInterface.(*net.TCPConn)
						break
					}
					core.Logger.Warn("[handleRacePlan]游戏服未连接, 6秒后重新选服, remote:%v", serverRemote)
				}
				time.Sleep(6 * time.Second)
			}
			core.Logger.Debug("select server success, conn:%v", conn.RemoteAddr().String())

			raceInfo.PlanRoomsMux.Lock()
			core.Logger.Debug("开始通知游戏服排赛, raceId:%v", raceId)
			for raceRoomId := range raceInfo.PlanRooms {
				raceRoom := raceRooms.Get(raceRoomId)
				if raceRoom != nil {
					raceRoom.ServerRemote = serverRemote
				}
				// notice game server open a game
				service.LeagueL2SPlanPush(raceRoomId).Send(conn)
				// _, err := conn.Write(service.LeagueL2SPlanPush(raceRoomId).Serialize())
				// core.Logger.Debug("通知游戏服拍赛, server remote:%v, raceRoomId:%v, err:%v", conn.RemoteAddr().String(), raceRoomId, err.Error())
				core.Logger.Debug("通知游戏服拍赛, server remote:%v, raceRoomId:%v, err:%v", conn.RemoteAddr().String(), raceRoomId, "")
			}
			raceInfo.PlanRoomsMux.Unlock()

			/*
				// 通知机器人进行重连
				// 应该移动到排赛完成
				if len(needPlanUserIds) > 0 {
					for _, userId := range needPlanUserIds {
						if robot.IsRobotUser(userId) {
							robotGameInfo := robot.NewGameInfo(serverRemote, 0, 0, 0, 1)
							robotGameInfo.RobotId = userId
							service.AddHallRobotRoom(serverRemote, robotGameInfo.String())
						}
					}
				}
			*/

			core.Logger.Debug("通知游戏服排赛完成, raceId:%v", raceId)

			select {
			case <-raceInfo.PlanWait:
				break
			case <-time.After(5 * time.Second):
				break
			}
		}
	}

	// 更新比赛开始时间
	raceInfo.Status = config.RACE_STATUS_PLAY
	o.Update(raceInfo, "status")

	core.Logger.Info("[handleRacePlan]completed, raceId:%v", raceId)
}

// 处理定时赛开赛前的踢人
func handleFixRaceKick(raceInfo *model.Race, leagueInfo *model.League) {
	core.Logger.Info("[handleFixRaceKick]距离开赛还有2小时, 检查并踢出还在房间中的用户")
	// 读取待排赛用户列表
	raceUsers := model.GetRaceUsers(raceInfo.Id)
	if raceUsers == nil {
		core.Logger.Error("[handleFixRaceKick]参与比赛的用户未找到, leagueId:%v, raceId:%v", leagueInfo.Id, raceInfo.Id)
		return
	}

	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	for _, ru := range raceUsers.Users {
		// 跳过机器人
		if robot.IsRobotUser(ru.UserId) {
			core.Logger.Debug("[handleFixRaceKick]跳过机器人用户, raceId:%v, userId:%v", raceInfo.Id, ru.UserId)
			continue
		}

		// 查找房间
		roomId := ss.GetUserRoomId(ru.UserId)
		if roomId == 0 {
			core.Logger.Debug("[handleFixRaceKick]用户不在房间中, 不踢出, raceId:%v, userId:%v", raceInfo.Id, ru.UserId)
			continue
		}

		// 将用户踢出房间
		// 删除报名信息
		delete(raceUsers.Users, ru.UserId)
		// 删除用户已报名的比赛id
		model.DelUserRace(ru.UserId)
		// 退费
		var chargeEntities map[int]int
		var chargeErr *ierror.Error
		if ru.Consume != nil && len(ru.Consume) > 0 {
			chargeEntities, chargeErr = service.ChargeReturnEntities(ru.UserId, ru.Consume, config.MONEY_CHANGE_TYPE_TF, config.MONEY_TRANS_TYPE_LEAGUE)
			if chargeErr != nil {
				core.Logger.Error("[handleFixRaceKick]退费失败,raceId:%v, userId:%v, consume:%v, err:%v", raceInfo.Id, ru.UserId, ru.Consume, chargeErr.Error())
			}
			core.Logger.Debug("[handleFixRaceKick]userId:%v, raceId:%v, consume:%+v", ru.UserId, raceInfo.Id, chargeEntities)
		} else {
			core.Logger.Debug("[handleFixRaceKick]未找到用户付费信息, userId:%v, raceId:%v", ru.UserId, raceInfo.Id)
		}
		leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)

		// 更新&广播比赛参与人数更新消息
		raceInfo.SignupUserCount--
		if leagueInfo.EnableSimulationUserCount() {
			leagueInfo.SimulationUserCount--
			hall.BrodcastMessage(service.LeagueRaceSignupCountPush(leagueInfo.Id, raceInfo.Id, leagueInfo.SimulationUserCount))
		} else {
			hall.BrodcastMessage(service.LeagueRaceSignupCountPush(leagueInfo.Id, raceInfo.Id, raceInfo.SignupUserCount))
		}

		// 推送比赛取消的消息给用户
		impacket := service.LeagueRaceCancelPush(raceInfo.Id)
		hall.PrivateMessage(ru.UserId, impacket)

		// 推送网关消息给用户
		messagePush := service.PrivateMessagePush(ru.UserId, fbsCommon.MessageIdLEAGUE_RACE_KICK, "由于你还在游戏中，已自动弃赛，距离开赛还有2分钟，想报名还来得及哦")
		gateway.SendPrivateMessage(messagePush.Serialize())

		// 记录推送消息
		data := make(map[string]interface{})
		data["user_id"] = ru.UserId
		data["race_id"] = raceInfo.Id
		s, _ := util.InterfaceToJsonString(data)
		core.RedisDo(core.RedisClient0, "rpush", config.CACHE_KEY_GIVEUP_MESSAGE_QUEUE, s)

		core.Logger.Info("[handleFixRaceKick]用户比赛开始前两分钟，还在房间内，将用户踢出房间, raceId:%v, userId:%v", raceInfo.Id, ru.UserId)
	}
}
