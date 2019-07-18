package main

import (
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	impb "cy/other/im/pb"
	friendpb "cy/other/im/pb/friend"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func printLoginReq() {
	req := &impb.LoginReq{}

	pay := codec.NewMsgPayload()
	pay.FromUID = 10001
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printSendMsgReq() {
	req := &impb.SendMsgReq{}
	req.Seq = 1
	req.To = 200
	req.From = 10001
	req.Content = []byte(`are you ok???`)

	pay := codec.NewMsgPayload()
	pay.Seq = 0
	pay.FromUID = req.From
	pay.ToUID = req.To
	pay.Flag.SetMultiCast(true)
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printMsgRecordReq() {
	req := &impb.MsgRecordReq{}
	req.Seq = 1
	req.To = 10002
	req.From = 10001
	req.StartMsgID = "1539602243852000001"
	req.Limit = 3

	pay := codec.NewMsgPayload()
	pay.Seq = 0
	pay.FromUID = req.From
	pay.ToUID = req.To
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printAddFriendReq() {
	req := &friendpb.AddFriendReq{}
	req.Source = 10001
	req.Target = 10002
	req.Msg = "hello "

	pay := codec.NewMsgPayload()
	pay.Seq = 0
	pay.FromUID = req.Source
	pay.ToUID = req.Target
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printQueryFriendReq() {
	req := &friendpb.QueryFriendReq{}
	req.Seq = 3

	pay := codec.NewMsgPayload()
	pay.FromUID = 20001

	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printAddFriendNotifAck() {
	req := &friendpb.AddFriendNotifAck{}
	req.Source = 32234
	req.Target = 10001
	//req.Msg = "good"
	req.Code = 1

	pay := codec.NewMsgPayload()
	pay.FromUID = 10001
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printQueryUnreadCntReq() {
	req := &impb.QueryUnreadCntReq{}
	req.UID = 20001

	pay := codec.NewMsgPayload()
	pay.FromUID = 20001
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printQueryUnreadNReq() {
	req := &impb.QueryUnreadNReq{}
	req.UID = 20001
	req.LastN = 11000
	req.OtherUID = 20005

	pay := codec.NewMsgPayload()
	pay.FromUID = 20001
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printMsgNotifyAck() {
	req := &impb.MsgNotifyAck{}
	req.MsgIDs = []string{"1540382616826000000"}

	pay := codec.NewMsgPayload()
	pay.FromUID = 20001
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printEnterExitRoom() {
	req := &impb.EnterExitRoom{}
	req.UID = 10002
	req.RoomID = 200
	req.EnterOrExit = 1

	pay := codec.NewMsgPayload()
	pay.FromUID = req.UID
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func printQueryInbox() {
	req := &friendpb.QueryInbox{}
	req.MyID = 67529
	req.StartMsgID = "0"
	req.Limit = 10

	pay := codec.NewMsgPayload()
	pay.FromUID = 67529
	pay.Seq = 0
	var err error
	pay.PayloadName, pay.Payload, err = protobuf.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}

	payjson, _ := json.Marshal(pay)
	fmt.Println(string(payjson))
}

func unmarsh() {
	data, _ := base64.StdEncoding.DecodeString("ChMxNTQyMTg4NzkwMzYwMDEyMjAwEMgBGJFOIgthcmUgeW91IG9rPyio6/2Pi5G9sxU=")

	pb, err := protobuf.Unmarshal("impb.MsgNotify", data)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	xx := pb.(*impb.MsgNotify)
	fmt.Printf("%+v\n", xx)

}
