package clientgame

import (
	csession "cy/game/client-demo/session"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_csmj "cy/game/pb/game/mj/changshu"

	"go.uber.org/zap"
)

type Changshu struct {
	*csession.Session
	gamename  string
	Waitchan  chan int
	BankerId  int32
	HandCards []int32
	chairId int32
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
	case *pbgame_csmj.S2CThrowDice:
		tlog.Info("recv GameNotif 轮到玩家投色子", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_csmj.S2CThrowDiceResult:
		tlog.Info("recv GameNotif 玩家投了色子", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_csmj.S2CChangePos:
		tlog.Info("recv GameNotif 换位结果", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
	case *pbgame_csmj.S2CStartGame:
		tlog.Info("recv GameNotif 游戏开始", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
		self.S2CStartGame(v)
	case *pbgame_csmj.S2CBuHua:
		if v.ChairId== {
			
		}
		tlog.Info("recv GameNotif 游戏开始", zap.String("NotifName", msg.NotifName), zap.Any("NotifValue", v))
		self.S2CStartGame(v)
	}
}

func (self *Changshu) S2CStartGame(msg *pbgame_csmj.S2CStartGame) {
	self.BankerId = msg.BankerId
	self.HandCards = msg.UserInfo.HandCards
	tlog.Info("self.HandCards", zap.Any("handcards", self.HandCards))
}

func (self *Changshu) MakeDeskReq() *pbgame.MakeDeskReq {
	makeDeskReq := &pbgame.MakeDeskReq{
		Head:     &pbcommon.ReqHead{UserID: self.Session.UserId, Seq: 1},
		GameName: gamename,
		ClubID:   0,
	}
	makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue, _ = protobuf.Marshal(&pbgame_csmj.CreateArg{
		Rule:        []*pbgame_csmj.CyU32String{},
		Barhead:     5,
		PlayerCount: 4,
		Dipiao:      1,
		RInfo:       &pbgame_csmj.RoundInfo{},
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
	self.Session.SendGameAction(&pbgame_csmj.C2SThrowDice{})
}
