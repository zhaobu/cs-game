package main

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/gamerecord"
	"flag"
	"fmt"
	//"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"time"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.0.117:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	cliCenter client.XClient
)

func main()  {
	flag.Parse()

	servicePath := "center"
	d := client.NewConsulDiscovery(*basePath, servicePath, []string{*consulAddr}, nil)
	cliCenter = client.NewXClient(servicePath, client.Failfast, client.RoundRobin, d, client.DefaultOption)

	//写数据测试
	//data := &pbgamerecord.WriteGameRecordReq{
	//	RoomRecordId        : "223456123",
	//	GameId				: "changshu",
	//	RoomType			: 1,
	//	ClubId				: 2256482,
	//	RoomId				: 223458,
	//	Index				: 2,
	//	GameStartTime		: time.Now().Unix(),
	//	GameEndTime			: time.Now().Unix(),
	//	PlayerInfos			: []*pbgamerecord.GamePlayerInfo{&pbgamerecord.GamePlayerInfo{UserId: 125458574,Name:"liwei1dao",BringinIntegral:10200,WinIntegral:160},
	//												{UserId:125458575,Name:"liwei2dao",BringinIntegral:10010,WinIntegral:-180},
	//												{UserId:125458576,Name:"liwei3dao",BringinIntegral:1500,WinIntegral:-260},
	//												{UserId:125458577,Name:"liwei4dao",BringinIntegral:11400,WinIntegral:380}},
	//	RePlayData			: []byte("测试复盘数据"),
	//}
	//WriteGameRecordReq(data)
	QueryRoomRecordReq();
}

func WriteGameRecordReq(GameRecord *pbgamerecord.WriteGameRecordReq)  {
	var cli client.XClient
	var err error
	ctx := context.Background()
	cli = cliCenter
	rsp := &codec.Message{}
	msg := &codec.Message{}
	err = codec.Pb2Msg(GameRecord, msg)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	err = cli.Call(ctx, "WriteGameRecordReq", msg, rsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Warn("执行测试脚本脚本失败")
		return
	}
}

//查询房间记录请求
func QueryRoomRecordReq()  {
	st, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-04-01 00:00:00", time.Local)
	et, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-05-01 00:00:00", time.Local)
	fmt.Println("查询时间显示",st.Unix(),et.Unix())
	req := &pbgamerecord.QueryRoomRecordReq{
		UserId:125458574,
		QueryType:1,
		QueryParam:125458574,
		QueryStartTime:st.Unix(),
		QueryEndTime:et.Unix(),
	}
	var cli client.XClient
	var err error
	ctx := context.Background()
	cli = cliCenter
	rsp := &codec.Message{}
	msg := &codec.Message{}
	err = codec.Pb2Msg(req, msg)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	err = cli.Call(ctx, "QueryRoomRecordReq", msg, rsp)
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Warn("执行测试脚本脚本失败")
		return
	}else{
		pb, err := codec.Msg2Pb(rsp)
		if err != nil {
			logrus.Error(err.Error())
			return
		}
		req, ok := pb.(*pbgamerecord.QueryRoomRecordRsp)
		if ok{
			fmt.Printf("查询结果 %v",req)
		}
	}
}


