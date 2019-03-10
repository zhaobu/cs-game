package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
	obModel "mahjong.go/model/ob"
	configService "mahjong.go/service/config"
	logService "mahjong.go/service/log"
	userService "mahjong.go/service/user"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 流程相关的方法
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 开始执行
func (m *Mahjong) run() {
	switch m.Progress {
	case config.MAHJONG_PROGRESS_INIT: // 初始化
		m.doInitializtion()
	case config.MAHJONG_PROGRESS_FLOWER: // 补花
		m.doFlower()
	case config.MAHJONG_PROGRESS_EXCHANGE: // 换牌
		m.doExchange()
	case config.MAHJONG_PROGRESS_LACK: // 定缺
		m.doLack()
	case config.MAHJONG_PROGRESS_INIT_TING: // 初始化ting
		m.initTing()
	case config.MAHJONG_PROGRESS_PLAY: // 游戏
		m.doPlay()
	case config.MAHJONG_PROGRESS_SETTLE: // 单局结算
		m.doFinish()
	default:
	}
}

// 执行下一步
func (m *Mahjong) next() {
	m.Progress++
	m.run()
}

// 初始化
func (m *Mahjong) doInitializtion() {
	m.initializtion()
	m.next()
}

// 开局补花
func (m *Mahjong) doFlower() {
	m.flower()
	m.next()
}

// 换牌
func (m *Mahjong) doExchange() {
	if m.setting.IsEnableExchange() {
		m.exchange()
	} else {
		m.next()
	}
}

// 定缺
func (m *Mahjong) doLack() {
	if m.setting.IsEnableLack() {
		m.lack()
	} else {
		m.next()
	}
}

// 初始化ting信息
func (m *Mahjong) initTing() {
	m.initialTingMap()
	m.next()
}

// 游戏
// 处理上一把结束的逻辑
// 发牌
func (this *Mahjong) doPlay() {
	// 检查牌局是否结束
	if this.checkFinish() {
		this.next()
		return
	}
	// 先检测补花
	// 若有补花操作，中断流程
	hasFlower := this.doFlowerBefor()
	if hasFlower {
		return
	}
	if this.firstOperateFlag {
		// 庄家第一次操作
		this.doDealerFirstOperation()
	} else {
		this.doDraw()
	}
}

// 检查是否需要补花操作
func (m *Mahjong) doFlowerBefor() bool {
	// 处理抓牌前操作-补花
	nextDrawer := m.getNextDrawer()
	var opList = m.calcOperation(nextDrawer, config.MAHJONG_OPERATION_CALC_BEFORE_DRAW, m.LastOperation.Tiles[0])
	if len(opList) != 0 {
		m.setWait(nextDrawer, NewWaitInfo(opList))
		// 发送杠牌消息
		m.getUser(nextDrawer).SendOperationPush(opList)
		return true
	}
	return false
}

// 庄家第一次操作
func (m *Mahjong) doDealerFirstOperation() {
	// 计算庄家能进行什么操作
	var opList = m.calcOperation(m.LastOperator, config.MAHJONG_OPERATION_CALC_DROW, m.getUser(m.LastOperator).HandTileList.GetLastAdd())
	// 设置等候操作队列
	m.setWait(m.LastOperator, NewWaitInfo(opList))
	// 推送庄家的操作
	m.getUser(m.LastOperator).SendOperationPush(opList)
	// 更新首次操作标志
	m.firstOperateFlag = false
}

