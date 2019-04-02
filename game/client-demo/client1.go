package main

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbcenter "cy/game/pb/center"
	pbclub "cy/game/pb/club"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"encoding/json"
	"io"
	"os"

	pbgame_ddz "cy/game/pb/game/ddz"
	pbgame_csmj "cy/game/pb/game/mj/changshu"
	pbhall "cy/game/pb/hall"
	pblogin "cy/game/pb/login"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
)

var (
	wxID = flag.String("wxid", "wx_1", "wx id")
	// addr = flag.String("addr", "localhost:9876", "tcp listen address")
	addr = flag.String("addr", "192.168.1.128:9876", "tcp listen address")
	c    net.Conn

	loginSucc = make(chan int, 1)
	userID    uint64
	desk      deskInfo
	fileName  = flag.String("fileName", "./desk.json", "deskInfo readfileName")
)

type deskInfo struct {
	DeskId uint64 `json:"DeskId"`
}

func login() {
	sendPb(&pblogin.LoginReq{
		Head:      &pbcommon.ReqHead{Seq: 1},
		LoginType: pblogin.LoginType_WX,
		ID:        *wxID,
	})
	<-loginSucc
}

func joindesk(hasLogin bool) {
	if !hasLogin {
		login()
	}
	//从文件读取桌子号
	err := json.Unmarshal(readFile(*fileName), &desk) //第二个参数要地址传递
	if err != nil {
		fmt.Println("err = ", err)
		return
	}
	sendPb(&pbgame.JoinDeskReq{
		Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
		DeskID: desk.DeskId,
	})
}

func sendGameAction(pb proto.Message) {
	pbAction := &pbgame.GameAction{
		Head: &pbcommon.ReqHead{UserID: userID},
	}
	pbAction.ActionName, pbAction.ActionValue, _ = protobuf.Marshal(pb)
	sendPb(pbAction)
}

func throwdice() {
	sendGameAction(&pbgame_csmj.C2SThrowDice{})
}

