package main

import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
)

//麻将操作
type OperAtion struct {
	game_config *pbgame_logic.CreateArg //游戏参数
	laiziCard   map[int32]int32         //癞子牌
}

type ChiCardTb [2]int32 //用来吃的2张牌

type CanOperInfo struct {
	CanChi  CanChiOper
	CanPeng CanPengOper
	CanGang CanGangOper
	CanHu   CanHuOper
}

func (self *CanOperInfo) Empty() bool {
	return self.CanChi.Empty() && self.CanPeng.Empty() && self.CanGang.Empty() && self.CanHu.Empty()
}

type CanChiOper struct {
	Card    int32  //被吃的牌
	ChairId int32  //吃牌玩家
	ChiType uint32 //吃牌类型
}

func (self *CanChiOper) Empty() bool {
	return self.ChiType == 0
}

type CanPengOper struct {
	ChairId   int32
	LoseChair int32
	Card      int32
}

func (self *CanPengOper) Empty() bool {
	return self.Card == 0
}

type CanGangOper struct {
	ChairId  int32
	GangList map[int32]int32
}

func (self *CanGangOper) Empty() bool {
	return len(self.GangList) == 0
}

type CanHuOper struct {
	HuMode    mj.EmHuMode   //胡牌方式
	LoseChair int32         //丢分玩家
	Card      int32         //胡的牌
	HuList    mj.HuTypeList //胡牌类型列表
}

func (self *CanHuOper) Empty() bool {
	return len(self.HuList) == 0
}

//默认创建函数
func NewCanChiOper() *CanChiOper {
	return &CanChiOper{
		ChairId: -1,
	}
}
func NewCanPengOper() *CanPengOper {
	return &CanPengOper{
		ChairId:   -1,
		LoseChair: -1,
	}
}
func NewCanGangOper() *CanGangOper {
	return &CanGangOper{
		ChairId:  -1,
		GangList: map[int32]int32{},
	}
}
func NewCanHuOper() *CanHuOper {
	return &CanHuOper{
		LoseChair: -1,
		HuList:    mj.HuTypeList{},
	}
}

func NewCanOper() *CanOperInfo {
	return &CanOperInfo{
		CanChi:  *NewCanChiOper(),
		CanPeng: *NewCanPengOper(),
		CanGang: *NewCanGangOper(),
		CanHu:   *NewCanHuOper(),
	}
}

// func (self *CanOperInfo) ResetCanOper() {
// 	self.CanChi = *NewCanChiOper()
// 	self.CanPeng = *NewCanPengOper()
// 	self.CanGang = *NewCanGangOper()
// 	self.CanHu = *NewCanHuOper()
// }

//操作优先级
type PriorityOrder int

const (
	NoneOrder PriorityOrder = iota
	ChiOrder
	PengOrder
	GangOrder
	HuOrder
)

//记录唤醒操作
type WaitOperRecord struct {
	Card    int32  //玩家选择吃碰杠的牌
	ChiType uint32 //吃牌时的吃类型
}

//操作优先级
type OperPriority struct {
	ChairId int32
	Op      PriorityOrder
	Info    interface{} //操作信息,记录环境操作时对应WaitOper结构
}

// type OperPriority struct {
// 	ChairId  int32
// 	Card     int32
// 	GangType string
// 	Op       PriorityOrder
// 	ChiCard  ChiCardTb
// }

// func (self *OperPriority) ResetOperPriority() {
// 	self.ChairId = -1
// 	self.Card = 0
// 	self.GangType = ""
// 	self.Op = NoneOrder
// 	self.ChiCard = ChiCardTb{}
// }

func (self *OperAtion) Init(config *pbgame_logic.CreateArg, laizi map[int32]int32) {
	self.game_config = config
	self.laiziCard = laizi
}

//获取两个优先级中的较大者
func (self *OperAtion) GetMaxOrder(a, b PriorityOrder) PriorityOrder {
	if a > b {
		return a
	}
	return b
}