// 给用户发牌
func (m *Mahjong) doDraw() {
	// 检查并设置冲锋鸡
	m.setChikenCharge()
	// 如果是杠后抓牌，记录一下处于预变鸡状态
	if oc.IsKongOperation(m.LastOperation.OperationCode) {
		// 刚后抓牌前，判断是否需要加倍
		if m.setting.EnableKongDouble {
			m.incrMultipleRound(1)
		}
		m.preChangeChikenRock = true
	} else {
		// 抓牌前，需要判断是否需要变鸡
		m.setChikenRock()
	}
	// 如果是暗杠红中，且是闲家的话，需要再从前面抓一张
	extraDraw := false
	if m.LastOperation.OperationCode == fbsCommon.OperationCodeKONG_DARK &&
		card.IsFlower(m.LastOperation.Tiles[0]) && !m.getUser(m.LastOperator).IsDealer {
		extraDraw = true
	}

	// 给用户发牌
	tile := m.draw(0)
	// 发送抓牌的消息
	drawOpList := make([]*Operation, 0)
	// 追加抓牌的operationpush，用于客户端播放抓牌动画
	drawOpList = append(drawOpList, m.LastOperation)
	// 给抓牌者推送消息
	m.getUser(m.LastOperator).SendOperationPush(drawOpList)
	time.Sleep(time.Millisecond)

	if extraDraw {
		core.Logger.Debug("闲家初始杠了补花的牌，多抓一次牌,roomId:%v,round:%v,userId:%v", m.RoomId, m.Round, m.LastOperator)
		tile = m.draw(m.LastOperator)
		// 发送抓牌的消息
		extraDrawOpList := make([]*Operation, 0)
		// 追加抓牌的operationpush，用于客户端播放抓牌动画
		extraDrawOpList = append(extraDrawOpList, m.LastOperation)
		// 给抓牌者推送消息
		m.getUser(m.LastOperator).SendOperationPush(extraDrawOpList)
		time.Sleep(time.Millisecond)
	}

	// 如果抓到了花，则需要进行一次补花操作
	if tile == card.MAHJONG_RED_FLOWER {
		m.flowerExchange(m.getUser(m.LastOperator), false)
	}
	// 计算抓牌者能进行的操作
	tile = m.getUser(m.LastOperator).HandTileList.GetLastAdd()
	opList := m.calcOperation(m.LastOperator, config.MAHJONG_OPERATION_CALC_DROW, tile)

	if m.getUser(m.LastOperator).MTC.IsBaoTing() && len(opList) == 0 {
		m.doAutoPlay()
	} else {
		// 设置等候操作队列
		m.setWait(m.LastOperator, NewWaitInfo(opList))
		// 给抓牌者推送消息
		m.getUser(m.LastOperator).SendOperationPush(opList)
	}
}

// 自动帮用户出牌（报听的用户）
func (m *Mahjong) doAutoPlay() {
	go func() {
		defer util.RecoverPanic()
		// 如果用户已报听，需要帮用户打牌
		// 自动帮用户打牌,延迟3秒
		time.Sleep(1 * time.Second)
		playTile := m.getUser(m.LastOperator).HandTileList.GetLastAdd()
		m.userOperationPlay(m.LastOperator, playTile)

		core.Logger.Debug("baoting userOperationPlay, userId:%v, tile:%v", m.LastOperator, playTile)
		time.Sleep(1 * time.Second)

		// 计算其他人的操作, 如果其他人不能操作，则继续执行run
		if m.calcAfterUserOperation(config.MAHJONG_OPERATION_CALC_PLAY) == true {
			m.run()
		}
	}()
}

// 单局结算
func (this *Mahjong) doFinish() {
	// 是否最后一局
	lastRound := this.TRound > 0 && this.Round >= this.TRound
	// 清除最后一次打牌的人
	this.LastPlayerId = 0
	// 翻牌
	this.showHandTile()
	// 翻开翻牌鸡
	this.setChikenDraw()
	// 将本把累积的积分先记录下来，用于回放
	var room, _ = RoomMap.GetRoom(this.RoomId)
	gameScores := make(map[int]int, room.GetIndexLen())
	room.ScoreInfo.Range(func(k, v interface{}) bool {
		ffInfo := v.(*FrontFinalInfo)
		gameScores[ffInfo.UserId] = ffInfo.Score
		return true
	})
	// 算分
	this.clacScore()
	// 统计出本局的所有鸡（不包括1条、8筒）
	this.genFrontData()
	// 存储单局游戏数据
	room.Record = append(room.Record, room.MI)
	// 记录上一局的完成时间
	room.LastRoundCompletedTime = util.GetTime()
	// 开启新的协程去处理数据落地
	// 生成回放数据
	playbackData := genPlayback(room, gameScores)
	playbackDataIntact := genPlaybackIntact(room, gameScores)
	// 记录用户单局
	roundScores := this.recordGameRoundData()
	// 数据落地
	room.recordData()
	// 发送给所有人积分信息
	settlementPushPacket := GameSettlementPush(room)
	this.SendMessageToUser(settlementPushPacket, 0)
	// 给观察者发送积分信息
	this.Ob.sendMessage(settlementPushPacket, 0)
	// 可以延后写入的数据，放到另外的routine
	// 这里需要注意，另外的routine可能会因为延迟，导致room的数据与最初有偏差
	go func() {
		defer util.RecoverPanic()
		// 保存用户的最近大牌型
		this.saveLastWinStatus()

		// 生成并保存回放数据
		this.playback.save(playbackData, false)
		this.playback.save(playbackDataIntact, true)
		// 如果房间支持观察，则记录观察房间
		if room.IsTV() {
			obModel.RecordObRoom(room.RoomId)
		}
		if room.Round == 1 {
			// 存储用户最后游戏记录
			this.saveUserLastGame()
		}
	}()
	// 如果是俱乐部房间, 更新排名、推送房间结束
	if room.IsLeague() {
		leagueGameRoundFinish(room, roundScores)
	}
	// 判断是否已经完成了所有局
	if lastRound {
		room.finish()
	} else {
		// 计算并设置下一把的庄家
		room.SetNextDealer(this.calcNextDealer())
		// 设置房间为待准备状态
		room.SetNotReady()

		/*
			if room.EnableAutoReady() {
				go func() {
					// 局间有间隔
					time.Sleep(5 * time.Second)
					// 自动准备
					room.ReadyList = room.getUserIds()
					room.SetReady()
					core.Logger.Debug("[room.autoready]roomId:%v, number:%v, round:%v", room.RoomId, room.Number, room.Round+1)
					room.nextGame()
				}()
			}
		*/
	}
}

