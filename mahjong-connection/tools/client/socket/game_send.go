package main

import (
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"mahjong-connection/response"

	flatbuffers "github.com/google/flatbuffers/go"
)

// client 激活游戏请求
func c2sGameActivateRequest() {
	showClientDebug("c2sGameActivateRequest starting...")

	builder := flatbuffers.NewBuilder(0)
	roomNum := builder.CreateString("RECONNECT_ROOM")
	fbsCommon.GameActivateRequestStart(builder)
	fbsCommon.GameActivateRequestAddUserId(builder, uint32(0))
	fbsCommon.GameActivateRequestAddRoomNum(builder, roomNum)
	orc := fbsCommon.GameActivateRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameActivateRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sGameActivateRequest completed")
}

// client 加入房间请求
func c2sJoinRoomRequest(number string) {
	showClientDebug("c2sJoinRoomRequest starting...")

	builder := flatbuffers.NewBuilder(0)
	roomNum := builder.CreateString(number)
	fbsCommon.JoinRoomRequestStart(builder)
	fbsCommon.JoinRoomRequestAddNumber(builder, roomNum)
	orc := fbsCommon.JoinRoomRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandJoinRoomRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sJoinRoomRequest completed")
}