// /返回摸到牌后所有能杠的牌值
func (self *OperAtion) moCanGang(stackCards map[int32]int32, pengCards map[int32]int32, _card int32) (bool, map[int32]int32) {
	ret := map[int32]int32{}
	for card, num := range stackCards {
		if num == 4 { //原来的暗杠
			ret[card] = card
		} else if num == 1 {
			if _, ok := pengCards[card]; ok { //补杠
				ret[card] = card
			}
		}
	}
	if num, ok := stackCards[_card]; ok && num == 3 { //摸到的牌组成暗杠
		ret[_card] = _card
	} else if _, ok := pengCards[_card]; ok { //补杠
		ret[_card] = _card
	}
	return len(ret) > 0, ret
}

//更新手牌数据
func (self *OperAtion) updateCardInfo(cardInfo *mj.PlayerCardInfo, addCards, subCards []int32) {
	if len(addCards) != 0 {
		mj.Add_stack(cardInfo.StackCards, addCards...)
		cardInfo.HandCards = append(cardInfo.HandCards, addCards...)
	} else if len(subCards) != 0 {
		mj.Sub_stack(cardInfo.StackCards, subCards...)
		for _, card := range subCards {
			cardInfo.HandCards, _ = mj.RemoveCard(cardInfo.HandCards, card, false)
		}
	}
}

//分析游戏开始发牌后庄家能做的操作
func (self *OperAtion) BankerAnalysis(playerInfo *mj.PlayerInfo, chairId int32, huMode mj.EmHuMode) *CanOperInfo {
	ret := NewCanOper()
	cardInfo := &playerInfo.CardInfo
	//判断是否能胡
	card := cardInfo.HandCards[len(cardInfo.HandCards)-1]
	if ok, huOper := huLib.CheckHuType(cardInfo, &playerInfo.BalanceInfo, huMode, nil); ok {
		ret.CanHu = CanHuOper{HuMode: huMode, HuList: huOper, Card: card, LoseChair: -1}
	}
	self.updateCardInfo(cardInfo, nil, []int32{card}) //减掉一张手牌
	stackCards, pengCards := cardInfo.StackCards, cardInfo.PengCards
	//检查能否杠(包括暗杠,补杠)
	if ok, gangOper := self.moCanGang(stackCards, pengCards, card); ok {
		ret.CanGang = CanGangOper{GangList: gangOper, ChairId: chairId}
	}
	self.updateCardInfo(cardInfo, []int32{card}, nil) //还原手牌
	return ret
}

//吃碰后分析能做的操作
func (self *OperAtion) AfterChiPengAnalysis(cardInfo *mj.PlayerCardInfo, chairId, pengCard int32) *CanOperInfo {
	ret := NewCanOper()
	ret.CanGang.ChairId = chairId
	//检查能否暗杠
	for card, num := range cardInfo.StackCards {
		if num == 4 { //原来的暗杠
			ret.CanGang.GangList[card] = card
		} else if num == 1 && pengCard != card {
			if _, ok := cardInfo.PengCards[card]; ok { //除了当前碰的那张牌外,其他的牌能补杠
				ret.CanGang.GangList[card] = card
			}
		}
	}
	return ret
}

//摸牌后分析能做的操作
func (self *OperAtion) DrawcardAnalysis(playerInfo *mj.PlayerInfo, chairId, card int32, leftCardNum int32, huModeTags map[mj.EmHuModeTag]bool) *CanOperInfo {
	ret := NewCanOper()
	cardInfo := &playerInfo.CardInfo
	stackCards, pengCards := cardInfo.StackCards, cardInfo.PengCards

	//检查能否杠(包括暗杠,补杠)
	if ok, gangOper := self.moCanGang(stackCards, pengCards, card); ok {
		ret.CanGang = CanGangOper{GangList: gangOper, ChairId: chairId}
	}

	if leftCardNum == 1 {
		huModeTags[mj.HuModeTag_HaiDiLaoYue] = true
	}
	//判断是否能胡
	self.updateCardInfo(cardInfo, []int32{card}, nil)
	if ok, huOper := huLib.CheckHuType(cardInfo, &playerInfo.BalanceInfo, mj.HuMode_ZIMO, huModeTags); ok {
		ret.CanHu = CanHuOper{HuMode: mj.HuMode_ZIMO, HuList: huOper, Card: card, LoseChair: -1}
	}
	self.updateCardInfo(cardInfo, nil, []int32{card}) //还原手牌
	return ret
}

