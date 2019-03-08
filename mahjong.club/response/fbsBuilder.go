package response

import (
	flatbuffers "github.com/google/flatbuffers/go"
	fbs "mahjong.club/fbs/Common"
	"mahjong.club/ierror"
	"mahjong.club/room"
)

// BuidGameResult 构建一个fbs coommonResult
func BuidGameResult(builder *flatbuffers.Builder, err *ierror.Error) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	code := 0
	msg := ""
	if err != nil {
		code = err.GetCode()
		msg = err.Error()
	}
	errmsg := builder.CreateString(msg)
	fbs.GameResultStart(builder)
	fbs.GameResultAddCode(builder, int32(code))
	fbs.GameResultAddMsg(builder, errmsg)
	commonResult := fbs.GameResultEnd(builder)
	return builder, commonResult
}

// BuildRoomList 构建俱乐部房间组
func BuildRoomList(builder *flatbuffers.Builder, rs *room.Set) (*flatbuffers.Builder, []flatbuffers.UOffsetT) {
	var roomBinary flatbuffers.UOffsetT
	cnt := rs.Len()
	rooms := make([]flatbuffers.UOffsetT, 0, cnt)
	rs.Rooms.Range(func(k, v interface{}) bool {
		builder, roomBinary = BuildRoom(builder, v.(*room.Room))
		rooms = append(rooms, roomBinary)
		return true
	})
	return builder, rooms
}

// BuildRoom 构建房间信息
func BuildRoom(builder *flatbuffers.Builder, r *room.Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var userListBinary flatbuffers.UOffsetT
	var infoBinary flatbuffers.UOffsetT
	var roomBinary flatbuffers.UOffsetT
	builder, infoBinary = BuildRoomInfo(builder, r)
	builder, userListBinary = BuildRoomUserList(builder, r.GetUserList())

	fbs.RoomStart(builder)
	fbs.RoomAddStatus(builder, int8(r.Status))
	fbs.RoomAddCreateTime(builder, int64(r.CreateTime))
	fbs.RoomAddStartTime(builder, int64(r.StartTime))
	fbs.RoomAddCurrentRound(builder, byte(r.CurrentRound))
	fbs.RoomAddRoomInfo(builder, infoBinary)
	fbs.RoomAddRoomUsers(builder, userListBinary)
	roomBinary = fbs.RoomEnd(builder)
	return builder, roomBinary
}

// BuildRoomInfo 构建房间信息
func BuildRoomInfo(builder *flatbuffers.Builder, r *room.Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 生成setting的fbs结构
	var settingBinary flatbuffers.UOffsetT
	s := r.Setting.GetSetting()
	sLen := len(s)
	fbs.RoomInfoStartSettingVector(builder, sLen)
	for i := sLen - 1; i >= 0; i-- {
		builder.PrependByte(uint8(s[i]))
	}
	settingBinary = builder.EndVector(sLen)

	number := builder.CreateString(r.Number)

	fbs.RoomInfoStart(builder)
	fbs.RoomInfoAddRoomId(builder, uint64(r.ID))
	fbs.RoomInfoAddGameType(builder, uint16(r.GameType))
	fbs.RoomInfoAddRound(builder, uint8(r.Round))
	fbs.RoomInfoAddNumber(builder, number)
	fbs.RoomInfoAddPlayerCount(builder, uint8(r.Setting.GetSettingPlayerCnt()))
	fbs.RoomInfoAddSetting(builder, settingBinary)
	fbs.RoomInfoAddRandomRoom(builder, uint8(r.CType))
	roomInfo := fbs.RoomInfoEnd(builder)
	return builder, roomInfo
}

// BuildRoomUserInfo 生成房间用户信息
func BuildRoomUserInfo(builder *flatbuffers.Builder, ru *room.User) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 生成对应的字符串
	nickname := builder.CreateString(ru.Nickname)
	avatar := builder.CreateString(ru.Avatar)

	fbs.RoomUserInfoStart(builder)
	fbs.RoomUserInfoAddUserId(builder, uint32(ru.ID))
	fbs.RoomUserInfoAddIndex(builder, byte(ru.Index))
	fbs.RoomUserInfoAddAvatar(builder, avatar)
	fbs.RoomUserInfoAddNickname(builder, nickname)
	// index=0的就是创建者
	if ru.Index == 0 {
		fbs.RoomUserInfoAddIsHost(builder, byte(1))
	} else {
		fbs.RoomUserInfoAddIsHost(builder, byte(0))
	}
	fbs.RoomUserInfoAddAvatarBox(builder, int32(ru.AvatarBox))
	roomUserInfo := fbs.RoomUserInfoEnd(builder)
	return builder, roomUserInfo
}

// BuildRoomUserList 房间用户列表
func BuildRoomUserList(builder *flatbuffers.Builder, rus map[int]*room.User) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var listBinary flatbuffers.UOffsetT
	var infoBinary flatbuffers.UOffsetT

	cnt := len(rus)
	users := make([]flatbuffers.UOffsetT, 0, cnt)
	for _, ru := range rus {
		builder, infoBinary = BuildRoomUserInfo(builder, ru)
		users = append(users, infoBinary)
	}

	fbs.RoomStartRoomUsersVector(builder, cnt)
	for _, infoBinary := range users {
		builder.PrependUOffsetT(infoBinary)
	}
	listBinary = builder.EndVector(cnt)
	return builder, listBinary
}
