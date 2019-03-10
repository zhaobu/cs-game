package main

import (
	"fmt"
	"io"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/protocal"
	"os"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
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

// s2cLeagueListPush 收到服务器推送联赛列表
func s2cLeagueListPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueListPush")
	response := fbsCommon.GetRootAsLeagueListPush(impacket.GetBody(), 0)
	league := new(fbsCommon.League)
	for i := 0; i < response.LeagueListLength(); i++ {
		response.LeagueList(league, i)
		showLeague(league)
	}
}

// s2cLeagueListPush 收到服务器推送联赛列表
func s2cLeagueRaceUserSignupCountPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueRaceUserSignupCountPush")
	response := fbsCommon.GetRootAsLeagueRaceSignupCountPush(impacket.GetBody(), 0)
	showClientDebug("比赛报名人数更新, leagueId:%v, raceId:%v, sign up count:%v", response.LeagueId(), response.RaceId(), response.Count())
}

// s2cLeagueListResponse 收到服务器推送联赛列表
func s2cLeagueListResponse(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueListResponse")
	response := fbsCommon.GetRootAsLeagueListResponse(impacket.GetBody(), 0)
	league := new(fbsCommon.League)
	for i := 0; i < response.LeagueListLength(); i++ {
		response.LeagueList(league, i)
		showLeague(league)
	}
}

func showLeague(league *fbsCommon.League) {
	fmt.Println("[[[showLeague>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Printf("id:%v, name:%v, icon:%v, leagueType:%v, gameType:%v, setting:%v.\n", league.Id(), string(league.Name()), string(league.Icon()), league.LeagueType(), league.GameType(), league.SettingBytes())
	fmt.Printf("requireUserCount:%v, signupUserCount:%v.\n", league.RequireUserCount(), league.SignupUserCount())
	fmt.Printf("price:%v.\n", league.Price())
	signTime := league.SignupTime()
	if signTime > 0 {
		fmt.Printf("报名开始时间:%v, 放弃报名时间:%v, 比赛开始时间:%v", util.FormatUnixTime(league.SignupTime()), util.FormatUnixTime(league.GiveupTime()), util.FormatUnixTime(league.StartTime()))
	}
	fmt.Println("比赛奖励:", league.RewardsLength())
	for i := 0; i < league.RewardsLength(); i += 2 {
		fmt.Println(string(league.Rewards(i)), string(league.Rewards(i+1)))
	}
	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<showLeague]]]")
}

// 收到报名参赛结果
func s2cLeagueApplyResponse(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueApplyResponse")
	response := fbsCommon.GetRootAsLeagueApplyResponse(impacket.GetBody(), 0)
	result := new(fbsCommon.GameResult)
	raceInfo := new(fbsCommon.Race)
	raceUserInfo := new(fbsCommon.RaceUser)
	response.S2cResult(result)
	response.RaceInfo(raceInfo)
	response.RaceUserInfo(raceUserInfo)
	showClientDebug("[s2cLeagueApplyResponse]result, code:%v, msg:%v", result.Code(), string(result.Msg()))
	if result.Code() == int32(0) {
		showRaceInfo(raceInfo)
		showRaceUserInfo(raceUserInfo)
	}
}
func showRaceInfo(raceInfo *fbsCommon.Race) {
	fmt.Println("--------------race info -------------->>>")
	fmt.Println("raceId:", raceInfo.RaceId())
	fmt.Println("leagueId:", raceInfo.LeagueId())
	fmt.Println("name:", string(raceInfo.Name()))
	fmt.Println("icon:", raceInfo.Icon())
	fmt.Println("gameType:", raceInfo.GameType())
	fmt.Println("setting:", raceInfo.SettingBytes())
	fmt.Println("rounds:", raceInfo.Rounds())
	fmt.Println("requireUserCount:", raceInfo.RequireUserCount())
	fmt.Println("signupUserCount:", raceInfo.SignupUserCount())
	fmt.Println("price:", raceInfo.Price())
	fmt.Println("leagueType:", raceInfo.LeagueType())
	fmt.Println("signupTime:", raceInfo.SignupTime())
	fmt.Println("giveupTime:", raceInfo.GiveupTime())
	fmt.Println("startTime:", raceInfo.StartTime())
	fmt.Println("round:", raceInfo.RaceId())
	// fmt.Println("rewards:", raceInfo.RaceId())
	fmt.Println("比赛奖励:", raceInfo.RewardsLength())
	for i := 0; i < raceInfo.RewardsLength(); i += 2 {
		fmt.Println(string(raceInfo.Rewards(i)), string(raceInfo.Rewards(i+1)))
	}

	fmt.Println("--------------race info --------------<<<")
}

