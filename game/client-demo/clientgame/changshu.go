package clientgame

import (
	csession "cy/game/client-demo/session"
	"cy/game/codec/protobuf"
	"encoding/json"

	// mj "cy/game/logic/changshu/majiang"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
	"fmt"

	"go.uber.org/zap"
)

type Changshu struct {
	*csession.Session
	gamename  string
	Waitchan  chan int
	BankerId  int32
	HandCards []int32
	curoper   *pbgame_logic.S2CHaveOperation
}

var (
	log      *zap.SugaredLogger //printf风格
	tlog     *zap.Logger        //structured 风格
	gamename = "11101"
)

func (self *Changshu) InitLog(_tlog *zap.Logger, _log *zap.SugaredLogger) {
	log, tlog = _log, _tlog
}

//客户端接收到消息
func (self *Changshu) DispatchRecv(msg *pbgame.GameNotif) {
	pb, err := protobuf.Unmarshal(msg.NotifName, msg.NotifValue)
	if err != nil {
		tlog.Info("Unmarshal err", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", msg.NotifValue))
		return
	}
	log.Infof("recv game msg,NotifName:%s NotifValue:%s", msg.NotifName, util.PB2JSON(pb, true))
	switch v := pb.(type) {
	case *pbgame_logic.S2CThrowDice:
		if v.ChairId == self.ChairId {
			tlog.Info("recv GameNotif 轮到我投色子")
		} else {
			tlog.Info("recv GameNotif 轮到玩家投色子", zap.Int32("chair", v.ChairId))
		}
	case *pbgame_logic.BS2CThrowDiceResult:
		tlog.Info("recv GameNotif 玩家投了色子")
	case *pbgame_logic.S2CChangePos:
		tlog.Info("recv GameNotif 换位结果")
		self.dealChangPos(v)
	case *pbgame_logic.S2CStartGame:
		tlog.Info("recv GameNotif 游戏开始")
		self.S2CStartGame(v)
	case *pbgame_logic.S2CHaveOperation:
		tlog.Info("recv GameNotif 我可以进行操作")
		self.readHaveOper(v)
	case *pbgame_logic.BS2COutCard:
		tlog.Info("recv GameNotif 玩家出了牌")
		self.dealOutCard(v)
	case *pbgame_logic.BS2CChiCard:
		tlog.Info("recv GameNotif 玩家吃牌")
		self.dealChiCard(v)
	case *pbgame_logic.BS2CPengCard:
		tlog.Info("recv GameNotif 玩家碰牌")
		self.dealPengCard(v)
	case *pbgame_logic.BS2CGangCard:
		tlog.Info("recv GameNotif 玩家杠牌")
		self.dealGangCard(v)
	case *pbgame_logic.BS2CHuCard:
		tlog.Info("recv GameNotif 玩家胡牌")
		self.dealHuCard(v)
	case *pbgame_logic.BS2CDrawCard:
		tlog.Info("recv GameNotif 玩家摸牌")
		self.dealDrawCard(v)
	default:
		tlog.Info("recv GameNotif 未处理", zap.String("NotifName", msg.NotifName))
	}
}

func getFirstBuHua(str string) (msg *pbgame_logic.Json_FirstBuHua) {
	msg = &pbgame_logic.Json_FirstBuHua{}
	json.Unmarshal([]byte(str), &msg)
	return
}
func (self *Changshu) dealChangPos(msg *pbgame_logic.S2CChangePos) {
	for _, v := range msg.PosInfo {
		if v.UserId == self.UserId {
			tlog.Info("我换了位置", zap.Int32("old chair", self.ChairId), zap.Int32("new chair", v.UserPos))
			self.ChairId = v.UserPos
		}
	}
}

func (self *Changshu) dealDrawCard(msg *pbgame_logic.BS2CDrawCard) {
	tmp := getFirstBuHua(msg.JsonDrawInfo)
	if msg.ChairId == self.ChairId {
		tlog.Info("我摸了牌", zap.Any("MoCards", tmp.MoCards), zap.Any("HuaCards", tmp.HuaCards))
		self.updateCardInfo(tmp.MoCards, nil)
	} else {
		tlog.Info("别人出了牌", zap.Any("MoCards", tmp.MoCards), zap.Any("HuaCards", tmp.HuaCards))
	}

}

func (self *Changshu) dealOutCard(msg *pbgame_logic.BS2COutCard) {
	if msg.ChairId == self.ChairId {
		tlog.Info("我出了牌", zap.Int32("Card", msg.Card))
		self.updateCardInfo(nil, []int32{msg.Card})
	} else {
		tlog.Info("别人出了牌", zap.Int32("Card", msg.Card), zap.Int32("chair", msg.ChairId))
	}

}

func (self *Changshu) dealChiCard(msg *pbgame_logic.BS2CChiCard) {
	if msg.ChairId == self.ChairId {
		tlog.Info("我吃牌", zap.Int32("Card", msg.Card), zap.Uint32("ChiType", msg.ChiType))
		if isFlag(msg.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)) {
			self.updateCardInfo(nil, []int32{msg.Card - 2, msg.Card - 1})
		} else if isFlag(msg.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)) {
			self.updateCardInfo(nil, []int32{msg.Card - 1, msg.Card + 1})
		} else if isFlag(msg.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)) {
			self.updateCardInfo(nil, []int32{msg.Card + 1, msg.Card + 2})
		}
	} else {
		tlog.Info("别人吃牌", zap.Int32("Card", msg.Card), zap.Uint32("ChiType", msg.ChiType), zap.Int32("chair", msg.ChairId))
	}

}

