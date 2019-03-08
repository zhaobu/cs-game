package service

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/robot"
	"math/rand"
	"time"

	"github.com/fwhappy/util"
)

// 单局结束
func gameFinish(raceId, raceRoomId int64, scores []int) {
	// 比赛房间列表
	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms == nil {
		core.Logger.Error("[GameFinish]raceRooms==nil, raceId:%v", raceId)
		return
	}
	// 更新房间最后修改时间
	raceRoom := raceRooms.Get(raceRoomId)
	if raceRoom == nil {
		core.Logger.Error("[GameFinish]raceRoom==nil, raceId:%v, raceRoomId:%v", raceId, raceRoomId)
		return
	}
	// 判断房间是否已结束
	if raceRoom.Status != config.RACE_ROOM_STATUS_NORMAL {
		core.Logger.Error("[GameFinish]房间已经结束过了, raceId:%v, raceRoomId:%v", raceId, raceRoomId)
		return
	}
	raceRoom.UpdateTime = util.GetTime()

	// 更新内存中用户积分,redis和db中的在game服务器更新，这里无需重复更新
	for i := 0; i < len(scores); i += 2 {
		raceUser := model.GetRaceUserInfo(raceId, scores[i])
		if raceUser == nil {
			core.Logger.Warn("[GameFinish]raceUser==nil, raceId:%v, raceRoomId:%v, userId:%v", raceId, raceRoomId, scores[i])
			continue
		}
		// 更新用户分数
		raceUser.Score += scores[i+1]
		core.Logger.Debug("[GameFinish]用户积分变化, raceId:%v, raceRoomId:%v, userId:%v, score:%v", raceId, raceRoomId, scores[i], scores[i+1])
	}
	core.Logger.Info("[GameFinish]parse request data, raceRoomId:%v, scores:%v", raceRoomId, scores)

	// 读取剩余房间数
	playingRoomCount := raceRooms.PlayingCount()

	// 广播排名变化到所有游戏服
	hall.BrodcastMessageToGameServers(LeagueL2SRankRefreshPush(raceId))

	// 广播排名变化到联赛服
	refreshHallRank(raceId, playingRoomCount)

	core.Logger.Info("[GameFinish]raceId:%v, raceRoomId:%v, scores:%+v", raceId, raceRoomId, scores)
}

