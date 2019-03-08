package main

import (
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"

	"github.com/fwhappy/util"
)

// 收到加入俱乐部请求
func s2cClubJoinResponse(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubJoinResponse(impacket.GetBody(), 0)
	s2cResult := new(fbsCommon.GameResult)
	response.S2cResult(s2cResult)
	if s2cResult.Code() < 0 {
		showClientError("[s2cClubJoinResponse],code:%v, msg:%v", s2cResult.Code(), string(s2cResult.Msg()))
	} else {
		showClientDebug("加入俱乐部成功:%v", response.ClubId())
	}
}

// 收到退出俱乐部请求
func s2cClubQuitResponse(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubQuitResponse(impacket.GetBody(), 0)
	s2cResult := new(fbsCommon.GameResult)
	response.S2cResult(s2cResult)
	if s2cResult.Code() < 0 {
		showClientError("[s2cClubQuitResponse],code:%v, msg:%v", s2cResult.Code(), string(s2cResult.Msg()))
	} else {
		showClientDebug("退出俱乐部成功:%v", response.ClubId())
	}
}

func sc2ClubRestorePush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubRestorePush(impacket.GetBody(), 0)
	showClientDebug("重新载入俱乐部房间列表,clubId:%v", response.ClubId())
	room := new(fbsCommon.Room)
	roomInfo := new(fbsCommon.RoomInfo)
	roomUserInfo := new(fbsCommon.RoomUserInfo)
	for i := 0; i < response.RoomListLength(); i++ {
		showClientDebug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		response.RoomList(room, i)
		roomInfo = room.RoomInfo(roomInfo)
		showClientDebug("房间信息,roomId:%v,number:%v,gType:%v,playerCount:%v,status:%v,currentRound:%v,round:%v, setting:%v",
			roomInfo.RoomId(), string(roomInfo.Number()), roomInfo.GameType(), roomInfo.PlayerCount(), room.Status(),
			room.CurrentRound(), roomInfo.Round(), roomInfo.SettingBytes())

		for j := 0; j < room.RoomUsersLength(); j++ {
			room.RoomUsers(roomUserInfo, j)
			showClientDebug("房间用户信息:userId:%v,index:%v,nickname:%v,avatar:%v",
				roomUserInfo.UserId(), roomUserInfo.Index(), string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))
		}
		showClientDebug("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	}
}

func s2cClubReloadRoomPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubReloadRoomPush(impacket.GetBody(), 0)
	showClientDebug("重新载入房间,clubId:%v", response.ClubId())
	room := new(fbsCommon.Room)
	room = response.Room(room)
	roomInfo := new(fbsCommon.RoomInfo)
	roomInfo = room.RoomInfo(roomInfo)
	showClientDebug("房间信息,roomId:%v,number:%v,gType:%v,playerCount:%v,status:%v,currentRound:%v,round:%v, setting:%v",
		roomInfo.RoomId(), string(roomInfo.Number()), roomInfo.GameType(), roomInfo.PlayerCount(), room.Status(),
		room.CurrentRound(), roomInfo.Round(), roomInfo.SettingBytes())

	roomUserInfo := new(fbsCommon.RoomUserInfo)
	for i := 0; i < room.RoomUsersLength(); i++ {
		room.RoomUsers(roomUserInfo, i)
		showClientDebug("房间用户信息:userId:%v,index:%v,nickname:%v,avatar:%v",
			roomUserInfo.UserId(), roomUserInfo.Index(), string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))
	}

}

func s2cClubJoinRoomPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubJoinRoomPush(impacket.GetBody(), 0)
	roomUserInfo := new(fbsCommon.RoomUserInfo)
	roomUserInfo = response.RoomUserInfo(roomUserInfo)
	showClientDebug("加入房间, clubId:%v, roomId:%v, userId:%v, index:%v,nickname:%v,avatar:%v",
		response.ClubId(), response.RoomId(), roomUserInfo.UserId(), roomUserInfo.Index(),
		string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))

}

