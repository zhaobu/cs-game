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

var hitCards []map[int32]bool //扳杠头

//开局色子类型
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
	pghuaShu     []int32                 //碰花,杠花
	duLongHua    int32                   //独龙杠花
	allCards     [][]int32               //扳的所有牌
	hitIndex     [][]int32               //扳到的杠头的索引
}

func (self *GameBalance) Init(config *pbgame_logic.CreateArg) {
	self.game_config = config
	hitCards = []map[int32]bool{
		0: map[int32]bool{ //庄家
			11: true,
			21: true,
			31: true,
			15: true,
			25: true,
			35: true,
			19: true,
			29: true,
			39: true,
			45: true,
			41: true,
			51: true,
		},
		1: map[int32]bool{ //下家
			14: true,
			24: true,
			34: true,
			18: true,
			28: true,
			38: true,
			44: true,
			54: true,
		},
		2: map[int32]bool{ //对家
			13: true,
			23: true,
			33: true,
			17: true,
			27: true,
			37: true,
			47: true,
			43: true,
			53: true,
		},
		3: map[int32]bool{ //上家
			12: true,
			22: true,
			32: true,
			16: true,
			26: true,
			36: true,
			46: true,
			42: true,
			52: true,
		},
	}
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
	self.pghuaShu = make([]int32, self.game_config.PlayerCount)
	self.duLongHua = 0
	self.allCards = make([][]int32, self.game_config.PlayerCount) //扳的所有牌
	self.hitIndex = make([][]int32, self.game_config.PlayerCount) //扳到的牌的索引
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
	if self.game_config.Barhead == 3 {
		self.duLongHua = 5
		if mj.GetCardColor(leftCards[len(leftCards)-1]) < 4 {
			self.duLongHua = mj.GetCardValue(int32(leftCards[len(leftCards)-1]))
		}
	} else {
		num := 4
		if self.game_config.Barhead == 2 {
			num = 8
		}

		getCanHit := func(chairId int32) map[int32]bool { //获取能中的牌
			for i, j := bankerId, int32(0); j < self.game_config.PlayerCount; i, j = mj.GetNextChair(i, self.game_config.PlayerCount), j+1 {
				if chairId == i {
					return hitCards[j]
				}
			}
			return nil
		}
		count := 0          //计数
		chairId := bankerId //从庄家开始算起数杠头
		for _, v := range leftCards {
			self.allCards[chairId] = append(self.allCards[chairId], v)
			if self.huChairs[chairId] != nil && getCanHit(chairId)[v] {
				self.hitIndex[chairId] = append(self.hitIndex[chairId], int32(len(self.allCards[chairId])-1))
				count++
				if count >= num {
					break
				}
			}
			chairId = mj.GetNextChair(chairId, self.game_config.PlayerCount)
		}
	}
}
