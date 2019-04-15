package clientgame

import (
	csession "cy/game/client-demo/session"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"

	"go.uber.org/zap"
)

type Changshu struct {
	*csession.Session
	gamename  string
	Waitchan  chan int
	BankerId  int32
	HandCards []int32
	chairId   int32
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
	case *pbgame_logic.BS2COutCard:
		if v.ChairId == self.ChairId {
			tlog.Info("recv GameNotif 轮到我补花", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
			self.S2CBuHua(v)
		}
	}
}

func (self *Changshu) S2CBuHua(msg *pbgame_logic.S2CBuHua) {
	tlog.Info("补花前self.HandCards", zap.Any("handcards", self.HandCards))
	for _, once := range msg.BuHuaResult {
		//去掉花牌
		self.removeCard(once.HuaCard, false)
		tlog.Info("补花一次", zap.Int32("once.HuaCard", once.HuaCard), zap.Int32("once.BuCard", once.BuCard))
		//增加补的牌
		self.HandCards = append(self.HandCards, once.BuCard)
	}
	tlog.Info("补花后self.HandCards", zap.Any("handcards", self.HandCards))
}

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
	}

}

//
func (self *Changshu) dothrowdice() {
	self.Session.SendGameAction(&pbgame_logic.C2SThrowDice{})
}
