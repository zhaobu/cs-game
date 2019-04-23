package main

//game_balance文件写算分方法
import (
	mj "cy/game/logic/changshu/majiang"
)

type HuScoreInfo struct {
	mj.HuTypeList                    //胡牌类型
	HuTypeExtra   []mj.EmExtraHuType //附属胡牌类型
}

//结算信息
type GameBalance struct {
	lastHuChair  int32                  //最后一个胡牌玩家
	gameIndex    int32                  //第几局
	bankerId     int32                  //庄家
	loseChair    int32                  //丢分玩家
	huMode       mj.EmHuMode            //胡牌方式
	gangHuaChair int32                  //杠上花玩家
	gangPaoHu    bool                   //杠上炮
	huCard       byte                   //胡的牌
	huChairs     map[int32]*HuScoreInfo //胡牌玩家信息
}

func (self *GameBalance) Reset() {
	self.lastHuChair = -1
	self.bankerId = -1
	self.loseChair = -1
	self.gangHuaChair = -1
	self.huChairs = map[int32]*HuScoreInfo{}
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
