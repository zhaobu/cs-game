package majiang

import pbgame_logic "game/pb/game/mj/changshu"

type PlayerInfo struct {
	CardInfo      PlayerCardInfo      //玩家牌局信息
	BalanceInfo   PlayserBalanceInfo  //局结算
	BaseInfo      PlayerBaseInfo      //玩家个人信息
	BalanceResult PlayerBalanceResult //总结算
}

func MakePlayers() *PlayerInfo {
	p := &PlayerInfo{}
	p.init()
	return p
}
func (self *PlayerInfo) init() {
	self.BalanceResult.init()
	self.Reset()
}

func (self *PlayerInfo) Reset() {
	self.CardInfo.reset()
	self.BalanceInfo.reset()
}

// type OperRecord struct {
// 	OperType  pbgame_logic.OperType //操作类型
// 	Card      int32                 //操作的牌
// 	LoseChari int32                 //操作来源玩家
// }
type PlayerCardInfo struct {
	HandCards  []int32                         //玩家手牌
	OutCards   []int32                         //出过的牌
	PengCards  map[int32]int32                 // {card=chair_id,..}
	GangCards  map[int32]pbgame_logic.OperType // {card=G_OP_TYPE类型,...}
	StackCards map[int32]int32                 //玩家手牌数量统计 {card=num,...}
	ChiCards   [][3]int32                      //{card1,card2,card3,card4,card5,card6}3个连续的能组成吃,吃的牌放第一个
	RiverCards []*pbgame_logic.OperRecord      //操作记录
	HuaCards   []int32                         //花牌
	GuoPeng    map[int32]int32                 //过碰的牌
	CanNotOut  map[int32]int32                 //不能打的牌,包括吃后,碰后不能打的牌
}

func (self *PlayerCardInfo) reset() {
	self.HandCards = []int32{}
	self.OutCards = []int32{}
	self.PengCards = map[int32]int32{}
	self.GangCards = map[int32]pbgame_logic.OperType{}
	self.StackCards = map[int32]int32{}
	self.ChiCards = [][3]int32{}
	self.RiverCards = []*pbgame_logic.OperRecord{}
	self.HuaCards = []int32{}
	self.GuoPeng = map[int32]int32{}
	self.CanNotOut = map[int32]int32{}
}

//单局结算信息
type PlayserBalanceInfo struct {
	Point        int32           //得分
	HuCard       int32           //胡的牌
	Baozi        int32           //豹子倍数
	HuPoint      int32           //胡牌分
	BuHuaPoint   int32           //补花分
	GangPoint    int32           //杠分
	SpecialPoint int32           //特殊牌型分
	JiangMaPoint int32           //奖码分
	FengPoint    int32           //风花
	DiPiaoPoint  int32           //底飘分
	HuType       []EmHuScoreType //emHuScoreType类型参数
}

func (self *PlayserBalanceInfo) reset() {
	self.Point = 0
	self.HuCard = 0
	self.Baozi = 0
	self.HuPoint = 0
	self.BuHuaPoint = 0
	self.GangPoint = 0
	self.SpecialPoint = 0
	self.JiangMaPoint = 0
	self.FengPoint = 0
	self.DiPiaoPoint = 0
	self.HuType = []EmHuScoreType{}
}

//返回平胡需要的花数
func (self *PlayserBalanceInfo) GetPingHuHua() int32 {
	return self.BuHuaPoint + self.GangPoint + self.FengPoint
}

//玩家基础信息
type PlayerBaseInfo struct {
	ChairId  int32
	Uid      uint64
	Nickname string
}

//总结算
type PlayerBalanceResult struct {
	ScoreTimes map[EmScoreTimes]int32 //统计次数
	Point      int32                  //累计总得分
}

func (self *PlayerBalanceResult) init() {
	self.ScoreTimes = map[EmScoreTimes]int32{}
}

//删除手牌中的某张牌,delAll为true时删除所有的delcard,front为false时表示从后往前删(删除出过的牌时)
func RemoveCard(handCards []int32, delcard int32, delAll, front bool) ([]int32, bool) {
	del := false
	if front { //从前往后删
		for i, card := range handCards {
			if card == delcard {
				handCards = append(handCards[:i], handCards[i+1:]...)
				del = true
				if !delAll {
					break
				}
			}
		}
	} else { //从后往前删
		for i := len(handCards) - 1; i >= 0; i-- {
			if handCards[i] == delcard {
				handCards = append(handCards[:i], handCards[i+1:]...)
				del = true
				if !delAll {
					break
				}
			}
		}
	}
	return handCards, del
}
