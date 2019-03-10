package system

import (
	"net"

	"github.com/fwhappy/mahjong/protocal"
	"mahjong.club/config"
	"mahjong.club/core"
	fbsCommon "mahjong.club/fbs/Common"
	"mahjong.club/hall"
	"mahjong.club/ierror"
)

// RoomList 读取俱乐部房间列表
func RoomList(conn *net.TCPConn, impacket *protocal.ImPacket) (int, *ierror.Error) {
	core.Logger.Debug("[RoomList]body:%v,", impacket.GetBody())
	request := fbsCommon.GetRootAsClubI2CRoomListRequest(impacket.GetBody(), 0)
	clubID := int(request.ClubId())
	if err := verifySystemKey(string(request.SystemKey())); err != nil {
		return clubID, err
	}
	if clubID == 0 {
		return clubID, ierror.NewError(-10101, "club.RoomListAction", "clubID")
	}
	// 读取俱乐部信息
	c, isExists := hall.ClubSet.Get(clubID)
	if !isExists {
		return clubID, ierror.NewError(-10300, clubID)
	}
	RoomListResponse(c).Send(conn)
	core.Logger.Info("[system.RoomList]clubID:%v,remote:%v", clubID, conn.RemoteAddr().String())
	return clubID, nil
}

// 验证系统密钥
func verifySystemKey(key string) *ierror.Error {
	if key != config.SYSTEM_KEY {
		return ierror.NewError(-2, key)
	}
	return nil
}