func GetNextChair(chairId, playerCount int32) int32 {
	if chairId+1 == playerCount {
		return 0
	}
	return chairId + 1
}

//能否吃
func (self *OperAtion) checkChi(stackCards map[int32]int32, outCard, chairId, outChair int32) (bool, uint32) {
	//下家才能吃
	if GetNextChair(outChair, self.game_config.PlayerCount) != chairId {
		return false, 0
	}
	//检测是否癞子,字牌
	if _, ok := self.laiziCard[outCard]; ok || outCard >= 41 {
		return false, 0
	}
	var ret uint32 = 0
	if stackCards[outCard+1] > 0 && stackCards[outCard+2] > 0 { //左吃(11,12,13,其中11为被吃的牌)
		ret = ret | uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)
	}
	if stackCards[outCard-1] > 0 && stackCards[outCard+1] > 0 { //中吃(11,12,13,其中12为被吃的牌)
		ret = ret | uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)
	}
	if stackCards[outCard-2] > 0 && stackCards[outCard-1] > 0 { //右吃(11,12,13,其中13为被吃的牌)
		ret = ret | uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)
	}
	return ret != 0, ret
}

//能否碰
func (self *OperAtion) checkPeng(stackCards map[int32]int32, card int32) bool {
	if _, ok := self.laiziCard[card]; ok {
		return false
	}
	return stackCards[card] >= 2
}

//能否明杠
func (self *OperAtion) checkPengGang(stackCards map[int32]int32, card int32) bool {
	if _, ok := self.laiziCard[card]; ok {
		return false
	}
	return stackCards[card] == 3
}

//出牌后分析能做的操作
func (self *OperAtion) OutCardAnalysis(playerInfo *mj.PlayerInfo, outCard, chairId, outChair, leftCardNum int32) *CanOperInfo {
	log.Infof("玩家%d出牌,检测玩家%d能做的操作", outChair, chairId)
	cardInfo := &playerInfo.CardInfo
	ret := NewCanOper()
	if leftCardNum > 0 {
		if self.game_config.PlayerCount != 3 { //三人麻将不能吃
			if ok, chi := self.checkChi(cardInfo.StackCards, outCard, chairId, outChair); ok {
				ret.CanChi = CanChiOper{ChiType: chi, Card: outCard, ChairId: chairId}
			}
		}
		if cardInfo.GuoPeng[outCard] != outCard && self.checkPeng(cardInfo.StackCards, outCard) {
			ret.CanPeng = CanPengOper{Card: outCard, ChairId: chairId, LoseChair: outChair}
		}
		if self.checkPengGang(cardInfo.StackCards, outCard) {
			ret.CanGang.GangList[outCard] = outCard
			ret.CanGang.ChairId = chairId
		}
	}
	//判断是否能胡
	self.updateCardInfo(cardInfo, []int32{outCard}, nil)
	if ok, huOper := huLib.CheckHuType(cardInfo, &playerInfo.BalanceInfo, mj.HuMode_PAOHU, nil); ok {
		ret.CanHu = CanHuOper{HuMode: mj.HuMode_PAOHU, HuList: huOper, Card: outCard, LoseChair: outChair}
	}
	self.updateCardInfo(cardInfo, nil, []int32{outCard}) //还原手牌
	return ret
}

//处理摸牌
func (self *OperAtion) HandleDrawCard(cardInfo *mj.PlayerCardInfo, card int32) {
	self.updateCardInfo(cardInfo, []int32{card}, nil)
}

//处理出牌
func (self *OperAtion) HandleOutCard(cardInfo *mj.PlayerCardInfo, card int32) {
	self.updateCardInfo(cardInfo, nil, []int32{card})
	cardInfo.OutCards = append(cardInfo.OutCards, card)
}

