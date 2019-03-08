package service

import (
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"mahjong-connection/response"
	"mahjong-connection/selectserver"

	flatbuffers "github.com/google/flatbuffers/go"
)

// 构建一个fbs coommonresult
func genGameResult(builder *flatbuffers.Builder, code int, msg string) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	errmsg := builder.CreateString(msg)
	fbsCommon.GameResultStart(builder)
	fbsCommon.GameResultAddCode(builder, int32(code))
	fbsCommon.GameResultAddMsg(builder, errmsg)
	commonResult := fbsCommon.GameResultEnd(builder)

	return builder, commonResult
}

// GameActivateResponse 用户激活协议回调
func GameActivateResponse(result *selectserver.Result, mNumber uint16) *protocal.ImPacket {
	var gameResult flatbuffers.UOffsetT

	// code 始终给0
	// code < 0 时， 放到roomId中
	if result.Code != 0 {
		result.RoomId = int64(result.Code)
		result.Code = 0
	}

	builder := flatbuffers.NewBuilder(0)
	builder, gameResult = genGameResult(builder, result.Code, "")
	fbsCommon.GameActivateResponseStart(builder)
	fbsCommon.GameActivateResponseAddS2cResult(builder, gameResult)
	fbsCommon.GameActivateResponseAddRaceId(builder, result.RaceId)
	fbsCommon.GameActivateResponseAddRoomId(builder, result.RoomId)
	orc := fbsCommon.GameActivateResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameActivateResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// GameRestoreNotify 构建一个游戏服重连的请求
func GameRestoreNotify() *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GameRestoreNotifyStart(builder)
	fbsCommon.GameRestoreNotifyAddRoomId(builder, uint64(0))
	fbsCommon.GameRestoreNotifyAddRound(builder, uint16(0))
	fbsCommon.GameRestoreNotifyAddStep(builder, uint16(0))
	orc := fbsCommon.GameRestoreNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameRestoreNotify, protocal.MSG_TYPE_NOTIFY, uint16(0), uint16(0), buf)
}

// CloseGamePush 游戏连接关闭的推送
func CloseGamePush(code int, msg string) *protocal.ImPacket {
	var gameResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, gameResult = genGameResult(builder, code, msg)
	fbsCommon.GatewayS2CCloseGamePushStart(builder)
	fbsCommon.GatewayS2CCloseGamePushAddS2cResult(builder, gameResult)
	orc := fbsCommon.GatewayS2CCloseGamePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayS2CCloseGamePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// CloseClubPush 俱乐部连接关闭的推送
func CloseClubPush() *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GatewayS2CCloseClubPushStart(builder)
	orc := fbsCommon.GatewayS2CCloseClubPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayS2CCloseClubPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// CloseLeaguePush 联赛大厅连接关闭的推送
func CloseLeaguePush() *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GatewayS2CCloseLeaguePushStart(builder)
	orc := fbsCommon.GatewayS2CCloseLeaguePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayS2CCloseLeaguePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}
