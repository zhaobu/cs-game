package game

import (
	"net"

	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"

	fbsCommon "mahjong.go/fbs/Common"
	gameService "mahjong.go/service/game"
)

// CreateRoomAction 创建房间
func CreateRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析参数
	request := fbsCommon.GetRootAsCreateRoomRequest(impacket.GetBody(), 0)
	// 游戏类型
	gType := int(request.GameType())
	// 游戏局数
	round := int(request.Round())
	// 解析组队设置选项
	setting := []int{}
	for i := 0; i < request.SettingLength(); i++ {
		setting = append(setting, int(request.Setting(i)))
	}
	// 俱乐部id
	clubId := int(request.ClubId())
	core.Logger.Debugf("[CreateRoomAction]userId:%v,gType:%v,round:%v,clubId:%v,setting:%#v", userId, gType, round, clubId, setting)

	// 创建房间
	err := gameService.CreateRoom(userId, gType, round, setting, clubId, impacket.GetMessageNumber())
	if err != nil {
		// 创建失败
		core.Logger.Error("创建房间失败,userId:%v,error:%s", userId, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// 加入房间
func JoinRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析出body中的数据
	request := fbsCommon.GetRootAsJoinRoomRequest(impacket.GetBody(), 0)
	// 房间号
	number := string(request.Number())

	// 加入房间
	err := gameService.JoinRoom(userId, number, impacket.GetMessageNumber())
	if err != nil {
		// 加入失败
		core.Logger.Error("[JoinRoomAction]失败,userId:%v, error:%v", userId, err.Error())

		// 回应失败消息
		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// 退出房间
func QuitRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.QuitRoom(userId)
	if err != nil {
		// 退出错误
		core.Logger.Error("[QuitRoomAction]userId:%v,error:%v", userId, err.Error())
	}
}

// 随机组队
func RandomJoinAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析参数
	request := fbsCommon.GetRootAsRandomRoomRequest(impacket.GetBody(), 0)
	// 游戏类型
	gType := int(request.GameType())

	// 随机组队
	err := gameService.RandomJoin(userId, gType, impacket.GetMessageNumber())
	if err != nil {
		// 组队失败
		core.Logger.Error("[RandomJoinAction]失败, userId:%v, error:%v", userId, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// 加入组队
func MatchJoinAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析参数
	request := fbsCommon.GetRootAsMatchRoomRequest(impacket.GetBody(), 0)
	// 游戏类型
	gType := int(request.GameType())

	// 随机组队
	err := gameService.MatchJoin(userId, gType, impacket.GetMessageNumber())
	if err != nil {
		// 组队失败
		core.Logger.Error("[MatchJoinAction]失败, userId:%v,gType:%v, error:%v", userId, gType, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// 金币场
func CoinJoinAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 随机组队
	err := gameService.CoinJoin(userId, impacket)
	if err != nil {
		// 组队失败
		core.Logger.Error("[CoinJoinAction]失败, userId:%v, error:%v", userId, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// RankRoomRequestAction 排位赛
func RankRoomRequestAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.RankRoomRequest(userId, impacket)
	if err != nil {
		// 组队失败
		core.Logger.Error("[RankRoomRequestAction]失败, userId:%v, error:%v", userId, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// RankRoomRequestRobotAction 机器人参加排位赛
func RankRoomRequestRobotAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.RankRoomRobotRequest(userId, impacket)
	if err != nil {
		// 组队失败
		core.Logger.Error("[RankRoomRequestRobotAction]失败, userId:%v, error:%v", userId, err.Error())

		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// 申请解散房间
func DismissRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析参数
	request := fbsCommon.GetRootAsDismissRoomNotify(impacket.GetBody(), 0)
	// 进行什么操作，-1：申请解散；0：同意解散；1：拒绝解散
	op := int(request.Op())

	if op == config.ROOM_DISMISS_APPLY {
		// 申请解散房间
		err := gameService.DismissApply(userId)
		if err != nil {
			core.Logger.Error("[DismissRoomAction]失败,userId:%v, error:%v", userId, err.Error())
		}
	} else {
		// 回应解散房间
		err := gameService.DismissReply(userId, op)
		if err != nil {
			core.Logger.Error("[DismissRoomAction]失败,userId:%v, error:%v", userId, err.Error())
		}
	}
}

// 客户端准备
func RoomReadyAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析出body中的数据
	// request := fbsCommon.GetRootAsGameReadyNotify(impacket.GetBody(), 0)
	// readying := int(request.Readying())
	// agree := int(request.Agree())

	err := gameService.GameReady(userId, impacket)
	if err != nil {
		core.Logger.Error("[RoomReadyAction]失败,userId:%v, error:%v", userId, err.Error())
	}
}

// RoomRestoreAction 客户端请求重连
func RoomRestoreAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.GameRestore(userId, impacket)
	if err != nil {
		core.Logger.Error("[RoomRestoreAction]失败,userId:%v, error:%v", userId, err.Error())
		gameService.SendMessageByUserId(userId, gameService.GameRestoreFailpush(err))
	}
}

// 客户端重连完成
func RoomRestoreDoneAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.GameRestoreDone(userId)
	if err != nil {
		core.Logger.Error("[RoomRestoreDoneAction]失败,userId:%v, error:%v", userId, err.Error())

		// 推送重连失败的push
		gameService.SendMessageByUserId(userId, gameService.GameRestoreFailpush(err))
	}
}

// 用户托管
func UserHostingAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	hostingStatus := fbsCommon.GetRootAsGameHostingNotify(impacket.GetBody(), 0).HostingStatus()
	err := gameService.UserHosting(userId, int(hostingStatus))
	if err != nil {
		core.Logger.Error("[UserHostingAction]失败,userId:%v, error:%v", userId, err.Error())
	}
}

// ObRoomAction 观察房间
func ObRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析出body中的数据
	request := fbsCommon.GetRootAsObRoomRequest(impacket.GetBody(), 0)
	// 房间号
	number := string(request.Number())

	// 观察房间
	err := gameService.ObRoom(userId, number, impacket.GetMessageNumber())
	if err != nil {
		// 加入失败
		core.Logger.Error("[ObRoomAction]失败,userId:%v, error:%v", userId, err.Error())

		// 回应失败消息
		gameService.JoinRoomFailedResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}

// EndRoomAction 观察员直接结束房间
func EndRoomAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.EndRoom(userId)
	if err != nil {
		// 退出错误
		core.Logger.Error("[EndRoomAction]userId:%v,error:%v", userId, err.Error())
	}
}

// UserDistanceAction 获取房间用户之间的距离
func UserDistanceAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	err := gameService.GetDistanceList(userId, impacket)
	if err != nil {
		core.Logger.Error("[UserDistanceAction]userId:%v,error:%v", userId, err.Error())
		gameService.GameUserDistanceFailResponse(err, impacket.GetMessageNumber()).Send(conn)
	}
}
