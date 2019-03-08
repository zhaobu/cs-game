package main

import (
	"fmt"
	"io"
	"os"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"

	simplejson "github.com/bitly/go-simplejson"
	fbs "mahjong.club/fbs/Common"
)

// 收到服务端消息
func onRecived() {
	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(conn)

		// 检查解析错误
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				showClientError("disconnected")
				os.Exit(0)
			} else {
				// 协议解析错误
				showClientError(err.Error())
			}
			break
		}

		// body, _ := msgpack.Unmarshal(impacket.GetBody())
		switch impacket.GetPackage() {
		case protocal.PACKAGE_TYPE_HANDSHAKE:
			s2cHandshake(impacket)
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK:
		case protocal.PACKAGE_TYPE_HEARTBEAT:
			s2cHeartbeat()
		case protocal.PACKAGE_TYPE_DATA:
			onRecivedData(impacket)
		}
	}
}

func onRecivedData(impacket *protocal.ImPacket) {
	command := impacket.GetMessageId()
	switch command {
	case fbs.CommandClubJoinResponse:
		s2cClubJoinResponse(impacket)
	case fbs.CommandClubQuitResponse:
		s2cClubQuitResponse(impacket)
	case fbs.CommandClubRestorePush:
		sc2ClubRestorePush(impacket)
	case fbs.CommandClubReloadRoomPush:
		s2cClubReloadRoomPush(impacket)
	case fbs.CommandClubJoinRoomPush:
		s2cClubJoinRoomPush(impacket)
	case fbs.CommandClubQuitRoomPush:
		s2cClubQuitRoomPush(impacket)
	case fbs.CommandClubDismissRoomPush:
		s2cClubDismissRoomPush(impacket)
	case fbs.CommandClubRoomStatusPush:
		s2cClubRoomStatusPush(impacket)
	case fbs.CommandClubClubMessagePush:
		s2cClubClubMessagePush(impacket)
	case fbs.CommandClubClubMessageListPush:
		s2cClubClubMessageListPush(impacket)
	case fbs.CommandClubC2IRoomListResponse:
		c2iRoomListResponse(impacket)
	default:
		showClientError("[onRecivedData]未支持的协议id:%v", command)
	}
}

// 收到服务端的握手消息
func s2cHandshake(impacket *protocal.ImPacket) {
	// 握手成功
	js, _ := simplejson.NewJson(impacket.GetMessage())
	vmap, _ := js.Map()
	fmt.Printf("s2cHandshake:%#v\n", vmap)
	heartbeatInterval, _ = js.Get("heartbeat").Int()

	go c2sHandShakeAck()
	showClientDebug("receive handshake, heartbeatInterval:%v", heartbeatInterval)

	// 收到握手ACK，开启心跳
	go c2sHeartBeat()
}

// 收到服务端的心跳
func s2cHeartbeat() {
	// showClientDebug("receive heartbeah")
}

// 收到加入俱乐部请求
func s2cClubJoinResponse(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubJoinResponse(impacket.GetBody(), 0)
	s2cResult := new(fbs.GameResult)
	response.S2cResult(s2cResult)
	if s2cResult.Code() < 0 {
		showClientError("[s2cClubJoinResponse],code:%v, msg:%v", s2cResult.Code(), string(s2cResult.Msg()))
	} else {
		showClientDebug("加入俱乐部成功:%v", response.ClubId())
	}
}

// 收到退出俱乐部请求
func s2cClubQuitResponse(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubQuitResponse(impacket.GetBody(), 0)
	s2cResult := new(fbs.GameResult)
	response.S2cResult(s2cResult)
	if s2cResult.Code() < 0 {
		showClientError("[s2cClubQuitResponse],code:%v, msg:%v", s2cResult.Code(), string(s2cResult.Msg()))
	} else {
		showClientDebug("退出俱乐部成功:%v", response.ClubId())
	}
}

