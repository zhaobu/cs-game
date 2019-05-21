package main

import (
	clientgame "cy/game/client-demo/clientgame"
	csession "cy/game/client-demo/session"
	"cy/game/codec"
	zaplog "cy/game/common/logger"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgamerecord "cy/game/pb/gamerecord"
	pblogin "cy/game/pb/login"
	"time"

	"cy/game/util"
	"encoding/json"
	"fmt"
	"net"

	"go.uber.org/zap"
)

type Player struct {
	curgame  *clientgame.Changshu //当前的游戏
	pos      int32                //选择的座位号
	wxID     string
	waitchan chan int
	csession.Session
}

type deskInfo struct {
	DeskId uint64 `json:"DeskId"`
}

func (self *Player) init() {
	//初始化
	self.waitchan = make(chan int, 0)
	self.Session = csession.Session{}
	self.curgame = &clientgame.Changshu{Waitchan: self.waitchan, Session: &self.Session}
}

func (self *Player) Connect(addr string) {
	self.init()
	var a int32 //要发送的消息
	for {
		fmt.Printf("请输入座位号: ")
		fmt.Scan(&a)
		if a != 1 && a != 2 && a != 3 && a != 4 {
			continue
		}
		name := map[int32]string{1: "wx_1", 2: "wx_2", 3: "wx_3", 4: "wx_4"}
		self.wxID = name[a]
		self.pos = a
		break
	}

	tlog = zaplog.InitLogger(self.wxID+".log", "debug", true)
	log = tlog.Sugar()
	self.curgame.InitLog(tlog, log)
	self.InitLog(tlog, log)

	var err error
	self.Conn, err = net.Dial("tcp4", addr)
	if err != nil {
		tlog.Error("connect err", zap.Error(err))
		return
	}
	tlog.Info("connect succ", zap.Any("addr", self.Conn.RemoteAddr()))
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
		case "4":
			self.exitdesk()
		case "5":
			self.destroyDesk()
		case "record":
			self.queryGameRecord()
		default:
			self.curgame.DoAction(a)
		}
	}
}

//查询战绩
func (self *Player) queryGameRecord() {
	tlog.Info("查询战绩类型:1按userid查,2按俱乐部id查,3按俱乐部+房间号查")
	var (
		queryType      int32
		queryUid       uint64 //要查询的uid
		queryClubId    int64  //俱乐部Id
		queryRoomId    uint64 //房间号
		queryStartTime int64  //查询时间范围
		queryEndTime   int64  //查询时间范围
	)

	for {
		fmt.Printf("请选择查询类型:")
		fmt.Scan(&queryType)
		if queryType != 1 && queryType != 2 && queryType != 3 {
			fmt.Printf("查询类型错误:")
			continue
		}
		break
	}
	queryStartTime = time.Now().Unix()
	queryEndTime = time.Now().Unix()
	if queryType == 1 {
		fmt.Printf("请输入要查询的uid,默认查询自己:")
		fmt.Scan(&queryUid)
		if queryUid == 0 {
			queryUid = self.UserId
		}
	} else if queryType == 2 {
		fmt.Printf("请输入要查询的俱乐部id:")
		fmt.Scan(&queryClubId)
	} else if queryType == 3 {
		fmt.Printf("请输入要查询的俱乐部id:")
		fmt.Scan(&queryClubId)
		fmt.Printf("请输入要查询的房间号:")
		fmt.Scan(&queryRoomId)
	}
	self.SendPb(&pbgamerecord.QueryRoomRecordReq{Head: &pbcommon.ReqHead{Seq: 1},
		QueryType:      queryType,
		QueryUserId:    queryUid,
		QueryClubId:    queryClubId,
		QueryRoomId:    queryRoomId,
		QueryStartTime: queryStartTime,
		QueryEndTime:   queryEndTime,
	})
}
func (self *Player) login() {
	self.SendPb(&pblogin.LoginReq{
		Head:      &pbcommon.ReqHead{Seq: 1},
		LoginType: pblogin.LoginType_WX,
		ID:        self.wxID,
	})
	<-self.waitchan
	tlog.Info("login suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) makedesk() {
	self.SendPb(self.curgame.MakeDeskReq())
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
	self.DeskId = desk.DeskId
	self.SendPb(&pbgame.JoinDeskReq{
		Head:   &pbcommon.ReqHead{Seq: 1, UserID: self.UserId},
		DeskID: desk.DeskId,
	})
	<-self.waitchan
	tlog.Info("joindesk suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
	self.sitdown(self.pos - 1)
}

func (self *Player) exitdesk() {
	self.SendPb(&pbgame.ExitDeskReq{
		Head: &pbcommon.ReqHead{Seq: 1, UserID: self.UserId},
	})
	<-self.waitchan
	tlog.Info("exitdesk suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) destroyDesk() {
	self.SendPb(&pbgame.DestroyDeskReq{
		Head:   &pbcommon.ReqHead{Seq: 1, UserID: self.UserId},
		DeskID: self.DeskId,
		Type:   pbgame.DestroyDeskType_DestroyTypeDebug,
	})
	<-self.waitchan
	tlog.Info("destroyDesk suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) sitdown(chairId int32) {
	self.SendPb(&pbgame.SitDownReq{
		Head:    &pbcommon.ReqHead{Seq: 1, UserID: self.UserId},
		ChairId: chairId,
	})
	<-self.waitchan
	self.ChairId = chairId
	tlog.Info("sitdown suc", zap.String("wxID", self.wxID), zap.Uint64("UserId", self.UserId))
}

func (self *Player) recv() {
	for {
		var err error
		pktRsp := codec.NewPacket()
		err = pktRsp.ReadFrom(self.Conn)
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
			// tlog.Info("recv", zap.String("msgName", msg.Name), zap.Any("msgValue", pb))
			if msg.Name != "pbgame.GameNotif" {
				log.Infof("revcv msgName:%s msgValue:%s", msg.Name, util.PB2JSON(pb, true))
			}
			switch v := pb.(type) {
			case *pblogin.LoginRsp:
				if v.Code == pblogin.LoginRspCode_Succ {
					self.UserId = v.User.UserID
					self.waitchan <- 1
				}
			case *pbgame.MakeDeskRsp:
				desk := &deskInfo{DeskId: v.Info.ID}
				//写入到文件中
				util.WriteJSON(*fileName, desk)
				// buf, err := json.MarshalIndent(desk, "", "	") //格式化编码
				// if err != nil {
				// 	fmt.Println("err = ", err)
				// 	return
				// }
				// writebuf(*fileName, string(buf))
				self.waitchan <- 1
			case *pbgame.JoinDeskRsp:
				self.waitchan <- 1
			case *pbgame.SitDownRsp:
				self.waitchan <- 1
			case *pbgame.ExitDeskRsp:
				self.waitchan <- 1
			case *pbgame.DestroyDeskRsp: //申请解散回应
			case *pbgame.DestroyDeskResultNotif: //房间解散成功
				self.waitchan <- 1
			case *pbgame.GameNotif:
				self.curgame.DispatchRecv(v)
			default:
				tlog.Info("未处理的消息", zap.String("msgName", msg.Name))
			}
		}
	}
}
