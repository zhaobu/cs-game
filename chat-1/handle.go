package main

import (
	. "cy/chat/def"
	"encoding/json"
	"fmt"
)

type Handle func(req interface{}, cli *client)

func handleQueryGroupReq(req interface{}, cli *client) {
	v, ok := req.(*QueryGroupReq)
	if !ok {
		fmt.Println("bad type", req)
		return
	}

	fmt.Printf("handleQueryGroupReq %+v\n", v)

	rsp := QueryGroupRsp{Kind: OpQueryGroupRsp}
	rsp.Infos = listGroup()
	cli.sends(rsp)
}

func handleJoinGroupReq(req interface{}, cli *client) {
	v, ok := req.(*JoinGroupReq)
	if !ok {
		fmt.Println("bad type", req)
		return
	}

	fmt.Printf("handleJoinGroupReq %+v\n", v)

	err := joinGroup(cli, v.GroupName)
	rsp := JoinGroupRsp{Kind: OpJoinGroupRsp}
	if err != nil {
		rsp.Msg = fmt.Sprint(err)
	} else {
		rsp.Msg = "ok"
	}
	cli.sends(rsp)
}

func handleExitGroupReq(req interface{}, cli *client) {
	v, ok := req.(*ExitGroupReq)
	if !ok {
		fmt.Println("bad type", req)
		return
	}

	fmt.Printf("handleExitGroupReq %+v\n", v)
	err := exitGroup(cli, v.GroupName)
	rsp := ExitGroupRsp{Kind: OpExitGroupRsp}
	if err != nil {
		rsp.Msg = fmt.Sprint(err)
	} else {
		rsp.Msg = "ok"
	}
	cli.sends(rsp)
}

func handleSendGroupMsgReq(req interface{}, cli *client) {
	v, ok := req.(*SendGroupMsgReq)
	if !ok {
		fmt.Println("bad type", req)
		return
	}

	//fmt.Printf("handleSendGroupMsgReq %+v\n", v)

	rsp := SendGroupMsgRsp{}
	defer func() {
		cli.sends(rsp)
	}()

	ok, err := inGroup(cli.id, v.ToGroup)
	if err != nil {
		rsp.Msg = err.Error()
		return
	}
	if !ok {
		rsp.Msg = fmt.Sprintf("not in group: %s", v.ToGroup)
		return
	}

	nif := MsgNotify{Kind: OpMsgNotify}
	nif.Seq = v.Seq
	nif.From = cli.id
	nif.To = v.ToGroup
	nif.Content = v.Content
	data, err := json.Marshal(nif)
	if err != nil {
		fmt.Println(err)
		return
	}
	sendMsgToGroup(data, v.ToGroup)
}
