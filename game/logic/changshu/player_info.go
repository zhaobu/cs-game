package main

type playerInfo struct {
	cardInfo      playerCardInfo
	balanceInfo   playserBalanceInfo
	baseInfo      playerBaseInfo
	balanceResult playerBalanceResult
}

type playerCardInfo struct {
	handCards  []int32              //玩家手牌
	outCards   []int32              //出过的牌
	pengCards  map[int32]int32      // {card=chair_id,..}
	gangCards  map[int32]emOperType // {card=G_OP_TYPE类型,...}
	stackCards map[int32]int32      //玩家手牌数量统计 {card=num,...}
	chiCards   []int32              //{card1,card2,card3,card4,card5,card6}3个连续的能组成吃
	huCard     int32
}

//单局结算信息
type playserBalanceInfo struct {
	gangPoint int32
	huPoint   int32
	huType    []emHuScoreType //emHuScoreType类型参数
}

type playerBaseInfo struct {
	chairId  int32
	uid      uint64
	nickname string
	point    int32 //玩家每局得分总和
}

//总结算
type playerBalanceResult struct {
	huPai     int32
	dianPao   int32
	anGang    int32
	mingGang  int32
	point     int32
	wintimes  int32 //赢的次数
	losetimes int32 //输的次数
}
