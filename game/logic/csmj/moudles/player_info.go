package moudles

type playerInfo struct {
	CardInfo      playerCardInfo
	BalanceInfo   playserBalanceInfo
	BaseInfo      playerBaseInfo
	BalanceResult playerBalanceResult
}

type playerCardInfo struct {
	HandCards  []uint8              //玩家手牌
	OutCards   []uint8              //出过的牌
	PengCards  map[uint8]int16      // {card=chair_id,..}
	GangCards  map[uint8]emOperType // {card=G_OP_TYPE类型,...}
	StackCards map[uint8]int16      //玩家手牌数量统计 {card=num,...}
	ChiCards   []uint8              //{card1,card2,card3,card4,card5,card6}3个连续的能组成吃
	HuCard     uint8
}

//单局结算信息
type playserBalanceInfo struct {
	GangPoint int16
	HuPoint   int16
	HuType    []emHuScoreType //emHuScoreType类型参数
}

type playerBaseInfo struct {
	ChairId  int16
	Uid      int32
	Nickname string
	Point    int32 //玩家每局得分总和
}

//总结算
type playerBalanceResult struct {
	HuPai     int16
	DianPao   int16
	AnGang    int16
	MingGang  int16
	Point     int16
	Wintimes  int16 //赢的次数
	Losetimes int16 //输的次数
}
