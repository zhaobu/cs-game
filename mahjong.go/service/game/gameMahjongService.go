package game

import (
	"fmt"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/chiken"
	"mahjong.go/mi/dice"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/protocal"
	"mahjong.go/mi/setting"
	"mahjong.go/mi/step"
	"mahjong.go/mi/suggest"
	"mahjong.go/mi/wall"
	roomService "mahjong.go/service/room"

	fbsCommon "mahjong.go/fbs/Common"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 类型定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
type Mahjong struct {
	// 锁
	Mux *sync.Mutex
	// 麻将类型
	MType int
	// 所属房间
	RoomId int64
	// 房间设置
	setting *setting.MSetting
	// 位置 => userId对照关系
	Index []int
	// 当前牌局数
	Round int
	// 总局数
	TRound int
	// 所有牌，在游戏开始时，根据游戏类型初始化
	// 游戏前要进行洗牌
	TileWall *wall.Wall
	// 骰子数字
	DiceList [2]int
	// 牌局用户数据
	Users *sync.Map
	// 牌局开始时间
	CreateTime int64

	Dealer      int // 庄家
	DealerCount int // 连庄数

	// 最后操作者
	LastOperator int
	// 最后进行的操作
	LastOperation *Operation
	// 最后打牌者
	LastPlayerId int

	// 当前进度
	Progress int

	// 定缺列表 index => tile
	LackList *sync.Map

	// 换牌列表 index => tiles
	ExchangeList *sync.Map

	// 鸡
	chiken *chiken.MChiken
	// 预变鸡状态
	preChangeChikenRock bool
	// 不包括1条、8筒外的鸡
	Chikens map[int]int

	// 胡牌信息
	HInfo HuInfo
	// 积分信息
	SInfo *UserScore
	// 结算数据
	FrontData *sync.Map

	// 用户回应操作队列
	WaitQueue *WaitMap
	// 操作回应产生时间, 用户回应时间-此时间，等于用户操作耗时
	replayInitTime int64
	// 首次操作标志
	firstOperateFlag bool

	Ob       *Ob       // 观察者
	playback *Playback // 回放

	// 选牌容器
	Selector *suggest.MSelector

	// 计分对照表
	scoreMap map[int]int

	// 结算组id
	settlementGroup uint16
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 工厂模式
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func MahjongFactory(room *Room) MahjongInterface {
	switch room.MType {
	case fbsCommon.GameTypeMAHJONG_GY: // 贵阳麻将
		return NewMahjongGY(room)
	case fbsCommon.GameTypeMAHJONG_BJ: // 毕节麻将
		return NewMahjongBJ(room)
	case fbsCommon.GameTypeMAHJONG_ZY: // 遵义麻将
		return NewMahjongZY(room)
	case fbsCommon.GameTypeMAHJONG_SD: // 三丁拐
		return NewMahjongSD(room)
	case fbsCommon.GameTypeMAHJONG_LD: // 两丁拐
		return NewMahjongLD(room)
	case fbsCommon.GameTypeMAHJONG_STT: // 72张
		return NewMahjongSTT(room)
	case fbsCommon.GameTypeMAHJONG_AS: // 安顺麻将
		return NewMahjongAS(room)
	case fbsCommon.GameTypeMAHJONG_XY: // 兴义麻将
		return NewMahjongXY(room)
	case fbsCommon.GameTypeMAHJONG_LPS: // 六盘水麻将
		return NewMahjongLPS(room)
	case fbsCommon.GameTypeMAHJONG_KL: // 凯里麻将
		return NewMahjongKL(room)
	case fbsCommon.GameTypeMAHJONG_DY: // 都匀麻将
		return NewMahjongDY(room)
	case fbsCommon.GameTypeMAHJONG_TR: // 铜仁麻将
		return NewMahjongTR(room)
	case fbsCommon.GameTypeMAHJONG_GYA: // 贵阳麻将（全鸡玩法）
		return NewMahjongGYA(room)
	case fbsCommon.GameTypeMAHJONG_QX: // 黔西麻将
		return NewMahjongQX(room)
	case fbsCommon.GameTypeMAHJONG_JS: // 金沙玩法
		return NewMahjongJS(room)
	case fbsCommon.GameTypeMAHJONG_ZYA: // 遵义麻将（全鸡玩法）
		return NewMahjongZYA(room)
	case fbsCommon.GameTypeMAHJONG_RH: // 仁怀麻将
		return NewMahjongRH(room)
	case fbsCommon.GameTypeMAHJONG_GFT: // 杠翻天玩法
		return NewMahjongGFT(room)
	case fbsCommon.GameTypeMAHJONG_GYE3: // 贵阳换3张
		return NewMahjongGYE3(room)
	case fbsCommon.GameTypeMAHJONG_GYE4: // 贵阳换4张
		return NewMahjongGFT(room)
	case fbsCommon.GameTypeExtra1: // 测试热更
		return NewMahjongGYA(room)
	case fbsCommon.GameTypeMAHJONG_MATCH_GZ_1: // 比赛：贵州麻将
		fallthrough
	case fbsCommon.GameTypeMAHJONG_MATCH_GZ_2: // 比赛：贵州麻将
		fallthrough
	case fbsCommon.GameTypeMAHJONG_MATCH_GZ_3: // 比赛：贵州麻将
		fallthrough
	case fbsCommon.GameTypeMAHJONG_MATCH_GZ_4: // 比赛：贵州麻将
		return NewMahjongMatchGZ(room)
	default:
		//return nil
		return NewMahjongGY(room)
	}
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 接口实现
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 获取用户缺的哪一门
func (this *Mahjong) getLackTile(id int) int {
	if v, ok := this.LackList.Load(this.getUser(id).Index); ok {
		return v.(int)
	}
	return 0
}

// 获取用户列表
func (this *Mahjong) getUsers() map[int]*MahjongUser {
	users := make(map[int]*MahjongUser)
	this.Users.Range(func(k, v interface{}) bool {
		users[k.(int)] = v.(*MahjongUser)
		return true
	})
	return users
}

// 获取用户信息
func (m *Mahjong) getUser(userId int) *MahjongUser {
	if mu, ok := m.Users.Load(userId); ok {
		return mu.(*MahjongUser)
	}
	return nil
}

// 获取用户个数
func (this *Mahjong) getUsersLen() int {
	return util.SMapLen(this.Users)
}

// 获取从前面抓了多少张
func (this *Mahjong) getForward() int {
	return this.TileWall.GetForward()
}

// 获取从后面多少张
func (this *Mahjong) getBackward() int {
	return this.TileWall.GetBackward()
}

// 获取牌局开始时间
func (this *Mahjong) getRoundCreateTime() int64 {
	return this.CreateTime
}

// 获取回应操作开始时间
func (this *Mahjong) getReplyInitTime() int64 {
	return this.replayInitTime
}

// 获取最后“打牌”或“报听”的用户id
func (this *Mahjong) getLastPlayerId() int {
	return this.LastPlayerId
}

// 获取最后操作
func (this *Mahjong) getLastOperation() *Operation {
	return this.LastOperation
}

// 获取最后操作者
func (this *Mahjong) getLastOperator() int {
	return this.LastOperator
}

// 翻牌
func (m *Mahjong) showHandTile() {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		// 给用户发送翻牌的UserOperationPush
		handTilesSlice := mu.sortHandTile(mu.WinTile)
		// 将用户最后操作的牌拿到最后
		m.SendUserOperationPush(NewUserOperation(mu.UserId, NewOperation(fbsCommon.OperationCodeSHOW_HAND_TILE, handTilesSlice)), 0)
		return true
	})
	core.Logger.Info("[showHandTile]roomId:%d, round:%v", m.RoomId, m.Round)
}

func (this *Mahjong) tileIsQue(userId, tile int) bool {
	if !this.setting.IsEnableLack() {
		return false
	}
	if card.IsSameSuit(this.getLackTile(userId), tile) {
		return true
	}
	return false
}

// 读取骰子数
func (this *Mahjong) getDiceList() [2]int {
	return this.DiceList
}

// 读取用户当局的输赢分数
func (this *Mahjong) getRoundScore(userId int) int {
	return this.SInfo.GetUserScore(userId).Total
}

// 获取庄家id、连庄数
func (this *Mahjong) getDealer() (int, int) {
	return this.Dealer, this.DealerCount
}

// 读取定缺列表
func (this *Mahjong) getLackList() map[int]int {
	lackList := make(map[int]int)
	this.LackList.Range(func(k, v interface{}) bool {
		lackList[k.(int)] = v.(int)
		return true
	})
	return lackList
}

// 获取牌墙剩余牌张数
func (m *Mahjong) getWallTileCount() int {
	return m.TileWall.RemainLength()
}

// 获取选牌器
func (m *Mahjong) getSelector() *suggest.MSelector {
	return m.Selector
}

// 是否处于换牌阶段
func (m *Mahjong) isExchanging() bool {
	return m.Progress == config.MAHJONG_PROGRESS_EXCHANGE
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 公用函数
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 开局初始化
func (this *Mahjong) begin(room *Room) {
	// 设置麻将类型
	this.MType = room.MType

	// 设置房间id
	this.RoomId = room.RoomId

	// 房间设置
	this.setting = room.setting

	// 初始化局数
	// 当前局数
	this.Round = room.Round
	// 总局数
	this.TRound = room.TRound

	// 初始化倍数
	// 是否支持最后一局翻倍
	if room.Round == room.TRound && room.setting.IsSettingDLR() {
		room.setting.MultipleRound = 2
	} else {
		room.setting.MultipleRound = 1
	}
	core.Logger.Debug("roomId:%v, number:%v, round:%v, tround:%v, IsSettingDLR:%v, MultipleRound:%v",
		room.RoomId, room.Number, room.Round, room.TRound, room.setting.IsSettingDLR(), room.setting.MultipleRound)

	// 初始化庄家
	this.Dealer = room.Dealer
	this.DealerCount = room.DealCount

	// 初始化牌墙
	this.TileWall = wall.NewWall()

	// 初始化选牌算法
	this.Selector = suggest.NewMSelector()

	// 开始时间
	this.CreateTime = util.GetTime()

	// 初始化用户
	this.Users = &sync.Map{}
	room.Users.Range(func(k, v interface{}) bool {
		userId := k.(int)
		ru := v.(*RoomUser)
		mu := newMahjongUser(userId, ru.Info.Version)
		if userId == this.Dealer {
			mu.IsDealer = true
		}
		// 加载用户的好牌率
		mu.DrawEffectExtraRate = roomService.GetDrawEffectExtraRate(mu.UserId, room.CType, room.GradeId, room.getLeagueId(), room.CoinType)

		// 加载用户的会员等级经验加成
		mu.MemberAddExp = ru.Info.MemberAddExp

		this.Users.Store(userId, mu)
		return true
	})

	// 初始化用户索引
	this.Index = make([]int, room.GetIndexLen(), room.GetIndexLen())
	room.Index.Range(func(k, v interface{}) bool {
		index := k.(int)
		userId := v.(int)
		this.Index[index] = userId
		this.getUser(userId).Index = index
		return true
	})

	// 初始化操作队列
	this.WaitQueue = NewWaitMap()

	// 初始化进度检查作弊
	this.Progress = config.MAHJONG_PROGRESS_INIT

	// 初始化定缺列表
	this.LackList = &sync.Map{}

	// 初始化换牌列表
	this.ExchangeList = &sync.Map{}

	// 初始化 ScoreInfo
	this.SInfo = &UserScore{Maps: &sync.Map{}}

	// 初始化锁
	this.Mux = &sync.Mutex{}

	// 初始化结算信息
	this.FrontData = &sync.Map{}

	// 设置预变鸡状态
	this.preChangeChikenRock = true

	// 初始化鸡
	this.chiken = chiken.NewMChiken()

	// 初始化所有鸡列表
	this.Chikens = make(map[int]int)

	// 初始化回放操作队列
	this.playback = NewPlayback(this.RoomId, this.Round)

	// 庄家第一次操作标志
	this.firstOperateFlag = true

	this.Ob = room.Ob

	// 初始化胡牌数据
	this.HInfo = HuInfo{[]int{}, []WinInfo{}, false, 0}
	for index := range this.getUsers() {
		this.SInfo.Maps.Store(index, &ScoreInfo{MahjongType: this.MType, Total: 0, Item: []ScoreItem{}})
	}

	core.Logger.Debug("mahjong.begin, roomId:%v, round:%v", this.RoomId, this.Round)
}

// 初始化游戏
func (this *Mahjong) initializtion() {
	// 初始化选牌器的牌
	this.Selector.SetTiles(this.TileWall.GetTiles())
	// 掷骰子
	this.dice()
	// 洗牌
	this.shuffle()
	// 分牌
	this.deal()
	// 为了保持后面的逻辑一致性
	// 初始化最后一次操作者是庄家
	dealLastDealTile := this.getUser(this.Dealer).HandTileList.GetLastAdd()
	this.setLastOperate(this.Dealer, NewOperation(fbsCommon.OperationCodeDRAW, []int{dealLastDealTile}))
	// 推送初始化数据给客户端
	for _, userId := range this.Index {
		var handTiles []int
		var rightTile int
		muser := this.getUser(userId)
		if this.LastOperator == userId {
			rightTile = this.LastOperation.Tiles[0]
			leftTiles := util.SliceDel(muser.HandTileList.ToSlice(), rightTile)
			handTiles = util.ShuffleSliceInt(leftTiles)
			handTiles = append(handTiles, rightTile)
		} else {
			handTiles = muser.HandTileList.ToSlice()
			handTiles = util.ShuffleSliceInt(handTiles)
		}
		core.Logger.Debug("GameInitPush,roomId:%v, round:%v, userId:%v,handTiles:%v", this.RoomId, this.Round, userId, handTiles)
		pushPacket := GameInitPush(this.Dealer, this.DealerCount, this.DiceList, handTiles, this.Round)
		SendMessageByUserId(userId, pushPacket)
		// 推送消息给观察员
		this.Ob.sendMessage(pushPacket, 0)
	}
	// 推送初始的倍数
	this.incrMultipleRound(0)
	// 因为客户端逻辑问题，防作弊移到这里
	if this.Round == 1 {
		room, _ := RoomMap.GetRoom(this.RoomId)

		// 兼容客户端逻辑，第一局init之后，再推送一次托管数据
		if len(room.HostingUsers) > 0 {
			for _, id := range room.HostingUsers {
				room.SendMessageToRoomUser(GameHostingPush(id, config.ROOM_USER_HOSTING_YES), 0)
			}
		}

		room.checkCheat()
	}

	// 刷新前后鸡
	this.setChikenRock()

	core.Logger.Info("[initializtioned]roomId:%v, round:%v", this.RoomId, this.Round)
}

// 补花
// 庄家补完下家补
func (m *Mahjong) flower() {
	// 庄家第一个
	dealUser := m.getUser(m.Dealer)
	m.flowerExchange(dealUser, true)

	index := dealUser.Index
	for i := 1; i < m.setting.GetSettingPlayerCnt(); i++ {
		index = m.calcNextIndex(index)
		userId := m.Index[index]
		m.flowerExchange(m.getUser(userId), true)
	}
}

// 换牌玩法
func (m *Mahjong) exchange() {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		// 计算推荐的牌
		suggestTiles := m.Selector.GetSuggestExchange(mu.HandTileList.ToSlice(), m.setting.GetExchangeCount())
		opList := []*Operation{NewOperation(fbsCommon.OperationCodeNEED_EXCHANGE_TILE, suggestTiles)}
		// 设置用户操作
		m.setWait(mu.UserId, NewWaitInfo(opList))
		// 推送消息
		mu.SendOperationPush(opList)
		core.Logger.Info("[exchange]roomId:%v, round:%v, userId:%v, suggestTiles:%v", m.RoomId, m.Round, mu.UserId, suggestTiles)
		return true
	})
}

