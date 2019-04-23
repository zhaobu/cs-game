package main

//game_balance文件写算分方法
import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
)

type HuScoreInfo struct {
	mj.HuTypeList                    //胡牌类型
	HuTypeExtra   []mj.EmExtraHuType //附属胡牌类型
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
	lastHuChair  int32                   //最后一个胡牌玩家
	gameIndex    int32                   //第几局
	bankerId     int32                   //庄家
	loseChair    int32                   //丢分玩家
	huMode       mj.EmHuMode             //胡牌方式
	gangHuaChair int32                   //杠上花玩家
	gangPaoHu    bool                    //杠上炮
	huCard       int32                   //胡的牌
	huChairs     map[int32]*HuScoreInfo  //胡牌玩家信息
	gangTou      map[int32]int32         //玩家杠头数
}

func (self *GameBalance) Init(config *pbgame_logic.CreateArg) {
	self.game_config = config
}

func (self *GameBalance) Reset() {
	self.lastHuChair = -1
	self.bankerId = -1
	self.loseChair = -1
	self.gangHuaChair = -1
	self.huChairs = map[int32]*HuScoreInfo{}
	self.gangTou = map[int32]int32{}
}

//统计次数
func (self *GameBalance) AddScoreTimes(balanceResult *mj.PlayerBalanceResult, op mj.EmScoreTimes) {
	if num, ok := balanceResult.ScoreTimes[op]; ok {
		balanceResult.ScoreTimes[op] = num + 1
	} else {
		balanceResult.ScoreTimes[op] = 1
	}
}

//计算杠分
func (self *GameBalance) CalGangScore(chairId, loseChair int32, gangType mj.EmOperType) {

}

//处理豹子翻倍
func (self *GameBalance) DealStartDice(randRes [2]int32) {
	if randRes[0] == randRes[1] {
		self.startDice = StartDice_One
		if randRes[0] == 1 || randRes[0] == 4 {
			self.startDice = StartDice_Two
		}
	}
}

//计算杠头数
func (self *GameBalance) CalGangTou(leftCards []int32) {
	// if self.game_config.Barhead == 3 { //独龙杠

	// }
	// num := 0
	// for k, v := range leftCards {

	// }
}