func sc2ClubRestorePush(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubRestorePush(impacket.GetBody(), 0)
	showClientDebug("重新载入俱乐部房间列表,clubId:%v", response.ClubId())
	room := new(fbs.Room)
	roomInfo := new(fbs.RoomInfo)
	roomUserInfo := new(fbs.RoomUserInfo)
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
	response := fbs.GetRootAsClubReloadRoomPush(impacket.GetBody(), 0)
	showClientDebug("重新载入房间,clubId:%v", response.ClubId())
	room := new(fbs.Room)
	room = response.Room(room)
	roomInfo := new(fbs.RoomInfo)
	roomInfo = room.RoomInfo(roomInfo)
	showClientDebug("房间信息,roomId:%v,number:%v,gType:%v,playerCount:%v,status:%v,currentRound:%v,round:%v, setting:%v",
		roomInfo.RoomId(), string(roomInfo.Number()), roomInfo.GameType(), roomInfo.PlayerCount(), room.Status(),
		room.CurrentRound(), roomInfo.Round(), roomInfo.SettingBytes())

	roomUserInfo := new(fbs.RoomUserInfo)
	for i := 0; i < room.RoomUsersLength(); i++ {
		room.RoomUsers(roomUserInfo, i)
		showClientDebug("房间用户信息:userId:%v,index:%v,nickname:%v,avatar:%v",
			roomUserInfo.UserId(), roomUserInfo.Index(), string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))
	}

}

func s2cClubJoinRoomPush(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubJoinRoomPush(impacket.GetBody(), 0)
	roomUserInfo := new(fbs.RoomUserInfo)
	roomUserInfo = response.RoomUserInfo(roomUserInfo)
	showClientDebug("加入房间, clubId:%v, roomId:%v, userId:%v, index:%v,nickname:%v,avatar:%v",
		response.ClubId(), response.RoomId(), roomUserInfo.UserId(), roomUserInfo.Index(),
		string(roomUserInfo.Nickname()), string(roomUserInfo.Avatar()))

}

func s2cClubQuitRoomPush(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubQuitRoomPush(impacket.GetBody(), 0)
	showClientDebug("退出房间, clubId:%v, roomId:%v, userId:%v, index:%v",
		response.ClubId(), response.RoomId(), response.UserId(), response.Index())
}

func s2cClubDismissRoomPush(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubDismissRoomPush(impacket.GetBody(), 0)
	showClientDebug("房间解散, clubId:%v, roomId:%v, code:%v",
		response.ClubId(), response.RoomId(), response.Code())
}

func s2cClubRoomStatusPush(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubRoomStatusPush(impacket.GetBody(), 0)
	showClientDebug("房间开始, clubId:%v, roomId:%v, status:%v, currentRound:%v",
		response.ClubId(), response.RoomId(), response.Status(), response.CurrentRound())
}

func s2cClubClubMessagePush(impacket *protocal.ImPacket) {
	msg := new(fbs.Msg)
	sender := new(fbs.MsgSender)
	response := fbs.GetRootAsClubClubMessagePush(impacket.GetBody(), 0)
	msg = response.Msg(msg)
	msg.Sender(sender)
	showClientDebug("收到俱乐部消息, clubId:%v, sender:%v, mId:%v, mType:%v, content:%v", response.ClubId(), sender.UserId(), msg.MId(), msg.MType(), string(msg.Content()))
}

func s2cClubClubMessageListPush(impacket *protocal.ImPacket) {
	msg := new(fbs.Msg)
	sender := new(fbs.MsgSender)
	response := fbs.GetRootAsClubClubMessageListPush(impacket.GetBody(), 0)
	for i := 0; i < response.MsgListLength(); i++ {
		response.MsgList(msg, i)
		sender = msg.Sender(sender)
		showClientDebug("收到俱乐部历史消息消息, clubId:%v, sender:%v, mId:%v, mType:%v, [%v]%v", response.ClubId(), sender.UserId(), msg.MId(), msg.MType(), util.FormatUnixTime(msg.T()), string(msg.Content()))
	}
}
func c2iRoomListResponse(impacket *protocal.ImPacket) {
	response := fbs.GetRootAsClubC2IRoomListResponse(impacket.GetBody(), 0)
	showClientDebug("服务器回应的俱乐部房间列表,clubId:%v", response.ClubId())
	result := new(fbs.GameResult)
	room := new(fbs.Room)
	roomInfo := new(fbs.RoomInfo)
	roomUserInfo := new(fbs.RoomUserInfo)

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