func (this *Mahjong) initialTingMap() {
	// 初始化user TingMap, 庄家无法初始化
	for _, u := range this.getUsers() {
		if u.UserId != this.Dealer {
			this.initialUserTingMap(u)
		}
		core.Logger.Debug("[ting_init]roomId:%d, round:%d, userId:%d, ting map:%#v, ready:%d", this.RoomId, this.Round, u.UserId, u.MTC.GetMaps(), u.MTC.GetStatus())
	}
}

func (m *Mahjong) initialUserTingMap(mu *MahjongUser) {
	if mu.HandTileList.IsPlayStatus() {
		return
	}
	if mu.hasLackTile() {
		mu.MTC.SetNormal()
	} else if tiles := m.getTingSlice(mu); len(tiles) == 0 {
		mu.MTC.SetNormal()
	} else {
		mu.MTC.SetTingTiles(tiles)
	}
}

// 定缺
func (this *Mahjong) lack() {
	opList := []*Operation{NewOperation(fbsCommon.OperationCodeNEED_LACK_TILE, nil)}
	this.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		// 设置用户操作
		this.setWait(mu.UserId, NewWaitInfo(opList))
		// 推送消息
		mu.SendOperationPush(opList)
		return true
	})
}

// 获取下一个抓牌的人
func (m *Mahjong) getNextDrawer() int {
	userId := 0
	if m.firstOperateFlag == true ||
		oc.IsKongOperation(m.LastOperation.OperationCode) ||
		(m.LastOperation.OperationCode == fbsCommon.OperationCodeBAO_TING && m.getUser(m.LastOperator).HandTileList.GetDrawTileCnt() == 13) {
		userId = m.LastOperator
	} else {
		userId = m.Index[m.calcNextIndex(m.getUser(m.LastOperator).Index)]
	}
	return userId
}

