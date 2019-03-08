package club

import (
	"github.com/fwhappy/mahjong/protocal"
	"mahjong.club/club"
	"mahjong.club/ierror"
	"mahjong.club/message"
	"mahjong.club/response"

	flatbuffers "github.com/google/flatbuffers/go"
	fbs "mahjong.club/fbs/Common"
)

// JoinResponse 加入俱乐部
func JoinResponse(mNumber uint32, clubID int, err *ierror.Error) *protocal.ImPacket {
	var s2cResult flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, err)
	fbs.ClubJoinResponseStart(builder)
	fbs.ClubJoinResponseAddS2cResult(builder, s2cResult)
	fbs.ClubJoinResponseAddClubId(builder, int32(clubID))
	orc := fbs.ClubJoinResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubJoinResponse,
		protocal.MSG_TYPE_RESPONSE,
		mNumber, buf)
}

// QuitResponse 退出俱乐部
func QuitResponse(mNumber uint32, clubID int, err *ierror.Error) *protocal.ImPacket {
	var s2cResult flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, err)
	fbs.ClubQuitResponseStart(builder)
	fbs.ClubQuitResponseAddS2cResult(builder, s2cResult)
	fbs.ClubQuitResponseAddClubId(builder, int32(clubID))
	orc := fbs.ClubQuitResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubQuitResponse,
		protocal.MSG_TYPE_RESPONSE,
		mNumber, buf)
}

// ClubRestorePush 俱乐部重连
func ClubRestorePush(c *club.Club) *protocal.ImPacket {
	var rooms flatbuffers.UOffsetT
	var roomList []flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)

	builder, roomList = response.BuildRoomList(builder, c.RoomSet)
	fbs.ClubRestorePushStartRoomListVector(builder, len(roomList))
	for _, roomBinary := range roomList {
		builder.PrependUOffsetT(roomBinary)
	}
	rooms = builder.EndVector(len(roomList))

	fbs.ClubRestorePushStart(builder)
	fbs.ClubRestorePushAddClubId(builder, int32(c.ID))
	fbs.ClubRestorePushAddRoomList(builder, rooms)
	orc := fbs.ClubRestorePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubRestorePush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

func genMsgSender(builder *flatbuffers.Builder, ms *message.MsgSender) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var msgSenderBinary flatbuffers.UOffsetT
	fbs.MsgSenderStart(builder)
	fbs.MsgSenderAddUserId(builder, uint32(ms.Info.ID))
	fbs.MsgSenderAddAvatarBox(builder, int32(ms.Info.AvatarBox))
	msgSenderBinary = fbs.MsgSenderEnd(builder)
	return builder, msgSenderBinary
}

func genMsg(builder *flatbuffers.Builder, m *message.Msg) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var msgBinary flatbuffers.UOffsetT
	var content flatbuffers.UOffsetT
	var msgSenderBinary flatbuffers.UOffsetT
	content = builder.CreateString(m.Content)
	builder, msgSenderBinary = genMsgSender(builder, m.Sender)
	fbs.MsgStart(builder)
	fbs.MsgAddMId(builder, m.MID)
	fbs.MsgAddMType(builder, byte(m.MType))
	fbs.MsgAddContent(builder, content)
	fbs.MsgAddSender(builder, msgSenderBinary)
	fbs.MsgAddT(builder, int64(m.CreateTime))
	msgBinary = fbs.MsgEnd(builder)
	return builder, msgBinary
}

func genMsgList(builder *flatbuffers.Builder, msgList []*message.Msg) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var msgListBinary flatbuffers.UOffsetT
	var msgBinary flatbuffers.UOffsetT
	mLen := len(msgList)
	msgBinaryList := make([]flatbuffers.UOffsetT, 0, mLen)
	for _, m := range msgList {
		builder, msgBinary = genMsg(builder, m)
		msgBinaryList = append(msgBinaryList, msgBinary)
	}
	fbs.ClubClubMessageListPushStartMsgListVector(builder, mLen)
	for _, msgBinary = range msgBinaryList {
		builder.PrependUOffsetT(msgBinary)
	}
	msgListBinary = builder.EndVector(mLen)
	return builder, msgListBinary
}

// ClubMessagePush 推送单条俱乐部消息
func ClubMessagePush(clubID int, m *message.Msg) *protocal.ImPacket {
	var msgBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, msgBinary = genMsg(builder, m)
	fbs.ClubClubMessagePushStart(builder)
	fbs.ClubClubMessagePushAddClubId(builder, int32(clubID))
	fbs.ClubClubMessagePushAddMsg(builder, msgBinary)
	orc := fbs.ClubClubMessagePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubClubMessagePush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubMessageListPush 推送俱乐部历史消息
func ClubMessageListPush(clubID int, msgList []*message.Msg) *protocal.ImPacket {
	var msgListBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, msgListBinary = genMsgList(builder, msgList)
	fbs.ClubClubMessageListPushStart(builder)
	fbs.ClubClubMessageListPushAddClubId(builder, int32(clubID))
	fbs.ClubClubMessageListPushAddMsgList(builder, msgListBinary)
	orc := fbs.ClubClubMessageListPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubClubMessageListPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}
