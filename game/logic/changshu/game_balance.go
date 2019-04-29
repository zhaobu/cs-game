package main

//game_balance文件写算分方法
import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
)

type HuScoreInfo struct {
	mj.HuTypeList //胡牌类型
}

type (
	StartDiceType uint8 //开局色子情况
)

//定时器ID
const (
	StartDice_None StartDiceType = iota //不加倍
	StartDice_One                       //本局加倍
	StartDice_Two                       //本局和下局加倍
)

//结算信息
type GameBalance struct {
	game_config  *pbgame_logic.CreateArg //游戏参数
	startDice    StartDiceType           //开局色子
	huangzhuang  bool                    //是否荒庄
	baozi        int32                   //本局豹子倍数
	gameIndex    int32                   //第几局
	lastHuChair  int32                   //最后一个胡牌玩家
	bankerId     int32                   //庄家
	loseChair    int32                   //丢分玩家
	gangHuaChair int32                   //杠上花玩家
	huCard       int32                   //胡的牌
	huMode       mj.EmHuMode             //胡牌方式
	huChairs     map[int32]*HuScoreInfo  //胡牌玩家信息
	gangTou      map[int32]int32         //玩家杠头数
	pghuaShu     []int32                 //碰花,杠花
}

func (self *GameBalance) Init(config *pbgame_logic.CreateArg) {
	self.game_config = config
}

func (self *GameBalance) Reset() {
	//依据上一局结果判断是否豹子翻倍
	if self.huangzhuang || self.startDice == StartDice_Two {
		self.baozi = 2
	} else {
		self.baozi = 1
	}
	self.startDice = StartDice_None
	self.huangzhuang = false
	self.lastHuChair = -1
	self.bankerId = -1
	self.loseChair = -1
	self.gangHuaChair = -1
	self.huCard = 0
	self.huMode = mj.HuMode_None
	self.huChairs = make(map[int32]*HuScoreInfo, self.game_config.PlayerCount)
	self.gangTou = make(map[int32]int32, self.game_config.PlayerCount)
	self.pghuaShu = make([]int32, self.game_config.PlayerCount)
}

//统计次数
func (self *GameBalance) AddScoreTimes(balanceResult *mj.PlayerBalanceResult, op mj.EmScoreTimes) {
	if num, ok := balanceResult.ScoreTimes[op]; ok {
		balanceResult.ScoreTimes[op] = num + 1
	} else {
		balanceResult.ScoreTimes[op] = 1
	}
}

//处理豹子翻倍
func (self *GameBalance) DealStartDice(randRes [2]int32) {
	if randRes[0] == randRes[1] {
		self.baozi = 2
		self.startDice = StartDice_One
		if randRes[0] == 1 || randRes[0] == 4 {
			self.startDice = StartDice_Two
		}
	}
}

//计算杠头数
func (self *GameBalance) CalGangTou(leftCards []int32, bankerId int32) { // 杠头  1 扳4个 2 扳8个 3 独龙杠
	// var huaShu int32 = 5
	// if self.game_config.Barhead == 3 {
	// 	if mj.GetCardColor(leftCards[len(leftCards)-1]) < 4 {
	// 		huaShu = mj.GetCardValue(int32(leftCards[len(leftCards)-1]))
	// 	}
	// } else {
	// 	var num int32 = 4
	// 	if self.game_config.Barhead == 2 {
	// 		num = 8
	// 	}

	// 	index := bankerId //从庄家开始算起数杠头
	// 	for _, v := range leftCards {

	// 	}
	// }
}
