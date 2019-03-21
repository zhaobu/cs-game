package majiang

type PlayerInfo struct {
	CardInfo      PlayerCardInfo
	BalanceInfo   PlayserBalanceInfo
	BaseInfo      PlayerBaseInfo
	BalanceResult PlayerBalanceResult
}

type PlayerCardInfo struct {
	HandCards  []int32              //玩家手牌
	OutCards   []int32              //出过的牌
	PengCards  map[int32]int32      // {card=chair_id,..}
	GangCards  map[int32]EmOperType // {card=G_OP_TYPE类型,...}
	StackCards map[int32]int32      //玩家手牌数量统计 {card=num,...}
	ChiCards   []int32              //{card1,card2,card3,card4,card5,card6}3个连续的能组成吃
	HuCard     int32
}

//单局结算信息
type PlayserBalanceInfo struct {
	GangPoint int32
	HuPoint   int32
	HuType    []EmHuScoreType //emHuScoreType类型参数
}

type PlayerBaseInfo struct {
	ChairId  int32
	Uid      uint64
	Nickname string
	Point    int32 //玩家每局得分总和
}

//总结算
type PlayerBalanceResult struct {
	HuPai     int32
	DianPao   int32
	AnGang    int32
	MingGang  int32
	Point     int32
	Wintimes  int32 //赢的次数
	Losetimes int32 //输的次数
}