func showRaceUserInfo(raceUserInfo *fbsCommon.RaceUser) {
	fmt.Println("--------------race user info -------------->>>")
	fmt.Println("raceId:", raceUserInfo.RaceId())
	fmt.Println("userId:", raceUserInfo.UserId())
	fmt.Println("round:", raceUserInfo.Round())
	fmt.Println("status:", raceUserInfo.Status())
	fmt.Println("score:", raceUserInfo.Score())
	fmt.Println("signupTime:", util.FormatUnixTime(raceUserInfo.SignupTime()))
	fmt.Println("giveupTime:", raceUserInfo.GiveupTime())
	fmt.Println("failTime:", raceUserInfo.FailTime())
	fmt.Println("rank:", raceUserInfo.Rank())
	fmt.Println("--------------race user info --------------<<<")
}

// 收到取消报名参赛结果
func s2cLeagueCancelResponse(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueCancelResponse")
	response := fbsCommon.GetRootAsLeagueCancelResponse(impacket.GetBody(), 0)
	result := new(fbsCommon.GameResult)
	response.S2cResult(result)

	showClientDebug("[s2cLeagueCancelResponse]result, code:%v, msg:%v", result.Code(), string(result.Msg()))
}

// 收到退赛接轨哦
func s2cLeagueQuitResponse(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueQuitResponse")
	response := fbsCommon.GetRootAsLeagueQuitResponse(impacket.GetBody(), 0)
	result := new(fbsCommon.GameResult)
	response.S2cResult(result)
	showClientDebug("[s2cLeagueCancelResponse]result, code:%v, msg:%v", result.Code(), string(result.Msg()))
	if result.Code() == int32(0) {
		raceInfo := new(fbsCommon.Race)
		raceInfo = response.RaceInfo(raceInfo)
		showRaceInfo(raceInfo)
	}
}

func s2cLeagueRacePush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueRacePush")
	response := fbsCommon.GetRootAsLeagueRacePush(impacket.GetBody(), 0)
	raceInfo := response.RaceInfo(nil)
	raceUserInfo := response.RaceUserInfo(nil)
	// fmt.Printf("======%#v", raceInfo)
	if raceInfo != nil {
		showRaceInfo(raceInfo)
	}
	if raceUserInfo != nil {
		showRaceUserInfo(raceUserInfo)
	}
}

// 收到比赛上架消息
func s2cLeagueListReloadPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueListReloadPush")
	response := fbsCommon.GetRootAsLeagueListReloadPush(impacket.GetBody(), 0)
	league := new(fbsCommon.League)
	response.League(league)
	showLeague(league)
}

// 收到比赛下架消息
func s2cLeagueListRemovePush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueListRemovePush")
	response := fbsCommon.GetRootAsLeagueListRemovePush(impacket.GetBody(), 0)
	showClientDebug("[s2cLeagueListRemovePush]result, leagueId:%v", response.LeagueId())
}

func s2cLeagueGameRankPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueGameRankPush")
	push := fbsCommon.GetRootAsLeagueGameRankPush(impacket.GetBody(), 0)
	showClientDebug("[s2cLeagueGameRankPush]result,比赛排名发生变化,rank:%v, score:%v, 剩余房间:%v", push.Rank(), push.Score(), push.RoomCnt())
}
func s2cLeagueRaceCancelPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueRaceCancelPush")
}
func s2cLeagueGameStartPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueGameStartPush")
	push := fbsCommon.GetRootAsLeagueGameStartPush(impacket.GetBody(), 0)
	showClientDebug("[s2cLeagueGameStartPush]result,raceId:%v, roomId:%v", push.RaceId(), push.RoomId())
}
func s2cLeagueRaceResultPush(impacket *protocal.ImPacket) {
	showClientDebug("receive s2cLeagueRaceResultPush")
	push := fbsCommon.GetRootAsLeagueRaceResultPush(impacket.GetBody(), 0)
	showClientDebug("[s2cLeagueRaceResultPush]userId:%v, rank:%v, rewards:%v", push.UserId(), push.Rank(), string(push.RewardsDesc()))
}