// 房间完成
func roomFinish(raceId, raceRoomId int64, code int) {
	// 比赛房间列表
	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms == nil {
		core.Logger.Error("[RoomFinish]raceRooms==nil, raceId:%v", raceId)
		return
	}
	raceRoom := raceRooms.Get(raceRoomId)
	if raceRooms == nil {
		core.Logger.Error("[RoomFinish]raceRooms==nil, raceId:%v", raceId)
		return
	}

	// 判断房间是否已结束
	if raceRoom.Status != config.RACE_ROOM_STATUS_NORMAL {
		core.Logger.Error("[RoomFinish]房间已经结束过了, raceId:%v, raceRoomId:%v", raceId, raceRoomId)
		return
	}

	if code == config.DISMISS_ROOM_CODE_FINISH {
		raceRoom.Status = config.RACE_ROOM_STATUS_FINISH
	} else {
		raceRoom.Status = config.RACE_ROOM_STATUS_DISMISS
	}

	// 比赛信息
	raceInfo := model.RaceList.Get(raceId)
	if raceInfo == nil {
		core.Logger.Error("[RoomFinish]raceInfo==nil, raceId:%v", raceId)
		return
	}

	// 联赛信息
	leagueInfo := model.LeagueList.Get(raceInfo.LeagueId)
	if leagueInfo == nil {
		core.Logger.Error("[RoomFinish]leagueInfo==nil, leagueId:%v", raceInfo.LeagueId)
		return
	}
	// 联赛奖励
	leagueRewards := model.LeagueRewardsList.Get(raceInfo.LeagueId)

	// 清除用户的当前房间
	for _, userId := range raceRoom.GetUsers() {
		if raceUserInfo := model.GetRaceUserInfo(raceId, userId); raceUserInfo != nil {
			// 如果非最后一轮，给机器人增加氛围分数
			baseScore := raceUserInfo.GetRobotFWScore(raceUserInfo.Round)
			from := raceUserInfo.Score
			if baseScore > 0 && raceUserInfo.Score < baseScore {
				raceUserInfo.Score = baseScore + util.RandIntn(11)
				model.UpdateRaceUserScoreAndRank(raceId, userId, (raceUserInfo.Score - from))
				core.Logger.Debug("[roomFinish]增加机器人的氛围积分, raceId:%v, userId:%v, from:%v, to:%v", raceId, userId, from, raceUserInfo.Score)
			}
			raceUserInfo.RoomId = 0
		}
	}

	if raceRooms.IsCompleted() {
		// 本轮比赛结束
		core.Logger.Info("[RoomFinish]本轮比赛已完全结束，等候结算, raceId:%v, round:%v", raceId, raceRoom.Round)

		// 将比赛状态更新为结算中
		raceInfo.Status = config.RACE_STATUS_SETTLEMENT
		core.GetWriter().Update(raceInfo, "status")

		// 计算比赛是否已经完全结束
		if raceInfo.IsCompleted() {
			raceFinish(raceInfo, leagueInfo, leagueRewards)
		} else {
			go raceRoundFinish(raceInfo, leagueInfo, leagueRewards)
		}
	} else {
		waitOtherRoomFinish(raceRoom)
	}

	// 广播排名变化到所有游戏服务器
	hall.BrodcastMessageToGameServers(LeagueL2SRankRefreshPush(raceId))
	// 广播排名变化到联赛服
	refreshHallRank(raceId, raceRooms.PlayingCount())
}

// 等待其他房间完成
func waitOtherRoomFinish(raceRoom *model.RaceRoom) {
	// 标记用户为"允许退赛"状态
	raceUsers := model.GetRaceUsers(raceRoom.RaceId)
	for _, userId := range raceRoom.GetUsers() {
		if raceUserInfo := raceUsers.Get(userId); raceUserInfo != nil {
			raceUserInfo.GiveupStatus = config.RACE_USER_GIVEUP_STATUS_ALLOW
			core.Logger.Debug("[RoomFinish]更新用户状态为允许退赛, raceId:%v, userId:%v, round:%v", raceUserInfo.Round, userId, raceUserInfo.Round)
		}
	}
	core.Logger.Info("[RoomFinish][waitOtherRoomFinish]等候其他房间完成, raceId:%v, raceRoomId:%v, status:%v", raceRoom.RaceId, raceRoom.Id, raceRoom.Status)
}