func s2cClubQuitRoomPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubQuitRoomPush(impacket.GetBody(), 0)
	showClientDebug("退出房间, clubId:%v, roomId:%v, userId:%v, index:%v",
		response.ClubId(), response.RoomId(), response.UserId(), response.Index())
}

func s2cClubDismissRoomPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubDismissRoomPush(impacket.GetBody(), 0)
	showClientDebug("房间解散, clubId:%v, roomId:%v, code:%v",
		response.ClubId(), response.RoomId(), response.Code())
}

func s2cClubRoomStatusPush(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubRoomStatusPush(impacket.GetBody(), 0)
	showClientDebug("房间开始, clubId:%v, roomId:%v, status:%v, currentRound:%v",
		response.ClubId(), response.RoomId(), response.Status(), response.CurrentRound())
}

func s2cClubClubMessagePush(impacket *protocal.ImPacket) {
	msg := new(fbsCommon.Msg)
	sender := new(fbsCommon.MsgSender)
	response := fbsCommon.GetRootAsClubClubMessagePush(impacket.GetBody(), 0)
	msg = response.Msg(msg)
	msg.Sender(sender)
	showClientDebug("收到俱乐部消息, clubId:%v, sender:%v, mId:%v, mType:%v, content:%v", response.ClubId(), sender.UserId(), msg.MId(), msg.MType(), string(msg.Content()))
}

func s2cClubClubMessageListPush(impacket *protocal.ImPacket) {
	msg := new(fbsCommon.Msg)
	sender := new(fbsCommon.MsgSender)
	response := fbsCommon.GetRootAsClubClubMessageListPush(impacket.GetBody(), 0)
	for i := 0; i < response.MsgListLength(); i++ {
		response.MsgList(msg, i)
		sender = msg.Sender(sender)
		showClientDebug("收到俱乐部历史消息消息, clubId:%v, sender:%v, mId:%v, mType:%v, [%v]%v", response.ClubId(), sender.UserId(), msg.MId(), msg.MType(), util.FormatUnixTime(msg.T()), string(msg.Content()))
	}
}
func c2iRoomListResponse(impacket *protocal.ImPacket) {
	response := fbsCommon.GetRootAsClubC2IRoomListResponse(impacket.GetBody(), 0)
	showClientDebug("服务器回应的俱乐部房间列表,clubId:%v", response.ClubId())
	result := new(fbsCommon.GameResult)
	room := new(fbsCommon.Room)
	roomInfo := new(fbsCommon.RoomInfo)
	roomUserInfo := new(fbsCommon.RoomUserInfo)

	response.C2iResult(result)
	if result.Code() != 0 {
		showClientDebug("请求出错,code:%v,error:%v", result.Code(), string(result.Msg()))
		return
	}

	for i := 0; i < response.RoomListLength(); i++ {
		showClientDebug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		response.RoomList(room, i)
		roomInfo = room.RoomInfo(roomInfo)
		showClientDebug("房间信息,roomId:%v,number:%v,gType:%v,playerCount:%v,status:%v,currentRound:%v,round:%v, setting:%v",
			roomInfo.RoomId(), string(roomInfo.Number()), roomInfo.GameType(), roomInfo.PlayerCount(), room.Status(),
			room.CurrentRound(), roomInfo.Round(), roomInfo.SettingBytes())

		for j := 0; j < room.RoomUsersLength(); j++ {
			room.RoomUsers(roomUserInfo, j)
			showClientDebug("房间用户信息:userId:%v,index:%v,nickname:%v,avatar:%v",
				roomUserInfo.UserId(), roomUserInfo.Index(), string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))
		}
		showClientDebug("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	}
}

// 收到联赛大厅关闭的消息
func s2cCloseClubPush(impacket *protocal.ImPacket) {
	showClientDebug("[s2cCloseClubPush]received")
}