// 抓牌
// specified 支持指定给某个用户从前面抓一张牌
// 如果未指定，则根据上次操作者，自动计算
func (this *Mahjong) draw(specified int) int {
	var u *MahjongUser // 抓牌的人
	var opCode int     // 操作码
	var tile int       // 抓到的牌
	if specified > 0 {
		u = this.getUser(specified)
		opCode = fbsCommon.OperationCodeDRAW
	} else if oc.IsKongOperation(this.LastOperation.OperationCode) {
		// 如果上一个操作是明杠、暗杠、转弯杠，需要给上一个操作者从后面发牌
		u = this.getUser(this.LastOperator)
		opCode = fbsCommon.OperationCodeDRAW_AFTER_KONG
	} else if this.LastOperation.OperationCode == fbsCommon.OperationCodeBAO_TING && this.getUser(this.LastOperator).HandTileList.GetDrawTileCnt() == 13 {
		// 如果是硬报听，需要给上一个操作者从前面发牌
		u = this.getUser(this.LastOperator)
		opCode = fbsCommon.OperationCodeDRAW
	} else {
		// 正常逻辑，给下一个用户从前面抓牌
		nextIndex := this.calcNextIndex(this.getUser(this.LastOperator).Index)
		u = this.getUser(this.Index[nextIndex])
		opCode = fbsCommon.OperationCodeDRAW
	}

	// 执行好牌率逻辑
	var handTiles []int
	var expectTiles []int
	if u.DrawEffectExtraRate > 0 {
		handTiles = u.HandTileList.ToSlice()
		expectTiles = this.getEffectTiles(u)
		core.Logger.Debug("[draw]roomId:%v, userId:%v, 用户手牌:%v, 一类有效牌:%v, 额外的好牌率:%v", this.RoomId, u.UserId, handTiles, expectTiles, u.DrawEffectExtraRate)
		if !core.IsProduct() {
			// 非生产环境，记录测试数据
			// 计算用户牌阶
			handStep := step.GetCardsStep(handTiles)
			// 牌墙剩余可期望的牌
			remainTiles := this.TileWall.GetExpectTiles()
			// 手牌张数
			handTilesCnt := len(handTiles)
			// 牌墙剩余张数
			remainTilesCnt := len(remainTiles)
			// 有效牌
			uniqueExpectTiles := util.SliceUniqueInt(expectTiles)
			// 有效牌种类数
			uniqueExpectTilesCnt := len(uniqueExpectTiles)
			// 牌墙中的有效牌张数
			var effectsCnt, totalCnt, rate int
			if len(expectTiles) > 0 {
				effectsCnt, totalCnt = this.TileWall.StatExpect(expectTiles)
				// 有效牌比例
				if totalCnt > 0 {
					rate = effectsCnt * 100 / totalCnt
				}
			}

			v := make(map[string]interface{}, 0)
			v["hand_step"] = handStep
			v["remain_tiles"] = remainTiles
			v["hand_tiles_cnt"] = handTilesCnt
			v["remain_tiles_cnt"] = remainTilesCnt
			v["unique_expect_tiles"] = uniqueExpectTiles
			v["unique_expect_tiles_cnt"] = uniqueExpectTilesCnt
			v["effects_cnt"] = effectsCnt
			v["rate"] = rate
			jsonString, _ := util.InterfaceToJsonString(v)
			core.RedisDo(core.RedisClient0, "LPUSH", "EXPECT_LOG", jsonString)
		}
		if len(expectTiles) > 0 {
			// 计算是否get到好牌
			if effectsCnt, totalCnt := this.TileWall.StatExpect(expectTiles); effectsCnt > 0 {
				// 计算好牌率
				rate := effectsCnt*100/totalCnt + u.DrawEffectExtraRate
				core.Logger.Debug("[draw]roomId:%v, userId:%v, 剩余有效牌数:%v, 总牌数:%v, 总的好牌率:%v", this.RoomId, u.UserId, effectsCnt, totalCnt, rate)
				if util.RandIntn(100)+1 < rate {
					if opCode == fbsCommon.OperationCodeDRAW_AFTER_KONG {
						// 设置可期待的牌
						this.TileWall.BackwardExpect(expectTiles)
					} else {
						// 设置可期待的牌
						this.TileWall.ForwardExpect(expectTiles)
					}
					core.Logger.Debug("[draw]获得好牌资格，roomId:%v, userId:%v, 设置抓牌范围:%v", this.RoomId, u.UserId, expectTiles)
				} else {
					if opCode == fbsCommon.OperationCodeDRAW_AFTER_KONG {
						// 设置不可期待的牌
						this.TileWall.BackwardUnexpect(expectTiles)
					} else {
						// 设置不可期待的牌
						this.TileWall.ForwardUnexpect(expectTiles)
					}
					core.Logger.Debug("[draw]获得好牌资格，roomId:%v, userId:%v, 设置不可抓牌范围:%v", this.RoomId, u.UserId, expectTiles)
				}
			}
		}
	}

	// 抓牌
	if opCode == fbsCommon.OperationCodeDRAW_AFTER_KONG {
		// 从后面抓一张牌
		tile = this.TileWall.BackwardDraw()
	} else {
		// 从前抓一张牌
		tile = this.TileWall.ForwardDraw()
	}

	core.Logger.Debug("[draw]用户手牌:%v, 一类有效牌:%v, 抓牌:%v", handTiles, expectTiles, tile)

	// 给用户增加手牌
	u.HandTileList.AddTile(tile, 1)
	// 更新用户过胡、过碰状态
	u.SkipWin = []int{}
	u.SkipPong = []int{}
	this.finishOperation(u.UserId, opCode, []int{tile})
	core.Logger.Info("[draw]roomId:%v, round:%v, userId:%v, opCode:%v, tile:%v", this.RoomId, this.Round, u.UserId, opCode, tile)
	return tile
}