//处理吃牌(cardInfo为吃牌玩家,loseCardInfo为出牌玩家)
func (self *OperAtion) HandleChiCard(cardInfo *mj.PlayerCardInfo, loseCardInfo *mj.PlayerCardInfo, card, loseChair int32, chiType uint32) {
	eatGroup := [3]int32{}
	var operType pbgame_logic.OperType
	//根据吃牌类型生成组合
	if isFlag(chiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)) {
		eatGroup = [3]int32{card, card + 1, card + 2}
		operType = pbgame_logic.OperType_Oper_LCHI
	} else if isFlag(chiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)) {
		eatGroup = [3]int32{card, card - 1, card + 1}
		operType = pbgame_logic.OperType_Oper_MCHI
	} else if isFlag(chiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)) {
		eatGroup = [3]int32{card, card - 2, card - 1}
		operType = pbgame_logic.OperType_Oper_RCHI
	}
	self.updateCardInfo(cardInfo, nil, eatGroup[1:])
	cardInfo.ChiCards = append(cardInfo.ChiCards, eatGroup)
	cardInfo.RiverCards = append(cardInfo.RiverCards, &pbgame_logic.OperRecord{Type: operType, Card: card, LoseChair: loseChair})

	//处理出牌玩家,把牌从出过的牌中拿走
	loseCardInfo.OutCards, _ = mj.RemoveCard(loseCardInfo.OutCards, card, false)
}

//处理碰牌(cardInfo为碰牌玩家,loseCardInfo为出牌玩家)
func (self *OperAtion) HandlePengCard(playerInfo *mj.PlayerInfo, loseCardInfo *mj.PlayerCardInfo, card, loseChair int32) {
	//处理碰牌玩家
	cardInfo := &playerInfo.CardInfo
	self.updateCardInfo(cardInfo, nil, []int32{card, card})
	cardInfo.PengCards[card] = loseChair
	cardInfo.RiverCards = append(cardInfo.RiverCards, &pbgame_logic.OperRecord{Type: pbgame_logic.OperType_Oper_PENG, Card: card, LoseChair: loseChair})
	//处理出牌玩家,把牌从出过的牌中拿走
	loseCardInfo.OutCards, _ = mj.RemoveCard(loseCardInfo.OutCards, card, false)
	if card >= 41 && card <= 46 { //东南西北中发白碰牌算1花
		playerInfo.BalanceInfo.FengPoint += 1
	}
}

//胡牌时留手3张算2花(cardInfo为碰牌玩家,loseCardInfo为出牌玩家)
func (self *OperAtion) HandleRemainPengCard(playerInfo *mj.PlayerInfo) {
	cardInfo := &playerInfo.CardInfo
	for card, v := range cardInfo.StackCards {
		if card >= 41 && card <= 46 && v == 3 { //东南西北中发白碰牌算2花
			playerInfo.BalanceInfo.FengPoint += 2
		}
	}
}

//处理杠牌
func (self *OperAtion) HandleGangCard(playerInfo *mj.PlayerInfo, loseCardInfo *mj.PlayerCardInfo, card int32, gangType pbgame_logic.OperType, loseChair int32) {
	cardInfo := &playerInfo.CardInfo
	if gangType == pbgame_logic.OperType_Oper_BU_GANG { //补杠
		self.updateCardInfo(cardInfo, nil, []int32{card})
		for _, v := range cardInfo.RiverCards { //补杠时去掉之前的碰
			if v.Card == card {
				v.Type = pbgame_logic.OperType_Oper_BU_GANG
				break
			}
		}
		delete(cardInfo.PengCards, card)
		if card >= 41 { //万条筒补杠算1花，东南西北中发白补杠算3花
			playerInfo.BalanceInfo.GangPoint += 3
		} else {
			playerInfo.BalanceInfo.GangPoint += 1
		}
	} else if gangType == pbgame_logic.OperType_Oper_AN_GANG { //暗杠
		self.updateCardInfo(cardInfo, nil, []int32{card, card, card, card})
		cardInfo.RiverCards = append(cardInfo.RiverCards, &pbgame_logic.OperRecord{Type: gangType, Card: card, LoseChair: loseChair})
		if card >= 41 { //万条筒暗杠算2花，东南西北中发白暗杠算4花
			playerInfo.BalanceInfo.GangPoint += 4
		} else {
			playerInfo.BalanceInfo.GangPoint += 2
		}
	} else if gangType == pbgame_logic.OperType_Oper_MING_GANG { //明杠
		self.updateCardInfo(cardInfo, nil, []int32{card, card, card})
		cardInfo.RiverCards = append(cardInfo.RiverCards, &pbgame_logic.OperRecord{Type: gangType, Card: card, LoseChair: loseChair})
		//处理出牌玩家,把牌从出过的牌中拿走
		loseCardInfo.OutCards, _ = mj.RemoveCard(loseCardInfo.OutCards, card, false)
		if card >= 41 { //万条筒明杠算1花，东南西北中发白明杠算3花
			playerInfo.BalanceInfo.GangPoint += 3
		} else {
			playerInfo.BalanceInfo.GangPoint += 1
		}
	} else {
		log.Errorf("杠类型错误,gangType=%d", gangType)
	}
	cardInfo.GangCards[card] = gangType

}

