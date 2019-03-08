package main

import (
	"mahjong-league/config"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/protocal"
	"mahjong-league/response"
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	flatbuffers "github.com/google/flatbuffers/go"
)

// 消息号码生成器
type numberGenerator struct {
	value int
	mux   *sync.Mutex
}

var (
	mg                *numberGenerator
	heartbeatInterval int
)

func init() {
	mg = &numberGenerator{mux: &sync.Mutex{}}
}

// 生成一个消息号
func (mg *numberGenerator) getNumber() uint16 {
	mg.mux.Lock()
	defer mg.mux.Unlock()
	mg.value++
	return uint16(mg.value)
}

// 客户端向服务端发送握手协议
func c2sHandShake(args ...string) {
	id = getParamsInt(0, args)
	if id == 0 {
		showClientError("用户id缺失")
		return
	}

	user := make(map[string]interface{})
	user["token"] = util.GenToken(id, "latest", config.TOKEN_SECRET_KEY)
	js := simplejson.New()
	js.Set("user", user)
	message, _ := js.Encode()

	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	conn.Write(imPacket.Serialize())

	showClientDebug("send handShake")
}

// client给server发送握手成功
func c2sHandShakeAck() {
	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, nil)
	conn.Write(imPacket.Serialize())
	showClientDebug("send handShakeAck")
}

// client每3秒给server发送一个心跳消息
// 服务端如果超过6秒没有收到包，则认为客户端已离线
func c2sHeartBeat() {
	for {
		time.Sleep(time.Duration(heartbeatInterval) * time.Second)
		response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT).Send(conn)
		// showClientDebug("send c2sHeartBeat")
	}
}

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
