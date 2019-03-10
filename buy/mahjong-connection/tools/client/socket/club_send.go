package main

import (
	"mahjong-connection/config"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"mahjong-connection/response"

	flatbuffers "github.com/google/flatbuffers/go"
)

// client发送加入俱乐部请求
func c2sClubJoinRequest(clubID int) {
	showClientDebug("c2sClubJoinRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubJoinRequestStart(builder)
	fbsCommon.ClubJoinRequestAddClubId(builder, int32(clubID))
	orc := fbsCommon.ClubJoinRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubJoinRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubJoinRequest completed")
}

// client发送退出俱乐部请求
func c2sClubQuitRequest(clubID int) {
	showClientDebug("c2sClubQuitRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubQuitRequestStart(builder)
	fbsCommon.ClubQuitRequestAddClubId(builder, int32(clubID))
	orc := fbsCommon.ClubQuitRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubQuitRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubQuitRequest completed")
}

// client发送俱乐部消息
func c2sClubClubMessageNotify(clubID int, mType int, content string) {
	showClientDebug("c2sClubClubMessageNotify starting...")
	builder := flatbuffers.NewBuilder(0)
	sendContent := builder.CreateString(content)
	fbsCommon.ClubClubMessageNotifyStart(builder)
	fbsCommon.ClubClubMessageNotifyAddClubId(builder, int32(clubID))
	fbsCommon.ClubClubMessageNotifyAddMType(builder, byte(mType))
	fbsCommon.ClubClubMessageNotifyAddContent(builder, sendContent)
	orc := fbsCommon.ClubClubMessageNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubClubMessageNotify, protocal.MSG_TYPE_NOTIFY, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubClubMessageNotify completed")
}

// client请求消息列表
func c2sClubClubMessageListNotify(clubID int, lastMsgId int, limit int) {
	showClientDebug("c2sClubClubMessageListNotify starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubClubMessageListNotifyStart(builder)
	fbsCommon.ClubClubMessageListNotifyAddClubId(builder, int32(clubID))
	fbsCommon.ClubClubMessageListNotifyAddMsgId(builder, uint64(lastMsgId))
	fbsCommon.ClubClubMessageListNotifyAddLimit(builder, byte(limit))
	orc := fbsCommon.ClubClubMessageListNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubClubMessageListNotify, protocal.MSG_TYPE_NOTIFY, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubClubMessageListNotify completed")
}

// client 请求俱乐部房间列表
func c2sClubI2CRoomListRequest(clubID int) {
	showClientDebug("c2sClubI2CRoomListRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	key := builder.CreateString(config.SYSTEM_KEY)
	fbsCommon.ClubI2CRoomListRequestStart(builder)
	fbsCommon.ClubI2CRoomListRequestAddClubId(builder, int32(clubID))
	fbsCommon.ClubI2CRoomListRequestAddSystemKey(builder, key)
	orc := fbsCommon.ClubI2CRoomListRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	impacket := response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubI2CRoomListRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf)
	impacket.Send(conn)
	showClientDebug("send c2sClubI2CRoomListRequest completed")
}

func c2sCloseClub() {
	showClientDebug("c2sCloseClub starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GatewayC2SCloseClubNotifyStart(builder)
	orc := fbsCommon.GatewayC2SCloseClubNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayC2SCloseClubNotify, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sCloseClub completed")
}