// 记录单局数据
func (this *Mahjong) recordGameRoundData() []int {
	// 日志记录
	data := make(map[string]interface{})
	userScores := make([]map[string]int, 0, len(this.Index))
	// 用户积分-slice版本
	userScoresSlice := make([]int, 0, len(userScores)*2)
	// 记录单局用户计算信息
	for userId, u := range this.getUsers() {
		userScore := this.SInfo.GetUserScore(userId).Total
		userScores = append(userScores, map[string]int{"userId": userId, "score": userScore})
		userScoresSlice = append(userScoresSlice, userId, userScore)
		// 记录当局用户的信息
		gameUserRound := new(config.GameUserRound)
		gameUserRound.RoomId = this.RoomId
		gameUserRound.Round = this.Round
		gameUserRound.UserId = userId
		// 听牌状态:-1未初始化;0未听牌;1:听牌;2:报听
		gameUserRound.TingStatus = u.MTC.GetStatus()
		// 胡牌类型:0未胡牌;1自模胡;2胡牌;3抢杠胡;4热炮胡;5杠后开花
		gameUserRound.WinStatus = u.WinStatus
		// 胡牌牌型
		gameUserRound.WinType = u.WinType
		// 点炮类型:0未点炮;1点炮;2:热炮;3:被抢杠
		gameUserRound.PaoStatus = u.PaoStatus
		// 冲锋鸡
		if this.chiken.GetChargeBam1() == userId {
			gameUserRound.ChikenChargeBam1 = 1
		}
		// 责任鸡
		if this.chiken.GetResponsibility() == userId {
			gameUserRound.ChikenResponsibility = 1
		}
		// 冲锋乌骨
		if this.chiken.GetChargeDot8() == userId {
			gameUserRound.ChikenChargeDot8 = 1
		}
		// 包鸡
		if u.KitchenStatus == config.USER_SETTLEMENT_CHIKEN_STATUS_LOSE {
			gameUserRound.ChikenBao = 1
		}
		// 所有的鸡个数、幺鸡个数、乌骨鸡个数
		var chikenCnt, chikenBam1Cnt, chikenDot8Cnt int
		fi, ok := this.FrontData.Load(userId)
		if ok {
			frontData := fi.(*FrontScoreInfo)
			for _, v := range frontData.HandChikens {
				if v.Tile == card.MAHJONG_BAM1 {
					chikenBam1Cnt++
				} else if v.Tile == card.MAHJONG_DOT8 {
					chikenDot8Cnt++
				}
				chikenCnt++
			}
			for _, v := range frontData.PlayChikens {
				if v.Tile == card.MAHJONG_BAM1 {
					chikenBam1Cnt++
				} else if v.Tile == card.MAHJONG_DOT8 {
					chikenDot8Cnt++
				}
				chikenCnt++
			}
			for _, v := range frontData.ShowCardChikens {
				if v.Tile == card.MAHJONG_BAM1 {
					chikenBam1Cnt++
				} else if v.Tile == card.MAHJONG_DOT8 {
					chikenDot8Cnt++
				}
				chikenCnt++
			}
		}

		// 明杠、转弯杠、暗杠个数
		var kongCnt, kongTurnCnt, kongDarkCnt int
		for _, showCard := range u.ShowCardList.GetAll() {
			switch showCard.GetOpCode() {
			case fbsCommon.OperationCodeKONG:
				kongCnt++
			case fbsCommon.OperationCodeKONG_TURN:
				fallthrough
			case fbsCommon.OperationCodeKONG_TURN_FREE:
				kongTurnCnt++
			case fbsCommon.OperationCodeKONG_DARK:
				kongDarkCnt++
			default:
				break
			}
		}

		gameUserRound.ChikenCnt = chikenCnt
		gameUserRound.ChikenBam1Cnt = chikenBam1Cnt
		gameUserRound.ChikenDot8Cnt = chikenDot8Cnt
		gameUserRound.KongCnt = kongCnt
		gameUserRound.KongTurnCnt = kongTurnCnt
		gameUserRound.KongDarkCnt = kongDarkCnt
		gameUserRound.PlayCnt = u.DiscardTileList.GetPlayedLen()
		gameUserRound.DrawCnt = u.HandTileList.GetDrawTileCnt()
		gameUserRound.ReplyTimeCnt = u.ReplyTimeCnt
		gameUserRound.Score = this.SInfo.GetUserScore(userId).Total
		gameUserRound.StartTime = this.CreateTime
		gameUserRound.CompleteTime = util.GetTime()
		logService.LogGameUserRound(gameUserRound)
	}
	// 记录单局数据
	// 是否黄牌
	huang := 1
	winPlayers := []int{}
	goldBam1 := 0
	goldDot8 := 0
	if this.HInfo.Hu == true {
		huang = 0
		// 统计胡牌者
		for _, v := range this.HInfo.WInfo {
			winPlayers = append(winPlayers, v.Id)
		}
	}
	if this.isGoldBam1() {
		goldBam1 = 1
	}
	if this.isGoldDot8() {
		goldDot8 = 1
	}
	logService.LogGameRoundData(this.RoomId, this.Round, userScores, data, huang, winPlayers, goldBam1, goldDot8, this.CreateTime)

	return userScoresSlice
}