// 掷骰子
func (m *Mahjong) dice() {
	m.DiceList = dice.GenerateDiceList()
	core.Logger.Info("[dice]roomId:%d, round:%d, diceList:%#v", m.RoomId, m.Round, m.DiceList)
}

// 分配手牌
func (this *Mahjong) deal() {
	// 给发牌的用户排个序，优先给庄家发牌
	ids := make([]int, 0, len(this.Index))
	ids = append(ids, this.Dealer)
	for _, v := range this.Index {
		if v != this.Dealer {
			ids = append(ids, v)
		}
	}
	for _, v := range ids {
		var initTileCnt int // 初始张数
		mu := this.getUser(v)
		if mu.IsDealer {
			// 庄家发14张牌
			initTileCnt = config.MAHJONG_INIT_TILE_CNT_DEALER
		} else {
			// 非庄家发13张牌
			initTileCnt = config.MAHJONG_INIT_TILE_CNT_NONDEALER
		}
		initTiles := this.TileWall.ForwardDrawMulti(initTileCnt)
		mu.HandTileList.InitTiles(initTiles)
		// 检查原缺状态
		mu.InitLack = mu.checkInitLack()
		core.Logger.Info("[deal]roomId:%v, round:%v, userId:%d,len:%d,tiles:%v,initLack:%v", this.RoomId, this.Round, v, initTileCnt, initTiles, mu.InitLack)
	}
}

