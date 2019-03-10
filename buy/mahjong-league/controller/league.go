package controller

import (
	"mahjong-league/core"
	"mahjong-league/protocal"
	"mahjong-league/service"
	"net"
)

// LeagueList 拉取LeagueList列表
func LeagueList(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.LegueListRequest(userID, impacket)
}

// Apply 用户报名
func Apply(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	err := service.ApplyRequest(userId, impacket)
	if err != nil {
		core.Logger.Error("[league.Apply]Error:%v", err.Error())
		service.LeagueApplyError(impacket.GetMessageNumber(), err).Send(conn)
	}
}

// Cancel 用户放弃报名
func Cancel(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	err := service.CancelRequest(userId, impacket)
	if err != nil {
		core.Logger.Error("[league.Cancel]Error:%v", err.Error())
		service.LeagueCancelResponse(impacket.GetMessageNumber(), err).Send(conn)
	}
}

// PlanResult 游戏服推送排赛结果
func PlanResult(conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.PlanResultPush(impacket)
}

// GameFinish 游戏服推送单局完成
func GameFinish(conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.GameFinish(impacket)
}

// RoomFinish 游戏服推送房间完成
func RoomFinish(conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.RoomFinish(impacket)
}

// RoomActive 收到房间活跃消息
func RoomActive(conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.RoomActive(impacket)
}

// RaceResultReceived 客户端收到比赛结果的反馈
func RaceResultReceived(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	service.RaceResultReceived(userID, impacket)
}

// Giveup 用户退赛
func Giveup(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	if err := service.GiveupRequest(userId, impacket); err != nil {
		core.Logger.Error("[league.Giveup]Error:%v", err.Error())
		service.LeagueQuitResponse(impacket.GetMessageNumber(), nil, nil, err).Send(conn)
	}
}

// RobotApply 机器人报名
func RobotApply(userId int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	err := service.RobotApply(userId, impacket)
	if err != nil {
		core.Logger.Error("[league.RobotApply]Error:%v", err.Error())
		service.LeagueApplyError(impacket.GetMessageNumber(), err).Send(conn)
	}
}