func (self *Changshu) dealPengCard(msg *pbgame_logic.BS2CPengCard) {
	if msg.ChairId == self.ChairId {
		tlog.Info("我碰牌", zap.Int32("Card", msg.Card))
		self.updateCardInfo(nil, []int32{msg.Card, msg.Card})
	} else {
		tlog.Info("别人碰牌", zap.Int32("Card", msg.Card), zap.Int32("chair", msg.ChairId))
	}

}

func (self *Changshu) dealGangCard(msg *pbgame_logic.BS2CGangCard) {
	if msg.ChairId == self.ChairId {
		tlog.Info("我杠牌", zap.Int32("Card", msg.Card))
		self.updateCardInfo(nil, []int32{msg.Card, msg.Card})
		if msg.Type == pbgame_logic.GangType_GangType_AN {

		}
		return
	} else {
		tlog.Info("别人杠牌", zap.Int32("Card", msg.Card), zap.Int32("chair", msg.ChairId))
	}

}

func (self *Changshu) dealHuCard(msg *pbgame_logic.BS2CHuCard) {
	if msg.ChairId == self.ChairId {
		tlog.Info("我胡牌", zap.Any("HandCards", msg.HandCards))
		return
	} else {
		tlog.Info("别人胡牌", zap.Any("HandCards", msg.HandCards), zap.Int32("chair", msg.ChairId))
	}

}

// func (self *Changshu) S2CBuHua(msg *pbgame_logic.S2CBuHua) {
// 	tlog.Info("补花前self.HandCards", zap.Any("handcards", self.HandCards))
// 	for _, once := range msg.BuHuaResult {
// 		//去掉花牌
// 		self.removeCard(once.HuaCard, false)
// 		tlog.Info("补花一次", zap.Int32("once.HuaCard", once.HuaCard), zap.Int32("once.BuCard", once.BuCard))
// 		//增加补的牌
// 		self.HandCards = append(self.HandCards, once.BuCard)
// 	}
// 	tlog.Info("补花后self.HandCards", zap.Any("handcards", self.HandCards))
// }

func (self *Changshu) removeCard(delcard int32, delAll bool) {
	for i, card := range self.HandCards {
		if card == delcard {
			self.HandCards = append(self.HandCards[:i], self.HandCards[i+1:]...)
			if !delAll {
				break
			}
		}
	}
}
func switchToInt32(cards []*pbgame_logic.Cyint32) []int32 {
	res := []int32{}
	for _, card := range cards {
		res = append(res, card.T)
	}
	return res
}

func (self *Changshu) S2CStartGame(msg *pbgame_logic.S2CStartGame) {
	self.BankerId = msg.BankerId
	// h := self.ChairId * 14
	// e := h + 13
	// self.HandCards = switchToInt32(msg.AllUserCards[h:e])
	tmp := pbgame_logic.Json_UserCardInfo{}
	json.Unmarshal([]byte(msg.JsonAllCards), &tmp)
	self.HandCards = tmp.HandCards[self.ChairId].Cards
	tlog.Info("self.HandCards", zap.Any("handcards", self.HandCards))
	self.readHaveOper(msg.BankerOper)
}