//处理被抢杠胡玩家
func (self *OperAtion) HandleCancelQiangGangHu(playerInfo *mj.PlayerInfo, card int32) {
	cardInfo := &playerInfo.CardInfo
	for _, v := range cardInfo.RiverCards { //去掉之前的补杠变为碰
		if v.Card == card {
			v.Type = pbgame_logic.OperType_Oper_PENG
			break
		}
	}
	cardInfo.PengCards[card] = card
	if card >= 41 { //万条筒补杠算1花，东南西北中发白补杠算3花
		playerInfo.BalanceInfo.GangPoint -= 3
	} else {
		playerInfo.BalanceInfo.GangPoint -= 1
	}
	delete(cardInfo.GangCards, card)
}

//处理补花
func (self *OperAtion) HandleBuHua(playerInfo *mj.PlayerInfo, huaCards []int32) {
	playerInfo.CardInfo.HuaCards = append(playerInfo.CardInfo.HuaCards, huaCards...)
	playerInfo.BalanceInfo.BuHuaPoint += int32(len(huaCards))
}

//判断杠的类型
func (self *OperAtion) GetGangType(cardInfo *mj.PlayerCardInfo, card int32) pbgame_logic.OperType {
	if cardInfo.StackCards[card] == 4 { //暗杠
		return pbgame_logic.OperType_Oper_AN_GANG
	} else if _, ok := cardInfo.PengCards[card]; ok { //补杠
		return pbgame_logic.OperType_Oper_BU_GANG
	} else if cardInfo.StackCards[card] == 3 { //明杠
		return pbgame_logic.OperType_Oper_MING_GANG
	}
	log.Errorf("杠类型判断错误")
	return pbgame_logic.OperType_Oper_None
}

func (self *OperAtion) QiangGangAnalysis(playerInfo *mj.PlayerInfo, card, chairId, loseChair int32) *CanOperInfo {
	log.Infof("玩家%d补杠,检测玩家%d能否抢杠胡", loseChair, chairId)
	ret := NewCanOper()
	cardInfo := &playerInfo.CardInfo
	//判断是否能胡
	self.updateCardInfo(cardInfo, []int32{card}, nil)
	if ok, huOper := huLib.CheckHuType(cardInfo, &playerInfo.BalanceInfo, mj.HuMode_PAOHU, map[mj.EmHuModeTag]bool{mj.HuModeTag_QiangGangHu: true}); ok {
		ret.CanHu = CanHuOper{HuMode: mj.HuMode_PAOHU, HuList: huOper, Card: card, LoseChair: loseChair}
	}
	self.updateCardInfo(cardInfo, nil, []int32{card}) //还原手牌
	return ret
}

func isFlag(value, checkType uint32) bool {
	return value&checkType == checkType
}

