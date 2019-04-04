package main

import (
	clientgame "cy/game/client-demo/clientgame"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	zaplog "cy/game/common/logger"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_csmj "cy/game/pb/game/mj/changshu"
	pblogin "cy/game/pb/login"
	"encoding/json"
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
)

type Player struct {
	ChairId  int32
	UserId   uint64
	wxID     string
	waitchan chan int
	gamename string
	session  net.Conn
}

var curgame *clientgame.Changshu

func (self *Player) Connect(addr string) {

	var a int //要发送的消息
	for {
		fmt.Printf("请输入座位号: ")
		fmt.Scan(&a)
		if a != 1 && a != 2 && a != 3 && a != 4 {
			continue
		}
		name := map[int]string{1: "wx_1", 2: "wx_2", 3: "wx_3", 4: "wx_4"}
		self.wxID = name[a]
		break
	}

	tlog = zaplog.InitLogger(self.wxID+".txt", "debug", true)
	log = tlog.Sugar()

	var err error
	self.session, err = net.Dial("tcp4", addr)
	if err != nil {
		tlog.Error("connect err", zap.Error(err))
		return
	}
	tlog.Info("connect succ", zap.Any("addr", self.session.RemoteAddr()))

	self.waitchan = make(chan int, 0)
	self.login()
	go self.recv()
}

func (self *Player) GameStart() {
	var a string
	for {
		fmt.Printf("请输入要发送的消息: ")
		fmt.Scan(&a)
		fmt.Println("发送消息 = ", a)
		switch a {
		case "1":
			self.login()
		case "2":
			self.makedesk()
		case "3":
			self.joindesk(false)
		case "dice":
			self.throwdice()
		}
	}
}

func (self *Player) login() {
	self.sendPb(&pblogin.LoginReq{
		Head:      &pbcommon.ReqHead{Seq: 1},
		LoginType: pblogin.LoginType_WX,
		ID:        self.wxID,
	})
	<-self.waitchan
	tlog.Info("login suc", zap.String("wxID", self.wxID))
}

func (self *Player) makedesk() {
	makeDeskReq := &pbgame.MakeDeskReq{
		Head:     &pbcommon.ReqHead{UserID: self.UserId, Seq: 1},
		GameName: self.gamename,
		ClubID:   0,
	}
	makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue = curgame.MakeDeskReq()
	self.sendPb(makeDeskReq)
	<-self.waitchan
	self.joindesk(true)
}

func (self *Player) joindesk(hasLogin bool) {
	if !hasLogin {
		self.login()
	}
	//从文件读取桌子号
	err := json.Unmarshal(readFile(*fileName), desk) //第二个参数要地址传递
	if err != nil {
		fmt.Println("err = ", err)
		return
	}
	self.sendPb(&pbgame.JoinDeskReq{
		Head:   &pbcommon.ReqHead{Seq: 1, UserID: self.UserID},
		DeskID: desk.DeskId,
	})
}

func (self *Player) recv() {
	game := &clientgame.Changshu{waitchan: self.waitchan}
	for {
		var err error
		pktRsp := codec.NewPacket()
		err = pktRsp.ReadFrom(self.session)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, msg := range pktRsp.Msgs {
			pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
			if err != nil {
				fmt.Println(err)
				return
			}

			switch v := pb.(type) {
			case *pblogin.LoginRsp:
				tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", v))
				if v.Code == pblogin.LoginRspCode_Succ {
					self.UserId = v.User.UserID
					self.waitchan <- 1
				}

			case *pbgame.MakeDeskRsp:
				tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", v))
				desk := &deskInfo{DeskId: v.Info.ID}
				//写入到文件中
				buf, err := json.MarshalIndent(desk, "", "	") //格式化编码
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				writebuf(*fileName, string(buf))
				self.waitchan <- 1
			}
			game.DetailMsg(msg)
		}
	}
}

func (self *Player) sendGameAction(pb proto.Message) {
	pbAction := &pbgame.GameAction{
		Head: &pbcommon.ReqHead{UserID: self.UserId},
	}
	pbAction.ActionName, pbAction.ActionValue, _ = protobuf.Marshal(pb)
	sendPb(pbAction)
}

func (self *Player) sendPb(pb proto.Message) {
	var err error
	m := &codec.Message{}
	m.Name, m.Payload, err = protobuf.Marshal(pb)
	if err != nil {
		tlog.Error("Marshal err", zap.Error(err))
		return
	}
	fmt.Println("send ", m.Name)
	pktReq := codec.NewPacket()
	pktReq.Msgs = append(pktReq.Msgs, m)

	err = pktReq.WriteTo(self.session)
	if err != nil {
		tlog.Error("WriteTo err", zap.Error(err))
		return
	}
}

func (self *Player) throwdice() {
	sendGameAction(&pbgame_csmj.C2SThrowDice{})
}
