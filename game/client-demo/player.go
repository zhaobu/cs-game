package main

import (
	clientgame "cy/game/client-demo/clientgame"
	csession "cy/game/client-demo/session"
	"cy/game/codec"
	zaplog "cy/game/common/logger"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pblogin "cy/game/pb/login"
	"encoding/json"
	"fmt"
	"net"

	"go.uber.org/zap"
)

type Player struct {
	curgame  *clientgame.Changshu //当前的游戏
	ChairId  int32
	UserId   uint64
	wxID     string
	waitchan chan int
	session  *csession.Session
}

type deskInfo struct {
	DeskId uint64 `json:"DeskId"`
}

func (self *Player) init() {
	//初始化
	self.waitchan = make(chan int, 0)
	self.session = &csession.Session{}
	self.curgame = &clientgame.Changshu{Waitchan: self.waitchan, Session: self.session}
}

func (self *Player) Connect(addr string) {
	self.init()
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
	self.curgame.InitLog(tlog, log)
	self.session.InitLog(tlog, log)

	var err error
	self.session.Conn, err = net.Dial("tcp4", addr)
	if err != nil {
		tlog.Error("connect err", zap.Error(err))
		return
	}
	tlog.Info("connect succ", zap.Any("addr", self.session.Conn.RemoteAddr()))
	go self.recv()

	self.login()
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
			self.joindesk()
		default:
			self.curgame.DoAction(a)
		}
	}
}

func (self *Player) login() {
	self.session.SendPb(&pblogin.LoginReq{
		Head:      &pbcommon.ReqHead{Seq: 1},
		LoginType: pblogin.LoginType_WX,
		ID:        self.wxID,
	})
	<-self.waitchan
	tlog.Info("login suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) makedesk() {
	self.session.SendPb(self.curgame.MakeDeskReq())
	<-self.waitchan
	self.joindesk()
}

func (self *Player) joindesk() {
	desk := &deskInfo{}
	//从文件读取桌子号
	err := json.Unmarshal(readFile(*fileName), desk) //第二个参数要地址传递
	if err != nil {
		tlog.Error("Unmarshal err", zap.Error(err))
		return
	}
	self.session.SendPb(&pbgame.JoinDeskReq{
		Head:   &pbcommon.ReqHead{Seq: 1, UserID: self.UserId},
		DeskID: desk.DeskId,
	})
	<-self.waitchan
	tlog.Info("joindesk suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) recv() {
	for {
		var err error
		pktRsp := codec.NewPacket()
		err = pktRsp.ReadFrom(self.session.Conn)
		if err != nil {
			tlog.Error("ReadFrom err", zap.Error(err))
			return
		}

		for _, msg := range pktRsp.Msgs {
			pb, err := codec.Msg2Pb(msg)
			if err != nil {
				tlog.Error("Unmarshal err", zap.Error(err))
				return
			}
			tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", pb))

			switch v := pb.(type) {
			case *pblogin.LoginRsp:
				if v.Code == pblogin.LoginRspCode_Succ {
					self.UserId = v.User.UserID
					self.session.UserId = self.UserId
					self.waitchan <- 1
				}
			case *pbgame.MakeDeskRsp:
				desk := &deskInfo{DeskId: v.Info.ID}
				//写入到文件中
				buf, err := json.MarshalIndent(desk, "", "	") //格式化编码
				if err != nil {
					fmt.Println("err = ", err)
					return
				}
				writebuf(*fileName, string(buf))
				self.waitchan <- 1
			case *pbgame.JoinDeskRsp:
				self.waitchan <- 1
			case *pbgame.GameNotif:
				self.curgame.DispatchRecv(v)
			default:
				tlog.Info("未处理的消息", zap.String("msgName", msg.Name), zap.Any("msgValue", v))
			}
		}
	}
}
