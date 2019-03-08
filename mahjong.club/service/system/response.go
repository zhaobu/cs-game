package system

import (
	"github.com/fwhappy/mahjong/protocal"
	flatbuffers "github.com/google/flatbuffers/go"
	"mahjong.club/club"
	fbs "mahjong.club/fbs/Common"
	"mahjong.club/ierror"
	"mahjong.club/response"
)

// RoomListResponse 请求房间列表的回应
func RoomListResponse(c *club.Club) *protocal.ImPacket {
	var rooms flatbuffers.UOffsetT
	var roomList []flatbuffers.UOffsetT
	var c2iResult flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, c2iResult = response.BuidGameResult(builder, nil)
	builder, roomList = response.BuildRoomList(builder, c.RoomSet)
	fbs.ClubRestorePushStartRoomListVector(builder, len(roomList))
	for _, roomBinary := range roomList {
		builder.PrependUOffsetT(roomBinary)
	}
	rooms = builder.EndVector(len(roomList))
	fbs.ClubC2IRoomListResponseStart(builder)
	fbs.ClubC2IRoomListResponseAddC2iResult(builder, c2iResult)
	fbs.ClubC2IRoomListResponseAddClubId(builder, int32(c.ID))
	fbs.ClubC2IRoomListResponseAddRoomList(builder, rooms)
	orc := fbs.ClubC2IRoomListResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubC2IRoomListResponse,
		protocal.MSG_TYPE_RESPONSE,
		0, buf)
}

// RoomListErrorResponse 请求房间列表的出错回应
func RoomListErrorResponse(clubID int, err *ierror.Error) *protocal.ImPacket {
	var c2iResult flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, c2iResult = response.BuidGameResult(builder, err)
	fbs.ClubC2IRoomListResponseStart(builder)
	fbs.ClubC2IRoomListResponseAddC2iResult(builder, c2iResult)
	fbs.ClubC2IRoomListResponseAddClubId(builder, int32(clubID))
	orc := fbs.ClubC2IRoomListResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA,
		fbs.CommandClubC2IRoomListResponse,
		protocal.MSG_TYPE_RESPONSE,
		0, buf)
}
