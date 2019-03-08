package main

import (
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"mahjong-connection/response"

	flatbuffers "github.com/google/flatbuffers/go"
)

// 客户端向服务器拉取大厅列表
func c2sLeaguelistRequest() {
	showClientDebug("c2sLeaguelistRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueListRequestStart(builder)
	orc := fbsCommon.LeagueListRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueListRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sLeaguelistRequest completed")
}

// 报名比赛
func c2sLeagueApplyRequest(leagueId int) {
	showClientDebug("c2sLeagueApplyRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueApplyRequestStart(builder)
	fbsCommon.LeagueApplyRequestAddLeagueId(builder, int32(leagueId))
	orc := fbsCommon.LeagueApplyRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueApplyRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sLeagueApplyRequest completed")
}

// 取消报名比赛
func c2sLeagueCancelRequest() {
	showClientDebug("c2sLeagueCancelRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueCancelRequestStart(builder)
	orc := fbsCommon.LeagueCancelRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueCancelRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sLeagueCancelRequest completed")
}

// 退赛
func c2sLeagueQuitRequest() {
	showClientDebug("c2sLeagueQuitRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueCancelRequestStart(builder)
	orc := fbsCommon.LeagueCancelRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueQuitRequest, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sLeagueQuitRequest completed")
}

// 收到比赛结果
func c2sLeagueRaceResultReceivedNotify() {
	showClientDebug("c2sLeagueRaceResultReceivedNotify starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueRaceResultRecievedNotifyStart(builder)
	orc := fbsCommon.LeagueRaceResultRecievedNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRaceResultReceivedNotify, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sLeagueRaceResultReceivedNotify completed")
}

func c2sCloseLeague() {
	showClientDebug("c2sCloseLeague starting...")
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GatewayC2SCloseLeagueNotifyStart(builder)
	orc := fbsCommon.GatewayC2SCloseLeagueNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayC2SCloseLeagueNotify, protocal.MSG_TYPE_REQUEST, uint16(0), mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sCloseLeague completed")
}