// 比赛结束
// 计算排名时，有个前几名不发奖励的算法
// 如果前几名中，有正常用户，则从后面，拉机器人到前面，直至把用户挤出去
// 如果后面没有机器人，则随机生成一个机器人，占位，同时将多个用户的排名，都弄到最后一名
func raceFinish(raceInfo *model.Race, leagueInfo *model.League, leagueRewards []*model.LeagueReward) {
	core.Logger.Info("[RaceFinish]比赛已经结束，等候生成排名、发放奖励, raceId:%v", raceInfo.Id)
	// 比赛已结束，更新比赛状态
	raceInfo.Status = config.RACE_STATUS_FINISH
	core.GetWriter().Update(raceInfo, "status")
	// 当前用户排名
	ranks := model.GetRaceUserRanks(raceInfo.Id)
	// 用户积分, 待排名的所有用户
	rankedScores := make([]map[string]int, 0)
	// 已被淘汰的机器人列表
	lostRobots := make([]int, 0)
	for i := 0; i < len(ranks); i += 2 {
		raceUserInfo := model.GetRaceUserInfo(raceInfo.Id, ranks[i])
		if raceUserInfo == nil {
			core.Logger.Warn("[RaceFinish]raceUserInfo==nil, raceId:%v, userId:%v, rank:%v", raceInfo.Id, ranks[i], ranks[i+1])
			continue
		}
		if raceUserInfo.Status == config.RACE_USER_STATUS_SIGNUP {
			// 清理用户游戏数据
			// 删除用户已报名的比赛id
			model.DelUserRaceSpecied(raceUserInfo.UserId, raceInfo.Id)
			// 删除机器人的占用
			robot.Remove(raceUserInfo.UserId)
			rankedScores = append(rankedScores, map[string]int{"user_id": ranks[i], "score": ranks[i+1]})
		} else {
			// 找出已淘汰的机器人
			if robot.IsRobotUser(ranks[i]) {
				lostRobots = append(lostRobots, ranks[i])
			}
		}
	}
	core.Logger.Debug("[RaceFinish]待排名用户列表:%+v", rankedScores)
	core.Logger.Debug("[RaceFinish]已被淘汰的机器人列表:%+v", lostRobots)
	// 最后排名数，因为可能会插入机器人，导致需要排名的人数增多，那么多的人，都是最后一名
	maxRank := len(rankedScores)
	// 最大积分
	maxScore := rankedScores[0]["score"]
	core.Logger.Debug("[RaceFinish]最大排名:%v, 最大积分:%v", maxRank, maxScore)

	// 找出需要几个机器人
	insertRobotCnt := 0
	// 首个真实用户的索引
	firstUserIndex := 0
	for i := 0; i < leagueInfo.SkipRank; i++ {
		if !robot.IsRobotUser(rankedScores[i]["user_id"]) {
			insertRobotCnt = leagueInfo.SkipRank - i
			firstUserIndex = i
			break
		}
	}
	core.Logger.Debug("[RaceFinish]需要插入机器人数量:%v, 第一个真实用户排名:%v", insertRobotCnt, (firstUserIndex + 1))

	// 需要前置的机器人
	frontRanks := make([]map[string]int, 0)
	// 被拉升的机器人所在的key
	incrKeys := make([]int, 0)
	// 第一步，将没有淘汰的机器人排名往上拉
	if insertRobotCnt > 0 {
		for i := firstUserIndex + 1; i < maxRank; i++ {
			if robot.IsRobotUser(rankedScores[i]["user_id"]) {
				incrKeys = append(incrKeys, i)
				rankInfo := rankedScores[i]
				rankInfo["front"] = 1
				frontRanks = append(frontRanks, rankInfo)
				core.Logger.Debug("[RaceFinish]排名被提升的机器人, key:%+v, value:%+v", i, rankedScores[i])
				insertRobotCnt--
				if insertRobotCnt == 0 {
					break
				}
			}
		}
	}

	// 第二步，再不行的话，从被淘汰的人中，继续晚上加
	// 如果需要有跳过的排名，则需要找机器人补这几个排名
	if leagueInfo.SkipRank > 0 {
		i := 0
		for ; i < insertRobotCnt; i++ {
			// 如果机器人不够，不再补
			// 极端情况下，整个比赛的机器人数不够，会导致跳过排名的算法失效
			if i > len(lostRobots)-1 {
				break
			}
			frontRanks = append(frontRanks, map[string]int{"user_id": lostRobots[i], "score": maxScore, "front": 1})
			core.Logger.Debug("[RaceFinish]被前置的已被淘汰的机器人, value:%+v", map[string]int{"user_id": lostRobots[i], "score": maxScore})
		}
	}

	// 如果补齐不足，则随便扔几个机器人进去
	if insertRobotCnt > 0 {
		core.Logger.Warning("[RaceFinish]补足跳过排名需要的机器人数不足，缺:%v", insertRobotCnt)
		for ; insertRobotCnt > 0; insertRobotCnt-- {
			frontRanks = append(frontRanks, map[string]int{"user_id": robot.FetchFill(), "score": maxScore, "fill": 1})
		}
	}

	// 最终排名
	finalScores := make([]map[string]int, 0)
	if len(frontRanks) > 0 {
		finalScores = frontRanks
		for k, v := range rankedScores {
			if util.IntInSlice(k, incrKeys) {
				continue
			}
			finalScores = append(finalScores, v)
		}
	} else {
		finalScores = rankedScores
	}

	// 落地
	rank := 0
	rankList := make([]model.RaceRank, 0)
	for _, scoreInfo := range finalScores {
		if _, isFill := scoreInfo["fill"]; isFill {
			// 生成最终排名
			rank++
			if rank >= maxRank {
				rank = maxRank
			}

			raceRank := model.RaceRank{
				RaceId: raceInfo.Id,
				UserId: scoreInfo["user_id"],
				Rank:   rank,
				Round:  raceInfo.Round,
				Score:  scoreInfo["score"],
				Status: config.RACE_USER_STATUS_SIGNUP,
			}
			rankList = append(rankList, raceRank)
			core.Logger.Debug("[RaceFinish]用户最终排名, raceId:%v, userId:%v, rank:%v, score:%v", raceInfo.Id, raceRank.UserId, rank, raceRank.Score)
		} else {
			raceUserInfo := model.GetRaceUserInfo(raceInfo.Id, scoreInfo["user_id"])
			if raceUserInfo == nil {
				core.Logger.Warn("[RaceFinish]raceUserInfo==nil, raceId:%v, userId:%v, rank:%v", raceInfo.Id, scoreInfo["user_id"], scoreInfo["score"])
				continue
			}
			// 生成最终排名
			rank++
			if rank >= maxRank {
				rank = maxRank
			}

			// 更新用户信息
			raceUserInfo.Rank = rank
			raceUserInfo.Score = scoreInfo["score"]
			// 生成比赛奖励
			if raceUserInfo.Rank <= len(leagueRewards) {
				// 推送比赛结果给发奖用户
				if !robot.IsRobot(raceUserInfo.UserId) {
					impacket := LeagueRaceResultPush(raceInfo, leagueInfo, raceUserInfo.UserId, raceUserInfo.Rank, leagueRewards[rank-1].Content)
					hall.PrivateMessage(raceUserInfo.UserId, impacket)
					hall.SetLastRaceResult(raceUserInfo.UserId, impacket)
				}
			} else {
				// 推送比赛结果给淘汰用户
				if !robot.IsRobot(raceUserInfo.UserId) {
					impacket := LeagueRaceResultPush(raceInfo, leagueInfo, raceUserInfo.UserId, raceUserInfo.Rank, "")
					hall.PrivateMessage(raceUserInfo.UserId, impacket)
					hall.SetLastRaceResult(raceUserInfo.UserId, impacket)
				}
				raceUserInfo.Status = config.RACE_USER_STATUS_FAIL
				raceUserInfo.FailTime = util.GetTime()
			}

			raceRank := model.RaceRank{
				RaceId: raceInfo.Id,
				UserId: raceUserInfo.UserId,
				Rank:   raceUserInfo.Rank,
				Round:  raceUserInfo.Round,
				Score:  raceUserInfo.Score,
				Status: raceUserInfo.Status,
			}

			// 如果是氛围机器人，改变机器人的id，这样游戏服务器那边的push名称就变了
			if _, ok := scoreInfo["robot"]; ok {
				raceRank.UserId = leagueRoomRobotOffset(raceInfo.Id, raceUserInfo.UserId)
			}

			rankList = append(rankList, raceRank)
			core.Logger.Debug("[RaceFinish]用户最终排名, raceId:%v, userId:%v, rank:%v, score:%v", raceInfo.Id, raceUserInfo.UserId, rank, raceUserInfo.Score)
			// 更新用户的最终分数
			_, err := core.GetWriter().Update(raceUserInfo, "score", "status", "fail_time")
			ierror.MustNil(err)
		}

	}

	// 写入最终排名
	num, err := model.InsertRaceRanks(rankList)
	core.Logger.Info("[RaceFinish]比赛最终排名, raceId:%v, 预计插入:%v, 实际插入:%v, err:%v", raceInfo.Id, len(rankList), num, err)

	// 通知比赛奖励生成完成
	noticeInfoRaceRewardsCompleted(raceInfo.Id)
	core.Logger.Info("[RaceFinish]比赛奖励发放完成，流程结束, raceId:%v, leagueId:%v", raceInfo.Id, raceInfo.LeagueId)
}