// 保存用户最后游戏记录
func (m *Mahjong) saveUserLastGame() {
	userIds := []int{}
	for _, userId := range m.Index {
		if configService.IsRobot(userId) {
			continue
		}
		userIds = append(userIds, userId)
	}
	userService.SaveUserLastGame(userIds, m.RoomId)
}

// 保存回放-完整版，用于异常解散时
func (m *Mahjong) savePlaybackIntact() {
	var room, _ = RoomMap.GetRoom(m.RoomId)
	if room != nil {
		gameScores := make(map[int]int, room.GetIndexLen())
		room.ScoreInfo.Range(func(k, v interface{}) bool {
			ffInfo := v.(*FrontFinalInfo)
			gameScores[ffInfo.UserId] = ffInfo.Score
			return true
		})
		m.playback.save(genPlaybackIntact(room, gameScores), true)
	}
}

// 保存用户的最后大牌型
func (m *Mahjong) saveLastWinStatus() {
	m.FrontData.Range(func(k, v interface{}) bool {
		info := v.(*FrontScoreInfo)
		index, ok := config.WinStatusIndex[info.WinStatus]
		if ok {
			// 找出用户上次的排名
			lastWinStatus := getLastWinStatus(info.UserId)
			if index > config.WinStatusIndex[lastWinStatus] {
				mu := m.getUser(info.UserId)
				// 读取明牌
				handTiles := mu.HandTileList.ToSlice()
				showCards := make([]map[string]interface{}, 0)
				for _, card := range mu.ShowCardList.GetAll() {
					var d = map[string]interface{}{
						"tiles":  card.GetTiles(),
						"opCode": card.GetOpCode(),
					}
					showCards = append(showCards, d)
				}
				// 读取手牌
				var data = map[string]interface{}{
					"winStatus": info.WinStatus,
					"handTiles": handTiles,
					"showCards": showCards,
				}
				setLastWinStatus(info.UserId, info.WinStatus, data)
				core.Logger.Debug("[setLastWinStatus]userId:%v, lastWinStauts:%v, winStatus:%v", info.UserId, lastWinStatus, info.WinStatus)
			}
		}
		return true
	})
}

func getLastWinStatus(userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LAST_WIN_STATUS, userId)
	v, _ := core.RedisDoBytes(core.RedisClient0, "get", cacheKey)
	if len(v) == 0 {
		return 0
	}
	data := make(map[string]interface{})
	json.Unmarshal([]byte(v), &data)
	return int(data["winStatus"].(float64))
}

func setLastWinStatus(userId, index int, data map[string]interface{}) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LAST_WIN_STATUS, userId)
	// 计算cache有效期
	expire := util.GetChinaWeekStartTime() + int64(7*86400)
	v, _ := util.InterfaceToJsonString(data)
	core.RedisDo(core.RedisClient0, "set", cacheKey, v)
	core.RedisDo(core.RedisClient0, "expireat", cacheKey, expire)
}
