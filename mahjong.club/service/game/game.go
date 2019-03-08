package game

import (
	"net"
	"time"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	"mahjong.club/club"
	"mahjong.club/config"
	"mahjong.club/core"
	fbs "mahjong.club/fbs/Common"
	"mahjong.club/hall"
	"mahjong.club/ierror"
	"mahjong.club/room"
)

// ReloadRoom 重载房间数据
func ReloadRoom(impacket *protocal.ImPacket) *ierror.Error {
	// 解析房间数据
	request := fbs.GetRootAsClubG2CReloadRoomPush(impacket.GetBody(), 0)
	fbsRoom := new(fbs.Room)
	fbsRoomInfo := new(fbs.RoomInfo)
	fbsRoomUser := new(fbs.RoomUserInfo)

	fbsRoom = request.Room(fbsRoom)
	fbsRoomInfo = fbsRoom.RoomInfo(fbsRoomInfo)

	// 新建一个房间
	roomId := int64(fbsRoomInfo.RoomId())

	r := room.NewRoom(roomId,
		string(fbsRoomInfo.Number()),
		int(fbsRoomInfo.GameType()),
		fbsRoomInfo.SettingBytes(),
		int(fbsRoom.Status()),
		int64(fbsRoom.CreateTime()),
		int64(fbsRoom.StartTime()),
		int(fbsRoomInfo.Round()),
		int(fbsRoomInfo.RandomRoom()), int(fbsRoom.CurrentRound()))
	if fbsRoom.RoomUsersLength() == 0 {
		r.H5Create = true
	}

	// 解析房间用户数据
	for i := 0; i < fbsRoom.RoomUsersLength(); i++ {
		fbsRoom.RoomUsers(fbsRoomUser, i)
		ru := room.NewUser(int(fbsRoomUser.UserId()),
			int(fbsRoomUser.Index()),
			string(fbsRoomUser.Avatar()),
			string(fbsRoomUser.Nickname()),
			int(fbsRoomUser.AvatarBox()))
		r.AddUser(ru)
	}

	// 通知俱乐部的用户重载房间
	clubId := int(request.ClubId())
	c, exists := hall.ClubSet.Get(clubId)
	if !exists {
		c = club.NewClub(clubId)
		hall.ClubSet.Add(c)
	}
	if roomExist := c.RoomSet.HasRoom(r.ID); !roomExist {
		// 开启房间的活跃检测
		go ListenRoomActive(clubId, r.ID)
	}
	// 俱乐部添加房间
	c.RoomSet.Add(r)

	// 给俱乐部的用户推送房间重载的消息
	hall.SendClubMessage(c, ClubReloadRoomPush(clubId, r))
	core.Logger.Info("[G2C.ReloadRoom]clubId:%v,roomId:%v", clubId, roomId)

	return nil
}

// JoinRoom 加入房间
func JoinRoom(impacket *protocal.ImPacket) *ierror.Error {
	// 解析用户数据
	fbsRoomUser := new(fbs.RoomUserInfo)
	request := fbs.GetRootAsClubG2CJoinRoomPush(impacket.GetBody(), 0)
	clubId := int(request.ClubId())
	roomId := int64(request.RoomId())
	fbsRoomUser = request.RoomUserInfo(fbsRoomUser)

	// 俱乐部是否存在
	c, exists := hall.ClubSet.Get(clubId)
	if !exists {
		return ierror.NewError(-10300, clubId)
	}
	// 俱乐部房间是否存在
	r, exists := c.RoomSet.Get(roomId)
	if !exists {
		return ierror.NewError(-10400, clubId, roomId)
	}
	userId := int(fbsRoomUser.UserId())
	ru := room.NewUser(userId, int(fbsRoomUser.Index()), string(fbsRoomUser.Avatar()), string(fbsRoomUser.Nickname()), int(fbsRoomUser.AvatarBox()))
	// 添加俱乐部房间用户
	r.AddUser(ru)
	// 给俱乐部的用户推送加入房间的消息
	hall.SendClubMessage(c, ClubJoinRoomPush(clubId, roomId, ru))
	core.Logger.Info("[G2C.JoinRoom]clubId:%v,roomId:%v,userId:%v", clubId, roomId, userId)

	return nil
}