// 单轮结束
func raceRoundFinish(raceInfo *model.Race, leagueInfo *model.League, leagueRewards []*model.LeagueReward) {
	defer util.RecoverPanic()
	time.Sleep(10 * time.Second)

	// 等候10秒，等待拍赛
	// 比赛还有下一轮，更新比赛状态为排赛中
	core.Logger.Info("[RaceRoundFinish]比赛还有下一轮，等候排赛, raceId:%v, nextRound:%v", raceInfo.Id, raceInfo.Round+1)
	raceInfo.Round++
	raceInfo.Status = config.RACE_STATUS_PLAN
	core.GetWriter().Update(raceInfo, "status", "round")
	core.Logger.Info("[RaceRoundFinish]更新比赛状态为排赛状态, raceId:%v, round:%v", raceInfo.Id, raceInfo.Round)

	// 读取最终排名
	ranks := model.GetRaceUserRanks(raceInfo.Id)
	// 被淘汰用户的排名
	rankList := make([]model.RaceRank, 0)
	// 根据用户排名、淘汰部分用户
	// 取晋级人数
	promotionCnt := leagueInfo.GetRoundRequireUserCount(raceInfo.Round - 1)
	// 晋级用户
	promotionUsers := make([]int, 0)
	// 被淘汰用户
	failUsers := make([]int, 0)
	rank := 0
	for i := 0; i < len(ranks); i += 2 {
		// 取比赛用户信息
		raceUserInfo := model.GetRaceUserInfo(raceInfo.Id, ranks[i])
		if raceUserInfo == nil {
			core.Logger.Warn("[RaceRoundFinish]raceUserInfo==nil, raceId:%v, userId:%v, rank:%v", raceInfo.Id, ranks[i], ranks[i+1])
			continue
		}
		// 只有游戏中的用户，才进行排名
		if raceUserInfo.Status == config.RACE_USER_STATUS_SIGNUP {
			rank++
			raceUserInfo.Rank = rank
			raceUserInfo.Score = ranks[i+1]
			if rank <= promotionCnt {
				// 晋级
				promotionUsers = append(promotionUsers, raceUserInfo.UserId)
				raceUserInfo.Round++
			} else {
				// 淘汰
				// 删除用户已报名的比赛id
				model.DelUserRaceSpecied(raceUserInfo.UserId, raceInfo.Id)
				// 被淘汰的机器人，删除机器人的占用
				robot.Remove(raceUserInfo.UserId)

				// 更新用户比赛状态
				failUsers = append(failUsers, raceUserInfo.UserId)
				raceUserInfo.Status = config.RACE_USER_STATUS_FAIL
				raceUserInfo.FailTime = util.GetTime()

				// 计算被淘汰用户的排名
				raceRank := model.RaceRank{
					RaceId: raceInfo.Id,
					UserId: raceUserInfo.UserId,
					Rank:   raceUserInfo.Rank,
					Round:  raceUserInfo.Round,
					Score:  raceUserInfo.Score,
					Status: raceUserInfo.Status,
				}
				rankList = append(rankList, raceRank)
				core.Logger.Debug("[RaceRoundFinish]本轮被淘汰用户的排名, raceId:%v, userId:%v, rank:%v, score:%v",
					raceInfo.Id, raceUserInfo.UserId, rank, raceUserInfo.Score)

				// 发送被淘汰的消息
				var content string
				if rank <= len(leagueRewards) {
					content = leagueRewards[rank-1].Content
				}
				if !robot.IsRobot(raceUserInfo.UserId) {
					impacket := LeagueRaceResultPush(raceInfo, leagueInfo, raceUserInfo.UserId, raceUserInfo.Rank, content)
					hall.PrivateMessage(raceUserInfo.UserId, impacket)
					hall.SetLastRaceResult(raceUserInfo.UserId, impacket)
				}
			}
			core.GetWriter().Update(raceUserInfo, "round", "score", "status", "fail_time")
		} else {
			core.Logger.Debug("[RaceRoundFinish]用户已弃赛或者已淘汰, 不参与本轮的排名结算, raceId:%v, userId:%v, status:%v, user round:%v, race round",
				raceInfo.Id, raceUserInfo.UserId, raceUserInfo.Status, raceUserInfo.Round, raceInfo.Round)
		}
	}
	core.Logger.Info("[RaceRoundFinish][统计晋级结果]leagueId:%v, raceId:%v, 晋级用户列表:%v", raceInfo.LeagueId, raceInfo.Id, promotionUsers)
	core.Logger.Info("[RaceRoundFinish][统计晋级结果]leagueId:%v, raceId:%v, 淘汰用户列表:%v", raceInfo.LeagueId, raceInfo.Id, failUsers)

	// 将淘汰用户从比赛用户排名的集合中删除
	model.DelRaceUserRank(raceInfo.Id, failUsers)

	// 写入最终排名
	num, err := model.InsertRaceRanks(rankList)
	core.Logger.Info("[RaceRoundFinish]写入本轮被淘汰的用户的排名, raceId:%v, 预计插入:%v, 实际插入:%v, err:%v", raceInfo.Id, len(rankList), num, err)

	// 发送排名奖励
	if rank <= len(leagueRewards) {
		// 通知比赛奖励生成完成
		noticeInfoRaceRewardsCompleted(raceInfo.Id)
		core.Logger.Info("[RaceFinish]通知发放本轮淘汰用户的奖励, raceId:%v, leagueId:%v, max rank:%v", raceInfo.Id, raceInfo.LeagueId, rank)
	}

	// 重新放入排赛队列
	model.LeagueRacePlanChannel <- raceInfo.Id
	core.Logger.Info("[RaceRoundFinish]重新插入排赛队列,leagueId:%v, raceId:%v, round:%v", raceInfo.LeagueId, raceInfo.Id, raceInfo.Round)
}

// 监听房间活跃
func listenRoomActive(raceRoom *model.RaceRoom) {
	defer util.RecoverPanic()
	for {
		// 每分钟执行一次
		time.Sleep(time.Minute)

		if raceRoom.Status != config.RACE_ROOM_STATUS_NORMAL {
			core.Logger.Info("[listenRoomActive]房间已结束，退出监听, raceId:%v, raceRoomId:%v, roomId:%v", raceRoom.RaceId, raceRoom.Id, raceRoom.RoomId)
			break
		}

		// 房间已经不活跃了，需要结束房间
		if !raceRoom.IsActive() {
			core.Logger.Info("[listenRoomActive]监听到房间不活跃了，关闭房间, raceId:%v, raceRoomId:%v, roomId:%v", raceRoom.RaceId, raceRoom.Id, raceRoom.RoomId)
			roomFinish(raceRoom.RaceId, raceRoom.Id, 0)
			break
		}
	}
}

// 机器人头像偏移
func leagueRoomRobotOffset(raceId int64, userId int) int {
	r := rand.New(rand.NewSource(raceId))
	offset := r.Intn(100) + 1
	if userId >= 999800 {
		userId -= offset
	} else {
		userId += offset
	}
	return userId
}