func main() {
	flag.Parse()

	var err error
	c, err = net.Dial("tcp4", *addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(c.RemoteAddr())

	go recv()

	var a int //要发送的消息
	for {
		fmt.Printf("请输入座位号: ")
		fmt.Scan(&a)
		if a != 1 && a != 2 && a != 3 && a != 4 {
			continue
		}
		name := map[int]string{1: "wx_1", 2: "wx_2", 3: "wx_3", 4: "wx_4"}
		*wxID = name[a]
		break
	}

	for {
		fmt.Printf("请输入要发送的消息: ")
		fmt.Scan(&a)
		fmt.Println("发送消息 = ", a)
		switch a {
		case 1:
			login()
		case 2:
			makedesk()
		case 3:
			joindesk(false)
		case 4:
			throwdice()
		}
	}

	//sendPb(&pbhall.QueryGameListReq{Head: &pbcommon.ReqHead{Seq: 1, UserID: userID}})

	// sendPb(&pbclub.CreateClubReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	Name:   "name1",
	// 	Notice: "notice1",
	// 	Arg:    "arg1",
	// })

	// sendPb(&pbclub.DestoryClubReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	ClubID: 3,
	// })

	// sendPb(&pbclub.UpdateClubReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	ClubID: 10,
	// 	Notice: "Notice",
	// 	Arg:    "Arg-555",
	// })

	// sendPb(&pbclub.JoinClubReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	ClubID: 10,
	// })

	// sendPb(&pbclub.ExitClubReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	ClubID: 10,
	// })

	// sendPb(&pbclub.QueryClubByIDReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	ClubID: 10,
	// })

	// sendPb(&pbclub.QueryClubByMemberReq{
	// 	Head:   &pbcommon.ReqHead{Seq: 1, UserID: userID},
	// 	UserID: 1137,
	// })

	// 匹配
	// sendPb(&pbcenter.MatchReq{
	// 	Head:     &pbcommon.ReqHead{UserID: userID},
	// 	GameName: "ddz",
	// 	RoomId:   1,
	// })

	// 取消匹配
	// sendPb(&pbcenter.CancelMatchReq{
	// 	Head: &pbcommon.ReqHead{UserID: userID},
	// })

	// makeDeskReq := &pbgame.MakeDeskReq{
	// 	Head:     &pbcommon.ReqHead{UserID: userID, Seq: 1},
	// 	GameName: "11101",
	// 	ClubID:   0,
	// }
	/*
			  repeated cyUint32 Rule = 1;
		  uint32 Barhead = 2;
		  uint32 PlayerCount = 3;
		  uint32 Dipiao = 4;
		  RoundInfo RInfo = 5;
		  uint32 PaymentType = 6;
		  uint32 LimitIP = 7;
		  uint32 Voice = 8;
	*/
	// makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue, _ = protobuf.Marshal(&pbgame_csmj.CreateArg{
	// 	Rule:        []*pbgame_csmj.CyUint32{},
	// 	Barhead:     5,
	// 	PlayerCount: 4,
	// 	Dipiao:      1,
	// 	RInfo:       &pbgame_csmj.RoundInfo{},
	// 	PaymentType: 3,
	// 	LimitIP:     1,
	// 	Voice:       0,
	// })
	// sendPb(makeDeskReq)

	// 新建桌子
	// makeDeskReq := &pbgame.MakeDeskReq{
	// 	Head:     &pbcommon.ReqHead{UserID: userID, Seq: 1},
	// 	GameName: "ddz",
	// 	ClubID:   4,
	// }
	// makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue, _ = protobuf.Marshal(&pbgame_ddz.RoomArg{
	// 	Type:        2,
	// 	BaseScore:   5,
	// 	FeeType:     1,
	// 	PaymentType: 1,
	// 	LoopCnt:     4,
	// 	Fee:         3,
	// })
	// sendPb(makeDeskReq)

	// 查询桌子
	// sendPb(&pbgame.QueryDeskInfoReq{
	// 	Head:   &pbcommon.ReqHead{UserID: userID, Seq: 1},
	// 	DeskID: 100001,
	// })

	// 加入桌子
	// sendPb(&pbgame.JoinDeskReq{
	// 	Head:   &pbcommon.ReqHead{UserID: userID},
	// 	DeskID: 1000,
	// })

	// 离开桌子
	// sendPb(&pbgame.ExitDeskReq{
	// 	Head: &pbcommon.ReqHead{UserID: userID},
	// })

	// sendPb(&pbgame.DestroyDeskReq{
	// 	Head:   &pbcommon.ReqHead{UserID: userID},
	// 	DeskID: 1001,
	// })

	// pbAction := &pbgame.GameAction{
	// 	Head: &pbcommon.ReqHead{UserID: userID},
	// }
	// pbAction.ActionName, pbAction.ActionValue, _ = protobuf.Marshal(&pbgame_ddz.UserReadyReq{})
	// sendPb(pbAction)

	// sendPb(&pbcenter.QuerySessionInfoReq{
	// 	Head: &pbcommon.ReqHead{UserID: *userID},
	// })

}

func makedesk() {
	login()
	makeDeskReq := &pbgame.MakeDeskReq{
		Head:     &pbcommon.ReqHead{UserID: userID, Seq: 1},
		GameName: "11101",
		ClubID:   0,
	}
	makeDeskReq.GameArgMsgName, makeDeskReq.GameArgMsgValue, _ = protobuf.Marshal(&pbgame_csmj.CreateArg{
		Rule:        []*pbgame_csmj.CyUint32{},
		Barhead:     5,
		PlayerCount: 4,
		Dipiao:      1,
		RInfo:       &pbgame_csmj.RoundInfo{},
		PaymentType: 3,
		LimitIP:     1,
		Voice:       0,
	})
	sendPb(makeDeskReq)
	<-loginSucc
	joindesk(true)
}

func sendPb(pb proto.Message) {
	var err error
	m := &codec.Message{}
	m.Name, m.Payload, err = protobuf.Marshal(pb)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("send ", m.Name)
	pktReq := codec.NewPacket()
	pktReq.Msgs = append(pktReq.Msgs, m)

	err = pktReq.WriteTo(c)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func recv() {
	for {
		var err error
		pktRsp := codec.NewPacket()
		err = pktRsp.ReadFrom(c)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, msg := range pktRsp.Msgs {
			detailMsg(msg)
		}
	}
}

func detailMsg(msg *codec.Message) {
	fmt.Println("recv", msg.Name)

	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch v := pb.(type) {
	case *pblogin.LoginRsp:
		fmt.Printf("	%+v\n", v)
		if v.Code == pblogin.LoginRspCode_Succ {
			userID = v.User.UserID
			loginSucc <- 1
		}

	case *pbgame.MakeDeskRsp:
		fmt.Printf("	%+v\n", v)
		desk.DeskId = v.Info.ID
		//写入到文件中
		buf, err := json.MarshalIndent(desk, "", "	") //格式化编码
		if err != nil {
			fmt.Println("err = ", err)
			return
		}
		writebuf(*fileName, string(buf))
		loginSucc <- 1
	case *pbcommon.ErrorTip, *pbgame.JoinDeskRsp,
		*pbhall.QuerySessionInfoRsp, *pbgame.QueryDeskInfoRsp,
		*pbcenter.MatchRsp, *pbcenter.CancelMatchRsp, *pbclub.QueryClubByIDRsp,
		*pbclub.CreateClubRsp, *pbclub.RemoveClubRsp, *pbclub.UpdateClubRsp,
		*pbclub.JoinClubRsp, *pbclub.ExitClubRsp:

		fmt.Printf("	%+v\n", v)
	case *pbgame.GameNotif:
		fmt.Printf("	%+v\n", v)
		pb2, err := protobuf.Unmarshal(v.NotifName, v.NotifValue)
		if err == nil {
			switch v := pb2.(type) {
			case *pbgame_ddz.CallNotif:
				if v.UserID == userID {
					fmt.Println("轮到我叫地主咯")

					time.Sleep(time.Second * 3)

					ga := &pbgame.GameAction{}
					ga.Head = &pbcommon.ReqHead{UserID: userID, Seq: 1}
					uc := &pbgame_ddz.UserCall{}
					uc.Code = pbgame_ddz.CallCode_CCall
					ga.ActionName, ga.ActionValue, err = protobuf.Marshal(uc)
					sendPb(ga)
				}
			case *pbgame_ddz.DeskInfo:
				fmt.Printf("DeskInfo %+v \n", v)
			}
		}
	case *pbhall.QueryGameListRsp:
		fmt.Printf("	%+v\n", v)
		for _, v := range v.GameNames {
			sendPb(&pbgame.QueryGameConfigReq{
				Head:     &pbcommon.ReqHead{Seq: 1, UserID: userID},
				GameName: v,
				Type:     1,
			})
		}
	case *pbgame.QueryGameConfigRsp:
		fmt.Printf("	%+v\n", v)
		if pb2, err := protobuf.Unmarshal(v.CfgName, v.CfgValue); err == nil {
			mc, ok := pb2.(*pbgame_ddz.MatchConfig)
			if ok {
				fmt.Printf("%+v\n", mc)
			}

			fc, ok := pb2.(*pbgame_ddz.FriendsConfigTpl)
			if ok {
				fmt.Printf("%+v\n", fc)
			}
		}
	}

}

func writebuf(path string, wbuf string) {
	//打开文件，新建文件
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("err = ", err)
		return
	}

	//使用完毕，需要关闭文件
	defer f.Close()
	_, err = f.WriteString(wbuf) //n表示写入的字节数
	if err != nil {
		fmt.Println("err = ", err)
	}
}

func readFile(path string) []byte {
	//打开文件
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("err = ", err)
		return nil
	}

	//关闭文件
	defer f.Close()

	buf := make([]byte, 1024*2) //2k大小

	//n代表从文件读取内容的长度
	n, err1 := f.Read(buf)
	if err1 != nil && err1 != io.EOF { //文件出错，同时没有到结尾
		fmt.Println("err1 = ", err1)
		return nil
	}
	fmt.Printf("read from%s:%s", path, string(buf[:n]))
	return buf[:n]
}
