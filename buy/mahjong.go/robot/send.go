package robot

import (
	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	flatbuffers "github.com/google/flatbuffers/go"
	"mahjong.go/config"
	"mahjong.go/fbs/Common"
	"mahjong.go/mi/protocal"
)

// 机器人心跳
func (this *Robot) HeartBeat() {
	message, _ := simplejson.New().Encode()
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HEARTBEAT, message)
	this.Conn.Write(imPacket.Serialize())
	this.trace("发送心跳:%d", this.UserId)
}

// 机器人握手
func (this *Robot) HandShake() {
	// 握手
	user := make(map[string]interface{})
	user["token"] = util.GenToken(this.UserId, "latest", config.TOKEN_SECRET_KEY)
	user["lat"] = float64(-1)
	user["lng"] = float64(-1)
	js := simplejson.New()
	js.Set("user", user)

	message, _ := js.Encode()
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	this.Conn.Write(imPacket.Serialize())

	this.trace("handShake,userId:%d", this.UserId)
}

// 机器人握手成功
func (this *Robot) HandShakeAck() {
	message, _ := simplejson.New().Encode()
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, message)
	this.Conn.Write(imPacket.Serialize())

	this.trace("handShake,userId:%d", this.UserId)
}

// 机器人聊天
func (this *Robot) chat(chatId int16, content string) {
	var contentBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	if len(content) > 0 {
		contentBinary = builder.CreateString(content)
	}
	Common.RoomChatNotifyStart(builder)
	Common.RoomChatNotifyAddChatId(builder, chatId)
	Common.RoomChatNotifyAddContent(builder, contentBinary)
	orc := Common.RoomChatNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandRoomChatNotify), uint16(0), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())
}

// 比赛组队
func (this *Robot) MatchJoin() {
	// 随机加入房间
	mType := protocal.MSG_TYPE_REQUEST
	gType := this.MType //游戏类型
	builder := flatbuffers.NewBuilder(0)
	Common.MatchRoomRequestStart(builder)
	Common.MatchRoomRequestAddGameType(builder, uint16(gType))
	orc := Common.MatchRoomRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandMatchRoomRequest), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("MatchJoin,userId:%d", this.UserId)
}

// 随机组队
func (this *Robot) RandomJoin() {
	// 随机加入房间
	mType := protocal.MSG_TYPE_REQUEST
	gType := this.MType //游戏类型
	builder := flatbuffers.NewBuilder(0)
	Common.RandomRoomRequestStart(builder)
	Common.RandomRoomRequestAddGameType(builder, uint16(gType))
	orc := Common.RandomRoomRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandRandomRoomRequest), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("RandomJoin,userId:%d", this.UserId)
}

// RankJoin 排位赛组队
func (this *Robot) RankJoin() {
	// 随机加入房间
	mType := protocal.MSG_TYPE_REQUEST
	builder := flatbuffers.NewBuilder(0)
	Common.RankRoomRobotRequestStart(builder)
	Common.RankRoomRobotRequestAddRoomId(builder, uint64(this.RoomId))
	orc := Common.RankRoomRobotRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandRankRoomRobotRequest), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("RankJoin,userId:%d,roomId:%v", this.UserId, this.RoomId)
}

// LeagueJoin 联赛报名
func (this *Robot) LeagueJoin() {
	// 随机加入房间
	mType := protocal.MSG_TYPE_REQUEST
	builder := flatbuffers.NewBuilder(0)
	Common.LeagueRobotApplyRequestStart(builder)
	Common.LeagueRobotApplyRequestAddLeagueId(builder, int32(this.GameInfo.LeagueId))
	Common.LeagueRobotApplyRequestAddRaceId(builder, this.GameInfo.RaceId)
	orc := Common.LeagueRobotApplyRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandLeagueRobotApplyRequest), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("LeagueJoin,userId:%d,leagueId:%v,raceId:%v", this.UserId, this.GameInfo.LeagueId, this.GameInfo.RaceId)
}

// 金币组队
func (this *Robot) CoinJoin() {
	// 随机加入房间
	mType := protocal.MSG_TYPE_REQUEST
	gType := this.MType //游戏类型
	coinType := this.GameInfo.CoinType
	builder := flatbuffers.NewBuilder(0)
	Common.CoinRoomRequestStart(builder)
	Common.CoinRoomRequestAddCoinType(builder, byte(coinType))
	Common.CoinRoomRequestAddGameType(builder, uint16(gType))
	orc := Common.CoinRoomRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandCoinRoomRequest), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("CoinJoin,userId:%d", this.UserId)
}

// 机器人退出房间
func (this *Robot) QuitRoom() {
	mType := protocal.MSG_TYPE_NOTIFY
	builder := flatbuffers.NewBuilder(0)
	Common.QuitRoomNotifyStart(builder)
	orc := Common.QuitRoomNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandQuitRoomNotify), uint16(mType), uint16(0), uint16(0), buf)

	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("退出房间请求,userId:%d,roomId:%d", this.UserId, this.RoomId)
}