func isFlag(mask, typeFlag uint32) bool {
	result := mask & typeFlag
	return result == typeFlag
}

func (self *Changshu) readHaveOper(oper *pbgame_logic.S2CHaveOperation) {
	if oper == nil {
		return
	}

	log.Infof("oper=%s", util.PB2JSON(oper, true))
	if isFlag(oper.OperMask, uint32(pbgame_logic.CanOperMask_OperMaskChi)) {
		tlog.Info("我能吃")
	}
	if isFlag(oper.OperMask, uint32(pbgame_logic.CanOperMask_OperMaskPeng)) {
		tlog.Info("我能碰")
	}
	if isFlag(oper.OperMask, uint32(pbgame_logic.CanOperMask_OperMaskGang)) {
		tlog.Info("我能杠")
	}
	if isFlag(oper.OperMask, uint32(pbgame_logic.CanOperMask_OperMaskHu)) {
		tlog.Info("我能胡")
	}
	self.curoper = oper
}
func (self *Changshu) MakeDeskReq() *pbgame.MakeDeskReq {
	makeDeskReq := &pbgame.MakeDeskReq{
		Head:     &pbcommon.ReqHead{UserID: self.Session.UserId, Seq: 1},
		GameName: gamename,
		ClubID:   0,
	}
	makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue, _ = protobuf.Marshal(&pbgame_logic.CreateArg{
		Rule:        []*pbgame_logic.CyU32String{},
		Barhead:     5,
		PlayerCount: 4,
		Dipiao:      1,
		RInfo:       &pbgame_logic.RoundInfo{},
		PaymentType: 3,
		LimitIP:     1,
		Voice:       0,
	})
	return makeDeskReq
}

//客户端操作
func (self *Changshu) DoAction(act string) {
	switch act {
	case "dice":
		self.dothrowdice()
	case "oper":
		self.operAction()
	case "out":
		self.outCard()
	}

}

//
func (self *Changshu) dothrowdice() {
	self.Session.SendGameAction(&pbgame_logic.C2SThrowDice{})
}

func (self *Changshu) operAction() {
	oper := self.curoper
	if oper.OperMask != 0 {
		tlog.Error("当前不能吃碰杠胡")
	}
	log.Infof("我能做的操作:oper=%s", util.PB2JSON(oper, true))
	var pick string
	fmt.Printf("请输入选择: ")
	fmt.Scan(&pick)
	switch pick {
	case "c":
		str := "我能"
		if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)) {
			str = str + "左吃,"
		}
		if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)) {
			str = str + "中吃,"
		}
		if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)) {
			str = str + "右吃"
		}
		tlog.Info(str)
		var chipick string
		fmt.Printf("请选择吃类型: ")
		fmt.Scan(&chipick)
		msg := &pbgame_logic.C2SChiCard{Card: oper.Card}
		switch chipick {
		case "l":
			msg.ChiType = uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)
		case "m":
			msg.ChiType = uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)
		case "r":
			msg.ChiType = uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)
		}
		self.SendGameAction(msg)
	case "p":
		self.SendGameAction(&pbgame_logic.C2SPengCard{Card: oper.Card})
	case "g":
		tlog.Info("能杠的牌", zap.Any("Cards", oper.CanGang.Cards))
		var card int32
		fmt.Printf("请选择杠的牌: ")
		fmt.Scan(&card)
		self.SendGameAction(&pbgame_logic.C2SGangCard{Card: card})
	case "h":
		self.SendGameAction(&pbgame_logic.C2SHuCard{})
	}
}

func (self *Changshu) outCard() {
	tlog.Info("我的手牌", zap.Any("self.Handcards", self.HandCards))
	var card int32
	fmt.Printf("请选择出的牌: ")
	fmt.Scan(&card)
	self.SendGameAction(&pbgame_logic.C2SOutCard{Card: card})
}

func (self *Changshu) updateCardInfo(addCards, subCards []int32) {
	if len(addCards) != 0 {
		self.HandCards = append(self.HandCards, addCards...)
	} else if len(subCards) != 0 {
		for _, card := range subCards {
			self.HandCards = RemoveCard(self.HandCards, card, false)
		}
	}
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
