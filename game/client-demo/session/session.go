package session

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"net"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

var (
	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func (self *Session) InitLog(_tlog *zap.Logger, _log *zap.SugaredLogger) {
	log, tlog = _log, _tlog
}

type Session struct {
	UserId uint64
	Conn   net.Conn
}

//发送游戏命令
func (self *Session) SendGameAction(pb proto.Message) {
	pbAction := &pbgame.GameAction{
		Head: &pbcommon.ReqHead{UserID: self.UserId},
	}
	pbAction.ActionName, pbAction.ActionValue, _ = protobuf.Marshal(pb)
	self.SendPb(pbAction)
}

//发送消息
func (self *Session) SendPb(pb proto.Message) {
	var err error
	m := &codec.Message{}
	m.Name, m.Payload, err = protobuf.Marshal(pb)
	if err != nil {
		tlog.Error("Marshal err", zap.Error(err))
		return
	}
	tlog.Info("SendPb", zap.String("msgName", m.Name))
	pktReq := codec.NewPacket()
	pktReq.Msgs = append(pktReq.Msgs, m)

	err = pktReq.WriteTo(self.Conn)
	if err != nil {
		tlog.Error("WriteTo err", zap.Error(err))
		return
	}
}
