package main

import (
	mj "cy/game/logic/changshu/majiang"
)

//麻将操作
type OperAtion struct {
}

type ChiCardTb [2]int32 //用来吃的2张牌

type CanOperInfo struct {
	CanChi  *CanChiOper
	CanPeng *CanPengOper
	CanGang *CanGangOper
	CanHu   *CanHuOper
}

func (self *CanOperInfo) Empty() bool {
	return self.CanChi == nil && self.CanPeng == nil && self.CanGang == nil && self.CanHu == nil
}

type CanChiOper struct {
	ChairId int32
	ChiList []ChiCardTb
}

type CanPengOper struct {
	ChairId   int32
	LoseChair int32
	Card      int32
}

type CanGangOper struct {
	ChairId  int32
	GangList map[int32]int32
}

type CanHuOper struct {
	HuMode    mj.EmHuMode   //胡牌方式
	LoseChair int32         //丢分玩家
	Card      int32         //胡的牌
	OpChair   int32         //
	HuList    mj.HuTypeList //胡牌类型列表
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

//优先级的操作
type PriorityOper struct {
	ChairId  int32
	Card     int32
	GangType string
	Op       PriorityOrder
	ChiCard  ChiCardTb
}

// type PriorityOper struct {
// 	ChairId  int32
// 	Card     int32
// 	GangType string
// 	Op       PriorityOrder
// 	ChiCard  ChiCardTb
// }

// func (self *PriorityOper) ResetPriorityOper() {
// 	self.ChairId = -1
// 	self.Card = 0
// 	self.GangType = ""
// 	self.Op = NoneOrder
// 	self.ChiCard = ChiCardTb{}
// }

// /返回摸到牌后所有能杠的牌值
func moCanGang(stackCards map[int32]int32, pengCards map[int32]int32, card int32) *CanGangOper {
	ret := &CanGangOper{GangList: map[int32]int32{}}
	for card, num := range stackCards {
		if num == 4 { //原来的暗杠
			ret.GangList[card] = card
		}
	}
	if num, ok := stackCards[card]; ok && num == 3 { //摸到的牌组成暗杠
		ret.GangList[card] = card
	} else if _, ok := pengCards[card]; ok {
		ret.GangList[card] = card
	}
	return ret
}

func updateCardInfo(cardInfo *mj.PlayerCardInfo, addCard, subCard int32) {
	if addCard != 0 {
		mj.Add_stack(cardInfo.StackCards, addCard)
		cardInfo.HandCards = append(cardInfo.HandCards, addCard)
	} else if subCard != 0 {
		mj.Sub_stack(cardInfo.StackCards, subCard)
		cardInfo.HandCards = mj.RemoveCard(cardInfo.HandCards, subCard, false)
	}
}

//摸牌后分析能做的操作
func DrawcardAnalysis(cardInfo *mj.PlayerCardInfo, card int32, leftCardNum int32) *CanOperInfo {
	ret := &CanOperInfo{}

	stackCards, pengCards := cardInfo.StackCards, cardInfo.PengCards

	//检查能否杠(包括暗杠,补杠)
	gangOper := moCanGang(stackCards, pengCards, card)
	if len(gangOper.GangList) > 0 {
		ret.CanGang = gangOper
	}

	//判断是否能胡
	tmpCardInfo := *cardInfo
	updateCardInfo(&tmpCardInfo, card, 0)
	ok, huOper := huLib.CheckHuType(&tmpCardInfo)
	if ok {
		ret.CanHu.HuList = huOper
	}
	return ret
}

//处理摸牌
func HandleDrawCard(cardInfo *mj.PlayerCardInfo, card int32) {
	updateCardInfo(cardInfo, card, 0)
}
