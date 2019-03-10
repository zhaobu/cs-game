package main

import (
	"sync"
	"time"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	"mahjong.club/config"
	"mahjong.club/response"

	simplejson "github.com/bitly/go-simplejson"
	flatbuffers "github.com/google/flatbuffers/go"
	fbs "mahjong.club/fbs/Common"
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
func (mg *numberGenerator) getNumber() uint32 {
	mg.mux.Lock()
	defer mg.mux.Unlock()
	mg.value++
	return uint32(mg.value)
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

// client发送加入俱乐部请求
func c2sClubJoinRequest(clubID int) {
	showClientDebug("c2sClubJoinRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubJoinRequestStart(builder)
	fbs.ClubJoinRequestAddClubId(builder, int32(clubID))
	orc := fbs.ClubJoinRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubJoinRequest, protocal.MSG_TYPE_REQUEST, mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubJoinRequest completed")
}

// client发送退出俱乐部请求
func c2sClubQuitRequest(clubID int) {
	showClientDebug("c2sClubQuitRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubQuitRequestStart(builder)
	fbs.ClubQuitRequestAddClubId(builder, int32(clubID))
	orc := fbs.ClubQuitRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubQuitRequest, protocal.MSG_TYPE_REQUEST, mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubQuitRequest completed")
}

// client发送俱乐部消息
func c2sClubClubMessageNotify(clubID int, mType int, content string) {
	showClientDebug("c2sClubClubMessageNotify starting...")
	builder := flatbuffers.NewBuilder(0)
	sendContent := builder.CreateString(content)
	fbs.ClubClubMessageNotifyStart(builder)
	fbs.ClubClubMessageNotifyAddClubId(builder, int32(clubID))
	fbs.ClubClubMessageNotifyAddMType(builder, byte(mType))
	fbs.ClubClubMessageNotifyAddContent(builder, sendContent)
	orc := fbs.ClubClubMessageNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubClubMessageNotify, protocal.MSG_TYPE_NOTIFY, mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubClubMessageNotify completed")
}

// client请求消息列表
func c2sClubClubMessageListNotify(clubID int, lastMsgId int, limit int) {
	showClientDebug("c2sClubClubMessageListNotify starting...")
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubClubMessageListNotifyStart(builder)
	fbs.ClubClubMessageListNotifyAddClubId(builder, int32(clubID))
	fbs.ClubClubMessageListNotifyAddMsgId(builder, uint64(lastMsgId))
	fbs.ClubClubMessageListNotifyAddLimit(builder, byte(limit))
	orc := fbs.ClubClubMessageListNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubClubMessageListNotify, protocal.MSG_TYPE_NOTIFY, mg.getNumber(), buf).Send(conn)
	showClientDebug("send c2sClubClubMessageListNotify completed")
}

// client 请求俱乐部房间列表
func c2sClubI2CRoomListRequest(clubID int) {
	showClientDebug("c2sClubI2CRoomListRequest starting...")
	builder := flatbuffers.NewBuilder(0)
	key := builder.CreateString(config.SYSTEM_KEY)
	fbs.ClubI2CRoomListRequestStart(builder)
	fbs.ClubI2CRoomListRequestAddClubId(builder, int32(clubID))
	fbs.ClubI2CRoomListRequestAddSystemKey(builder, key)
	orc := fbs.ClubI2CRoomListRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	impacket := response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubI2CRoomListRequest, protocal.MSG_TYPE_REQUEST, mg.getNumber(), buf)
	impacket.Send(conn)
	showClientDebug("send c2sClubI2CRoomListRequest completed")
}

/*
// C2sClubRestore 客户端请求重载俱乐部房间
func C2sClubRestore(clubID int) {
	showClientDebug("C2sClubRestore starting...")
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubRestoreNotifyStart(builder)
	fbs.ClubRestoreNotifyAddClubId(builder, int32(clubID))
	orc := fbs.ClubRestoreNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubRestoreNotify, protocal.MSG_TYPE_NOTIFY, uint32(0), buf).Send(conn)

	showClientDebug("send C2sClubRestore completed")
}

// C2sClubRestoreDone 客户端回应重载完成
func C2sClubRestoreDone() {
	showClientDebug("C2sClubRestoreDone starting...")
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubRestoreDoneNotifyStart(builder)
	orc := fbs.ClubRestoreDoneNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbs.CommandClubRestoreDoneNotify, protocal.MSG_TYPE_NOTIFY, uint32(0), buf).Send(conn)

	showClientDebug("send C2sClubRestoreDone completed")
}
*/
