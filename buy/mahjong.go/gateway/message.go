package gateway

import (
	flatbuffers "github.com/google/flatbuffers/go"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/response"
	"mahjong.go/mi/protocal"
)

// 构建消息
func buildMessage(builder *flatbuffers.Builder, messageId int, content, avatar string, avatarBox int) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	s := builder.CreateString(content)
	url := builder.CreateString(avatar)
	fbsCommon.S2cMessageStart(builder)
	fbsCommon.S2cMessageAddMessageId(builder, uint32(messageId))
	fbsCommon.S2cMessageAddContent(builder, s)
	fbsCommon.S2cMessageAddUrl(builder, url)
	fbsCommon.S2cMessageAddAvatarBox(builder, int32(avatarBox))
	v := fbsCommon.RaceEnd(builder)
	return builder, v
}

// PrivateMessagePush 消息推送
func PrivateMessagePush(userId int, messageId int, content, avatar string, avatarBox int) *protocal.ImPacket {
	var message flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, message = buildMessage(builder, messageId, content, avatar, avatarBox)

	fbsCommon.GatewayS2CPrivateMessagePushStart(builder)
	fbsCommon.GatewayS2CPrivateMessagePushAddUserId(builder, uint32(userId))
	fbsCommon.GatewayS2CPrivateMessagePushAddS2cMessage(builder, message)
	orc := fbsCommon.GatewayS2CPrivateMessagePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayS2CPrivateMessagePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}