// 砌牌
func (m *Mahjong) shuffle() {
	// 洗牌
	m.TileWall.Shuffle()
	// 载入配牌内容
	m.loadTestTiles()
	core.Logger.Info("[shuffle]roomId:%v, round:%v, tiles:%v", m.RoomId, m.Round, m.TileWall.GetTiles())
}

// 计算下一个索引
func (this *Mahjong) calcNextIndex(index int) int {
	if index >= this.getUsersLen()-1 {
		return 0
	}
	return index + 1
}

// 计算下一把的庄家
// 情况一：庄家胡牌了，庄家继续坐庄；
// 情况二：庄家未胡牌，一炮多响的情况下按逆时针方向判定先胡的接庄；
// 情况三：黄牌时庄家连庄；
func (this *Mahjong) calcNextDealer() int {
	core.Logger.Debug("calcNextDealer, roomId:%v, round:%v, HInfo:%#v", this.RoomId, this.Round, this.HInfo)
	// 如果黄牌，则庄家连庄
	if !this.HInfo.Hu || len(this.HInfo.WInfo) == 0 {
		return this.Dealer
	}

	// 读取所有的胡牌者
	allHuUsers := make([]int, 0, len(this.HInfo.WInfo))
	for _, wInfo := range this.HInfo.WInfo {
		allHuUsers = append(allHuUsers, wInfo.Id)
	}

	// 庄家胡牌了，庄家继续坐庄；
	if util.IntInSlice(this.Dealer, allHuUsers) {
		return this.Dealer
	}

	// 庄家未胡牌，一炮多响的情况下按逆时针方向判定先胡的接庄；
	nextIndex := this.getUser(this.Dealer).Index
	for i := 0; i < 3; i++ {
		nextIndex = this.calcNextIndex(nextIndex)
		if util.IntInSlice(this.Index[nextIndex], allHuUsers) {
			return this.Index[nextIndex]
		}
	}

	core.Logger.Warn("未按照规则找到下把的庄家，让庄家连任，roomId:%d, round:%v, HInfo:%#v", this.RoomId, this.Round, this.HInfo)

	return this.Dealer
}