//听牌分析
func (self *OperAtion) GetListenInfo(chairId int32, players []*mj.PlayerInfo, huModeTags map[mj.EmHuModeTag]bool, leftCard []int32) (res *pbgame_logic.S2CListenCards, canListen bool) {
	jsonStr := &pbgame_logic.Json_Listen{Info: map[int32]*pbgame_logic.Json_ListenCards{}}
	cardInfo := &players[chairId].CardInfo
	//统计剩余数量
	leftStack := mj.CalStackCards(leftCard, true)
	//加上其他玩家手中的数量
	for k, v := range players {
		if int32(k) != chairId {
			for card, num := range v.CardInfo.StackCards {
				if !mj.IsHuaCard(card) {
					leftStack[card] += num
				}
			}
		}
	}
	bakHandCards := make([]int32, len(cardInfo.HandCards)) //保留原手牌顺序
	copy(bakHandCards, cardInfo.HandCards)
	for removeCard, _ := range cardInfo.StackCards {
		if _, ok := cardInfo.CanNotOut[removeCard]; !ok {
			self.updateCardInfo(cardInfo, nil, []int32{removeCard})
			//先判断能胡
			if huLib.OneCardCanListen(cardInfo, &players[chairId].BalanceInfo, huModeTags) {
				jsonStr.Info[removeCard] = &pbgame_logic.Json_ListenCards{}
				//找出所有能听的牌
				for addCard, num := range leftStack {
					self.updateCardInfo(cardInfo, []int32{addCard}, nil)
					if huLib.OneCardCanHu(cardInfo, &players[chairId].BalanceInfo, huModeTags) {
						oneListen := &pbgame_logic.Json_ListenOnecard{Card: addCard, Num: num} //能听的一张牌
						jsonStr.Info[removeCard].Cards = append(jsonStr.Info[removeCard].Cards, oneListen)
					}
					self.updateCardInfo(cardInfo, nil, []int32{addCard})
				}
				//如果能听的牌在牌库和其他玩家手牌中不存在则去掉该种听牌情况
				if len(jsonStr.Info[removeCard].Cards) == 0 {
					delete(jsonStr.Info, removeCard)
				} else {
					canListen = true
				}
			}
			self.updateCardInfo(cardInfo, []int32{removeCard}, nil) //还原手牌
		}
	}
	cardInfo.HandCards = bakHandCards
	if canListen {
		res = &pbgame_logic.S2CListenCards{ListenResult: util.PB2JSON(jsonStr, false)}
	}
	return
}

//听牌分析(不用打牌判断听)
func (self *OperAtion) GetListenInfo2(chairId int32, players []*mj.PlayerInfo, huModeTags map[mj.EmHuModeTag]bool, leftCard []int32) (res *pbgame_logic.S2CListenCards, canListen bool) {
	jsonStr := &pbgame_logic.Json_Listen{Info: map[int32]*pbgame_logic.Json_ListenCards{}}
	cardInfo := &players[chairId].CardInfo
	//统计剩余数量
	leftStack := mj.CalStackCards(leftCard, true)
	//加上其他玩家手中的数量
	for k, v := range players {
		if int32(k) != chairId {
			for card, num := range v.CardInfo.StackCards {
				if !mj.IsHuaCard(card) {
					leftStack[card] += num
				}
			}
		}
	}
	//先判断能胡
	if huLib.OneCardCanListen(cardInfo, &players[chairId].BalanceInfo, huModeTags) {
		bakHandCards := make([]int32, len(cardInfo.HandCards)) //保留原手牌顺序
		copy(bakHandCards, cardInfo.HandCards)
		jsonStr.Info[0] = &pbgame_logic.Json_ListenCards{}
		//找出所有能听的牌
		for addCard, num := range leftStack {
			self.updateCardInfo(cardInfo, []int32{addCard}, nil)
			if huLib.OneCardCanHu(cardInfo, &players[chairId].BalanceInfo, huModeTags) {
				oneListen := &pbgame_logic.Json_ListenOnecard{Card: addCard, Num: num} //能听的一张牌
				jsonStr.Info[0].Cards = append(jsonStr.Info[0].Cards, oneListen)
			}
			self.updateCardInfo(cardInfo, nil, []int32{addCard})
		}
		//如果能听的牌在牌库和其他玩家手牌中不存在则去掉该种听牌情况
		if len(jsonStr.Info[0].Cards) == 0 {
			delete(jsonStr.Info, 0)
		} else {
			canListen = true
		}
		cardInfo.HandCards = bakHandCards
	}
	if canListen {
		res = &pbgame_logic.S2CListenCards{ListenResult: util.PB2JSON(jsonStr, false)}
	}
	return
}
