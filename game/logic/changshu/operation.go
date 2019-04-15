package main

import (
	mj "cy/game/logic/changshu/majiang"
	pbgame_logic "cy/game/pb/game/mj/changshu"
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
	OpChair   int32         //
	HuList    mj.HuTypeList //胡牌类型列表
}

func (self *CanHuOper) Empty() bool {
	return len(self.HuList) == 0
}

//默认创建函数
// func NewCanChiOper() *CanChiOper {
// 	return &CanChiOper{
// 		ChairId: -1,
// 	}
// }
// func NewCanPengOper() *CanPengOper {
// 	return &CanPengOper{
// 		ChairId:   -1,
// 		LoseChair: -1,
// 	}
// }
// func NewCanGangOper() *CanGangOper {
// 	return &CanGangOper{
// 		ChairId: -1,
// 	}
// }
// func NewCanHuOper() *CanHuOper {
// 	return &CanHuOper{
// 		LoseChair: -1,
// 		OpChair:   -1,
// 	}
// }

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

//操作优先级
type OperPriority struct {
	ChairId int32
	Op      PriorityOrder
	Info    interface{} //操作信息
	// Card     int32
	// GangType string
	// ChiCard  ChiCardTb
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
func (self *OperAtion) moCanGang(stackCards map[int32]int32, pengCards map[int32]int32, card int32) (bool, map[int32]int32) {
	ret := map[int32]int32{}
	for card, num := range stackCards {
		if num == 4 { //原来的暗杠
			ret[card] = card
		}
	}
	if num, ok := stackCards[card]; ok && num == 3 { //摸到的牌组成暗杠
		ret[card] = card
	} else if _, ok := pengCards[card]; ok {
		ret[card] = card
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
			cardInfo.HandCards = mj.RemoveCard(cardInfo.HandCards, card, false)
		}
	}
}

//分析游戏开始发牌后庄家能做的操作
func (self *OperAtion) BankerAnalysis(cardInfo mj.PlayerCardInfo) *CanOperInfo {
	ret := &CanOperInfo{}

	//判断是否能胡
	if ok, huOper := huLib.CheckHuType(&cardInfo); ok {
		ret.CanHu.HuList = huOper
	}

	card := cardInfo.HandCards[len(cardInfo.HandCards)-1]
	self.updateCardInfo(&cardInfo, nil, []int32{card})

	stackCards, pengCards := cardInfo.StackCards, cardInfo.PengCards

	//检查能否杠(包括暗杠,补杠)
	if ok, gangOper := self.moCanGang(stackCards, pengCards, card); ok {
		ret.CanGang.GangList = gangOper
	}
	return ret
}

//摸牌后分析能做的操作
func (self *OperAtion) DrawcardAnalysis(cardInfo *mj.PlayerCardInfo, card int32, leftCardNum int32) *CanOperInfo {
	ret := &CanOperInfo{}

	stackCards, pengCards := cardInfo.StackCards, cardInfo.PengCards

	//检查能否杠(包括暗杠,补杠)
	if ok, gangOper := self.moCanGang(stackCards, pengCards, card); ok {
		ret.CanGang.GangList = gangOper
	}

	//判断是否能胡
	tmpCardInfo := *cardInfo
	self.updateCardInfo(&tmpCardInfo, []int32{card}, nil)
	if ok, huOper := huLib.CheckHuType(&tmpCardInfo); ok {
		ret.CanHu.HuList = huOper
	}
	return ret
}