// SendMessageToUser 给麻将用户发送消息
// 如果excludeUserId设置成大于0, 表示跳过excludeUserId这个用户
func (this *Mahjong) SendMessageToUser(imPacket *protocal.ImPacket, excludeUserId int) {
	for _, userId := range this.Index {
		if userId != excludeUserId {
			SendMessageByUserId(userId, imPacket)
		}
	}
}

// SendOperationPush 给房间用户推送operationpush
func (m *Mahjong) SendOperationPush(opList []*Operation) {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		mu.SendOperationPush(opList)
		return true
	})
}

// SendUserOperationPush 给牌局中的用户发送用户操作的消息
func (m *Mahjong) SendUserOperationPush(operation *UserOperation, excludeUserId int) {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		if excludeUserId != mu.UserId {
			mu.SendUserOperationPush(operation)
		}
		return true
	})
}

// SendClientOperationPush 给牌局中的用户发送系统操作的消息
func (m *Mahjong) SendClientOperationPush(operation *ClientOperation) {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		mu.SendClientOperationPush(operation)
		return true
	})
}

// 判断牌局是否结束
func (this *Mahjong) checkFinish() bool {
	return this.TileWall.IsAllDrawn()
}

// 获取除"id"外的其他成员id
func (m *Mahjong) getOtherUserId(id int) []int {
	return util.SliceDel(m.Index, id)
}

