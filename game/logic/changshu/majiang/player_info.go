package majiang

type PlayerInfo struct {
	CardInfo      PlayerCardInfo
	BalanceInfo   PlayserBalanceInfo //局结算
	BaseInfo      PlayerBaseInfo
	BalanceResult PlayerBalanceResult //总结算
}

func MakePlayers() *PlayerInfo {
	p := &PlayerInfo{}
	p.init()
	return p
}
func (self *PlayerInfo) init() {
	self.CardInfo.reset()
	self.BalanceInfo.reset()
	self.BalanceResult.init()
}

func (self *PlayerInfo) reset() {
	self.CardInfo.reset()
	self.BalanceInfo.reset()
}

type PlayerCardInfo struct {
	HandCards  []int32              //玩家手牌
	OutCards   []int32              //出过的牌
	PengCards  map[int32]int32      // {card=chair_id,..}
	GangCards  map[int32]EmOperType // {card=G_OP_TYPE类型,...}
	StackCards map[int32]int32      //玩家手牌数量统计 {card=num,...}
	ChiCards   [][3]int32           //{card1,card2,card3,card4,card5,card6}3个连续的能组成吃,吃的牌放第一个
}

func (self *PlayerCardInfo) reset() {
	self.HandCards = []int32{}
	self.OutCards = []int32{}
	self.PengCards = map[int32]int32{}
	self.GangCards = map[int32]EmOperType{}
	self.StackCards = map[int32]int32{}
	self.ChiCards = [][3]int32{}
}

//单局结算信息
type PlayserBalanceInfo struct {
	HuCard    int32
	GangPoint int32
	HuPoint   int32
	HuType    []EmHuScoreType //emHuScoreType类型参数
}

func (self *PlayserBalanceInfo) reset() {
	self.HuCard = 0
	self.GangPoint = 0
	self.HuPoint = 0
	self.HuType = []EmHuScoreType{}
}

//玩家基础信息
type PlayerBaseInfo struct {
	ChairId  int32
	Uid      uint64
	Nickname string
	Point    int32 //玩家每局得分总和
}

//总结算
type PlayerBalanceResult struct {
	ScoreTimes map[EmScoreTimes]int32 //统计次数
	Point      int32                  //总得分
}

func (self *PlayerBalanceResult) init() {
	self.ScoreTimes = map[EmScoreTimes]int32{}
}

//删除手牌中的某张牌,delAll为true时删除所有的delcard
func RemoveCard(handCards []int32, delcard int32, delAll bool) []int32 {
	for i, card := range handCards {
		if card == delcard {
			handCards = append(handCards[:i], handCards[i+1:]...)
			if !delAll {
				break
			}
		}
	}
	return handCards
}
