package main

//game_balance文件写算分方法
import (
	mj "cy/game/logic/changshu/majiang"
)

type GameBalance struct {
	isdeuce   bool    //是否流局
	gameIndex int32   //第几局
	huChair   []int32 //胡牌的玩家
	banker    int32   //庄家

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