// QuitRoom 退出房间
func QuitRoom(impacket *protocal.ImPacket) *ierror.Error {
	// 解析用户数据
	request := fbs.GetRootAsClubG2CQuitRoomPush(impacket.GetBody(), 0)
	clubId := int(request.ClubId())
	roomId := int64(request.RoomId())
	userId := int(request.UserId())
	index := int(request.Index())

	// 俱乐部是否存在
	c, exists := hall.ClubSet.Get(clubId)
	if !exists {
		return ierror.NewError(-10300, clubId)
	}
	// 俱乐部房间是否存在
	r, exists := c.RoomSet.Get(roomId)
	if !exists {
		return ierror.NewError(-10400, clubId, roomId)
	}

	// 从房间中删除用户
	r.DelUser(userId)
	// 给俱乐部的用户推送加入房间的消息
	hall.SendClubMessage(c, ClubQuitRoomPush(clubId, roomId, userId, index))
	core.Logger.Info("[G2C.QuitRoom]clubId:%v,roomId:%v,userId:%v", clubId, roomId, userId)

	return nil
}

// DismissRoom 解散房间
func DismissRoom(impacket *protocal.ImPacket) *ierror.Error {
	// 解析数据
	request := fbs.GetRootAsClubDismissRoomPush(impacket.GetBody(), 0)
	clubId := int(request.ClubId())
	roomId := int64(request.RoomId())
	code := int(request.Code())

	// 俱乐部是否存在
	c, exists := hall.ClubSet.Get(clubId)
	if !exists {
		return ierror.NewError(-10300, clubId)
	}
	// 将房间从俱乐部删除
	c.RoomSet.Del(roomId)
	// 给俱乐部的用户推送房间解散的消息
	hall.SendClubMessage(c, ClubDismissRoomPush(clubId, roomId, code))
	core.Logger.Info("[G2C.DismissRoom]clubId:%v,roomId:%v", clubId, roomId)

	return nil
}

// StartRoom 房间开始
func StartRoom(impacket *protocal.ImPacket) *ierror.Error {
	// 解析数据
	request := fbs.GetRootAsClubG2CStartRoomPush(impacket.GetBody(), 0)
	clubId := int(request.ClubId())
	roomId := int64(request.RoomId())
	round := int(request.Round())

	// 俱乐部是否存在
	c, exists := hall.ClubSet.Get(clubId)
	if !exists {
		return ierror.NewError(-10300, clubId)
	}
	// 俱乐部房间是否存在
	r, exists := c.RoomSet.Get(roomId)
	if !exists {
		return ierror.NewError(-10400, clubId, roomId)
	}
	// 设置房间为已开始
	r.Start(round)
	// 给俱乐部的用户推送房间开始的消息
	hall.SendClubMessage(c, ClubStartRoomPush(clubId, roomId, r.Status, r.CurrentRound))
	core.Logger.Info("[G2C.StartRoom]clubId:%v,roomId:%v,currentRound:%v", clubId, roomId, r.CurrentRound)

	return nil
}

// RoomActive 检测房间是否活跃
func RoomActive(conn *net.TCPConn, impacket *protocal.ImPacket) *ierror.Error {
	// 解析数据
	request := fbs.GetRootAsClubG2CStartRoomPush(impacket.GetBody(), 0)
	clubId := int(request.ClubId())
	roomId := int64(request.RoomId())

	// 房间是否活跃
	isActive := 0
	// 俱乐部是否存在
	c, exists := hall.ClubSet.Get(clubId)
	if exists {
		if r, exists := c.RoomSet.Get(roomId); exists {
			isActive = 1
			r.LastActiveTime = util.GetTime()
		}
	}
	ClubC2GRoomActiveResponse(clubId, roomId, isActive).Send(conn)
	core.Logger.Info("[G2C.RoomActive]clubId:%v,roomId:%v,isActive:%v", clubId, roomId, isActive)

	return nil
}

// ListenRoomActive 监听房间活跃时间
func ListenRoomActive(clubId int, roomId int64) {
	for {
		time.Sleep(time.Minute)
		// 超过2分钟未收到房间活跃的消息，则认为游戏服的房间已解散了
		// 俱乐部是否存在
		c, exists := hall.ClubSet.Get(clubId)
		if !exists {
			core.Logger.Info("[ListenRoomActive]俱乐部未发找到, clubId:%v,roomId:%v", clubId, roomId)
			return
		}
		r, exists := c.RoomSet.Get(roomId)
		if !exists {
			core.Logger.Info("[ListenRoomActive]俱乐部房间, clubId:%v,roomId:%v", clubId, roomId)
			return
		}
		if util.GetTime()-r.LastActiveTime > 120 {
			// 将房间从俱乐部删除
			c.RoomSet.Del(r.ID)
			hall.SendClubMessage(c, ClubDismissRoomPush(clubId, roomId, config.DISMISS_ROOM_CODE_TIMEOUT))
			core.Logger.Info("[ListenRoomActive]超时, clubId:%v,roomId:%v", clubId, roomId)
			return
		}
		core.Logger.Debug("[ListenRoomActive]正常, clubId:%v,roomId:%v", clubId, r.ID)
	}
}
