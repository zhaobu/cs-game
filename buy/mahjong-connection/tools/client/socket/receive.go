package main

import (
	"io"
	fbsCommon "mahjong-connection/fbs/Common"
	"mahjong-connection/protocal"
	"os"

	simplejson "github.com/bitly/go-simplejson"
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
	case fbsCommon.CommandLeagueListPush:
		s2cLeagueListPush(impacket)
	case fbsCommon.CommandLeagueRaceSignupCountPush:
		s2cLeagueRaceUserSignupCountPush(impacket)
	case fbsCommon.CommandLeagueListResponse:
		s2cLeagueListResponse(impacket)
	case fbsCommon.CommandLeagueApplyResponse:
		s2cLeagueApplyResponse(impacket)
	case fbsCommon.CommandLeagueCancelResponse:
		s2cLeagueCancelResponse(impacket)
	case fbsCommon.CommandLeagueQuitResponse:
		s2cLeagueQuitResponse(impacket)
	case fbsCommon.CommandLeagueRacePush:
		s2cLeagueRacePush(impacket)
	case fbsCommon.CommandLeagueListReloadPush:
		s2cLeagueListReloadPush(impacket)
	case fbsCommon.CommandLeagueListRemovePush:
		s2cLeagueListRemovePush(impacket)
	case fbsCommon.CommandLeagueGameRankPush:
		s2cLeagueGameRankPush(impacket)
	case fbsCommon.CommandLeagueRaceCancelPush:
		s2cLeagueRaceCancelPush(impacket)
	case fbsCommon.CommandLeagueGameStartPush:
		s2cLeagueGameStartPush(impacket)
	case fbsCommon.CommandLeagueRaceResultPush:
		s2cLeagueRaceResultPush(impacket)
	case fbsCommon.CommandClubJoinResponse:
		s2cClubJoinResponse(impacket)
	case fbsCommon.CommandClubQuitResponse:
		s2cClubQuitResponse(impacket)
	case fbsCommon.CommandClubRestorePush:
		sc2ClubRestorePush(impacket)
	case fbsCommon.CommandClubReloadRoomPush:
		s2cClubReloadRoomPush(impacket)
	case fbsCommon.CommandClubJoinRoomPush:
		s2cClubJoinRoomPush(impacket)
	case fbsCommon.CommandClubQuitRoomPush:
		s2cClubQuitRoomPush(impacket)
	case fbsCommon.CommandClubDismissRoomPush:
		s2cClubDismissRoomPush(impacket)
	case fbsCommon.CommandClubRoomStatusPush:
		s2cClubRoomStatusPush(impacket)
	case fbsCommon.CommandClubClubMessagePush:
		s2cClubClubMessagePush(impacket)
	case fbsCommon.CommandClubClubMessageListPush:
		s2cClubClubMessageListPush(impacket)
	case fbsCommon.CommandClubC2IRoomListResponse:
		c2iRoomListResponse(impacket)
	case fbsCommon.CommandGameActivateResponse:
		s2cGameActivateResponse(impacket)
	case fbsCommon.CommandGatewayS2CCloseClubPush:
		s2cCloseClubPush(impacket)
	case fbsCommon.CommandGatewayS2CCloseLeaguePush:
		s2cCloseLeaguePush(impacket)
	case fbsCommon.CommandGatewayS2CCloseGamePush:
		s2cCloseGamePush(impacket)
	case fbsCommon.CommandGatewayS2CPrivateMessagePush:
		s2cPrivateMessagePush(impacket)
	default:
		showClientError("[onRecivedData]未支持的协议id:%v", command)
	}
}

// 收到服务端的握手消息
func s2cHandshake(impacket *protocal.ImPacket) {
	// 握手成功
	js, _ := simplejson.NewJson(impacket.GetMessage())
	vmap, _ := js.Map()
	showClientDebug("s2cHandshake:%#v\n", vmap)
	heartbeatInterval, _ = js.Get("heartbeat").Int()
	code, _ := js.Get("code").Int()
	message, _ := js.Get("message").String()
	if code != 0 {
		showClientError("[s2cHandshake]握手失败, code:%v, message:%v", code, message)
		return
	}

	go c2sHandShakeAck()
	showClientDebug("receive handshake, heartbeatInterval:%v", heartbeatInterval)

	if heartbeatInterval > 0 {
		// 收到握手，开启心跳
		go c2sHeartBeat()
	} else {
		showClientDebug("[s2cHandshake]心跳间隔为0，不开启心跳")
	}
}

// 收到服务端的心跳
func s2cHeartbeat() {
	// showClientDebug("receive heartbeah")
}