func GetNextChair(chairId, playerCount int32) int32 {
	if chairId == playerCount {
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
func (self *OperAtion) OutCardAnalysis(cardInfo *mj.PlayerCardInfo, outCard, chairId, outChair, leftCardNum int32) *CanOperInfo {
	log.Infof("玩家%d出牌,检测玩家%d能做的操作", outChair, chairId)
	ret := &CanOperInfo{}
	if leftCardNum > 0 {
		if ok, chi := self.checkChi(cardInfo.StackCards, outCard, chairId, outChair); ok {
			ret.CanChi.ChiType = chi
		}
		if self.checkPeng(cardInfo.StackCards, outCard) {
			ret.CanPeng.Card = outCard
		}
		if self.checkPengGang(cardInfo.StackCards, outCard) {
			ret.CanGang.GangList[outCard] = outCard
		}
	}
	//判断是否能胡
	tmpCardInfo := *cardInfo
	self.updateCardInfo(&tmpCardInfo, []int32{outCard}, nil)
	if ok, huOper := huLib.CheckHuType(&tmpCardInfo); ok {
		ret.CanHu.HuList = huOper
	}
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
func (self *OperAtion) HandleChiCard(cardInfo *mj.PlayerCardInfo, loseCardInfo *mj.PlayerCardInfo, card int32, chiType uint32) {
	eatGroup := [3]int32{}
	//根据吃牌类型生成组合
	if chiType == chiType&uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft) {
		eatGroup = [3]int32{card, card + 1, card + 2}
	} else if chiType == chiType&uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle) {
		eatGroup = [3]int32{card, card - 1, card + 1}
	} else if chiType == chiType&uint32(pbgame_logic.ChiTypeMask_ChiMaskRight) {
		eatGroup = [3]int32{card, card - 2, card - 1}
	}
	self.updateCardInfo(cardInfo, nil, eatGroup[:])
	cardInfo.ChiCards = append(cardInfo.ChiCards, eatGroup)

	//处理出牌玩家,把牌从出过的牌中拿走
	loseCardInfo.OutCards = mj.RemoveCard(loseCardInfo.OutCards, card, false)
}

//处理碰牌(cardInfo为碰牌玩家,loseCardInfo为出牌玩家)
func (self *OperAtion) HandlePengCard(cardInfo *mj.PlayerCardInfo, loseCardInfo *mj.PlayerCardInfo, card, loseChair int32) {
	//处理碰牌玩家
	self.updateCardInfo(cardInfo, nil, []int32{card, card})
	cardInfo.PengCards[card] = loseChair
	//处理出牌玩家,把牌从出过的牌中拿走
	loseCardInfo.OutCards = mj.RemoveCard(loseCardInfo.OutCards, card, false)
}

func (self *OperAtion) HandleGangCard(cardInfo *mj.PlayerCardInfo, loseCardInfo *mj.PlayerCardInfo, card int32, gangType mj.EmOperType) {
	if gangType == mj.OperType_BU_GANG { //补杠
		self.updateCardInfo(cardInfo, nil, []int32{card})
		delete(cardInfo.PengCards, card)
	} else if gangType == mj.OperType_AN_GANG { //暗杠
		self.updateCardInfo(cardInfo, nil, []int32{card, card, card, card})
	} else if gangType == mj.OperType_MING_GANG { //明杠
		self.updateCardInfo(cardInfo, nil, []int32{card, card, card})
		//处理出牌玩家,把牌从出过的牌中拿走
		loseCardInfo.OutCards = mj.RemoveCard(loseCardInfo.OutCards, card, false)
	} else {
		log.Errorf("杠类型错误,gangType=%d", gangType)
	}
	cardInfo.GangCards[card] = gangType
}

//判断杠的类型
func (self *OperAtion) GetGangType(cardInfo *mj.PlayerCardInfo, card int32) mj.EmOperType {
	if cardInfo.StackCards[card] == 4 { //暗杠
		return mj.OperType_AN_GANG
	} else if _, ok := cardInfo.PengCards[card]; ok { //补杠
		return mj.OperType_BU_GANG
	} else if cardInfo.StackCards[card] == 3 { //明杠
		return mj.OperType_MING_GANG
	}
	log.Errorf("杠类型判断错误")
	return mj.OperType_None
}

func (self *OperAtion) QiangGangAnalysis(cardInfo *mj.PlayerCardInfo, outCard, chairId, loseChair int32) *CanOperInfo {
	log.Infof("玩家%d补杠,检测玩家%d能否抢杠胡", loseChair, chairId)
	ret := &CanOperInfo{}

	//判断是否能胡
	tmpCardInfo := *cardInfo
	self.updateCardInfo(&tmpCardInfo, []int32{outCard}, nil)
	if ok, huOper := huLib.CheckHuType(&tmpCardInfo); ok {
		ret.CanHu.HuList = huOper
	}
	return ret
}
