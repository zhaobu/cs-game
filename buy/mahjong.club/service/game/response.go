package game

import (
	flatbuffers "github.com/google/flatbuffers/go"
	fbs "mahjong.club/fbs/Common"
	"github.com/fwhappy/mahjong/protocal"
	"mahjong.club/response"
	"mahjong.club/room"
)

// ClubReloadRoomPush 重载房间消息
func ClubReloadRoomPush(clubId int, r *room.Room) *protocal.ImPacket {
	var roomBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, roomBinary = response.BuildRoom(builder, r)
	fbs.ClubReloadRoomPushStart(builder)
	fbs.ClubReloadRoomPushAddClubId(builder, int32(clubId))
	fbs.ClubReloadRoomPushAddRoom(builder, roomBinary)
	orc := fbs.ClubReloadRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubReloadRoomPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubJoinRoomPush 用户加入房间
func ClubJoinRoomPush(clubId int, roomId int64, ru *room.User) *protocal.ImPacket {
	var roomUserBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, roomUserBinary = response.BuildRoomUserInfo(builder, ru)
	fbs.ClubJoinRoomPushStart(builder)
	fbs.ClubJoinRoomPushAddClubId(builder, int32(clubId))
	fbs.ClubJoinRoomPushAddRoomId(builder, uint64(roomId))
	fbs.ClubJoinRoomPushAddRoomUserInfo(builder, roomUserBinary)
	orc := fbs.ClubJoinRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubJoinRoomPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubQuitRoomPush 用户退出房间
func ClubQuitRoomPush(clubId int, roomId int64, userId, index int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubQuitRoomPushStart(builder)
	fbs.ClubQuitRoomPushAddClubId(builder, int32(clubId))
	fbs.ClubQuitRoomPushAddRoomId(builder, uint64(roomId))
	fbs.ClubQuitRoomPushAddUserId(builder, uint32(userId))
	fbs.ClubQuitRoomPushAddIndex(builder, uint8(index))
	orc := fbs.ClubQuitRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubQuitRoomPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubDismissRoomPush 用户退出房间
func ClubDismissRoomPush(clubId int, roomId int64, code int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubDismissRoomPushStart(builder)
	fbs.ClubDismissRoomPushAddClubId(builder, int32(clubId))
	fbs.ClubDismissRoomPushAddRoomId(builder, uint64(roomId))
	fbs.ClubDismissRoomPushAddCode(builder, int32(code))
	orc := fbs.ClubDismissRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubDismissRoomPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubStartRoomPush 开始房间
func ClubStartRoomPush(clubId int, roomId int64, status int, currentRound int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubRoomStatusPushStart(builder)
	fbs.ClubRoomStatusPushAddClubId(builder, int32(clubId))
	fbs.ClubRoomStatusPushAddRoomId(builder, uint64(roomId))
	fbs.ClubRoomStatusPushAddStatus(builder, int8(status))
	fbs.ClubRoomStatusPushAddCurrentRound(builder, byte(currentRound))
	orc := fbs.ClubRoomStatusPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubRoomStatusPush,
		protocal.MSG_TYPE_PUSH,
		uint32(0), buf)
}

// ClubC2GRoomActivePesponse 返回房间活跃状态
func ClubC2GRoomActiveResponse(clubId int, roomId int64, active int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbs.ClubC2GRoomActivePesponseStart(builder)
	fbs.ClubC2GRoomActivePesponseAddClubId(builder, int32(clubId))
	fbs.ClubC2GRoomActivePesponseAddRoomId(builder, uint64(roomId))
	fbs.ClubC2GRoomActivePesponseAddActive(builder, uint8(active))
	orc := fbs.ClubRoomStatusPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(
		protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubC2GRoomActiveResponse,
		protocal.MSG_TYPE_RESPONSE,
		uint32(0), buf)
}