// 载入自定义牌墙
// 需要配置支持（线上不支持）
// 需要在配置文件夹下有tiles-userId.txt这个文件
func (m *Mahjong) loadTestTiles() {
	// 是否支持自定义
	if core.AppConfig.EnableDefineTiles != 1 {
		return
	}
	tiles, err := util.GetIntSliceFromFile(fmt.Sprintf("conf/%s/tiles-%v.txt", core.AppConfig.Env, m.Dealer), ",")
	if err == nil && len(tiles) == m.TileWall.Length() {
		m.TileWall.SetTiles(tiles)
	}
}

func (m *Mahjong) getPlaybackOperationList() []*playbackOperation {
	return m.playback.operationList
}

// 获取已定缺的用户列表
func (m *Mahjong) getLackedUsers() []int {
	lackedUsers := make([]int, 0)
	m.Users.Range(func(k, v interface{}) bool {
		if v.(*MahjongUser).LackTile > 0 {
			lackedUsers = append(lackedUsers, k.(int))
		}
		return true
	})
	return lackedUsers
}

// 判断胡牌时，是否是首圈
func (m *Mahjong) isFirstCycle() bool {
	var isFirst = true
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		if mu.DiscardTileList.GetPlayedLen() > 1 {
			isFirst = false
			return false
		}
		return true
	})
	return isFirst
}

// 获取某张牌的总出牌张数
func (m *Mahjong) getTilePlayedCnt(tile int) int {
	var cnt = 0
	m.Users.Range(func(k, v interface{}) bool {
		cnt += v.(*MahjongUser).DiscardTileList.GetTilePlayedCnt(tile)
		return true
	})
	return cnt
}
