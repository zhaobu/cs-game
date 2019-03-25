package main

import (
	mj "cy/game/logic/changshu/majiang"
)

//麻将操作
type OperAtion struct {
}

type ChiCardTb [2]int32 //用来吃的2张牌

type CanOperInfo struct {
	CanChi  CanChiOper
	CanPeng CanPengOper
	CanGang CanGangOper
	CanHu   CanHuOper
}

type CanChiOper struct {
	ChairId int32
	ChiList []ChiCardTb
}

type CanPengOper struct {
	ChairId   int32
	LoseChair int32
	Card      map[int32]int32
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
