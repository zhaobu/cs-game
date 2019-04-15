package clientgame

import (
	csession "cy/game/client-demo/session"
	"cy/game/codec/protobuf"
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
	chairId   int32
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
	switch v := pb.(type) {
	case *pbgame_logic.S2CThrowDice:
		tlog.Info("recv GameNotif 轮到玩家投色子", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_logic.BS2CThrowDiceResult:
		tlog.Info("recv GameNotif 玩家投了色子", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_logic.S2CChangePos:
		tlog.Info("recv GameNotif 换位结果", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_logic.S2CStartGame:
		tlog.Info("recv GameNotif 游戏开始", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
		self.S2CStartGame(v)
	case *pbgame_logic.S2CHaveOperation:
		tlog.Info("recv GameNotif 我可以进行操作", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
		self.readHaveOper(v)
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
	self.HandCards = switchToInt32(msg.HandCards)
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
		oper := self.curoper
		if oper.OperMask != 0 {
			var pick string
			fmt.Printf("请输入选择: ")
			fmt.Scan(&pick)
			switch pick {
			case "c":
				if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)) {
					tlog.Info("我能左吃")
				}
				if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)) {
					tlog.Info("我能中吃")
				}
				if isFlag(oper.CanChi.ChiType, uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)) {
					tlog.Info("我能右吃")
				}
				var chipick string
				fmt.Printf("请选择吃类型: ")
				fmt.Scan(&chipick)
				switch chipick {
				case "l":
					self.SendPb(&pbgame_logic.C2SChiCard{Card: oper.Card, ChiType: uint32(pbgame_logic.ChiTypeMask_ChiMaskLeft)})
				case "m":
					self.SendPb(&pbgame_logic.C2SChiCard{Card: oper.Card, ChiType: uint32(pbgame_logic.ChiTypeMask_ChiMaskMiddle)})
				case "r":
					self.SendPb(&pbgame_logic.C2SChiCard{Card: oper.Card, ChiType: uint32(pbgame_logic.ChiTypeMask_ChiMaskRight)})
				}
			case "p":
				self.SendPb(&pbgame_logic.C2SPengCard{Card: oper.Card})
			case "g":
				tlog.Info("能杠的牌", zap.Any("Cards", oper.CanGang.Cards))
				var card int32
				fmt.Printf("请选择杠的牌: ")
				fmt.Scan(&card)
				self.SendPb(&pbgame_logic.C2SGangCard{Card: card})
			case "h":
				self.SendPb(&pbgame_logic.C2SHuCard{})
			}
		}
	}

}

//
func (self *Changshu) dothrowdice() {
	self.Session.SendGameAction(&pbgame_logic.C2SThrowDice{})
}
