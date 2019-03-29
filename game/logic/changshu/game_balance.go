package main

//game_balance文件写算分方法
import (
	mj "cy/game/logic/changshu/majiang"
)

type HuScoreInfo struct {
	mj.HuTypeList                    //胡牌类型
	HuTypeExtra   []mj.EmExtraHuType //附属胡牌类型
}
type GameBalance struct {
	isdeuce   bool                   //是否流局
	gameIndex int32                  //第几局
	huChair   []int32                //胡牌的玩家
	banker    int32                  //庄家
	loseChair int32                  //丢分玩家
	huMode    mj.EmHuMode            //胡牌方式
	huCard    int32                  //胡的牌
	huChairs  map[int32]*HuScoreInfo //胡牌玩家信息
}

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
