package router

import (
	"net"

	"github.com/fwhappy/mahjong/protocal"
	cc "mahjong.club/controller/club"
	gc "mahjong.club/controller/game"
	sc "mahjong.club/controller/system"
	uc "mahjong.club/controller/user"
	"mahjong.club/core"
	fbsCommon "mahjong.club/fbs/Common"
)

// Dispatch userID
// 用户连接成功之后，userId，需要回传
func Dispatch(userID *int, conn *net.TCPConn, impacket *protocal.ImPacket, c chan int) {
	// 解析数据包
	packageID := impacket.GetPackage()
	switch packageID {
	case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手
		// 记录当前连接的userID, 这样客户端无需每次传给服务端
		*userID = uc.HandShake(conn, impacket)
	case protocal.PACKAGE_TYPE_HANDSHAKE_ACK: // 握手成功
		uc.HandShakeAck(*userID, conn, impacket)
		c <- 1
	case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳
		uc.HeartBeat(*userID, conn, impacket)
	case protocal.PACKAGE_TYPE_DATA: // 数据包
		// 数据包路由分发
		routerData(*userID, conn, impacket)
	case protocal.PACKAGE_TYPE_KICK: // 下线
		uc.Logout(*userID, conn, impacket)
	case protocal.PACKAGE_TYPE_SYSTEM: // 系统
		// 用户socket连接后，必须在规定时间内完成握手协议，不然俱乐部端会主动断开连接
		// 游戏服务器与俱乐部端的连接，不会有握手协议，
		// 为了避免连接被服务端关闭，所以游戏服会在socket连接成功之后，发送一个system指令到俱乐部端
		// 俱乐部端收到system后，就会关闭连接的握手检测
		c <- 1
	default:
		core.Logger.Error("[router.Dispatch]未支持的数据包id:%d", packageID)
	}
}

// 数据包分发
func routerData(userID int, conn *net.TCPConn, impacket *protocal.ImPacket) {
	messageID := int(impacket.GetMessageId())
	switch messageID {

	// 来自客户端的消息
	case fbsCommon.CommandClubJoinRequest:
		cc.JoinAction(userID, conn, impacket)
	case fbsCommon.CommandClubQuitRequest:
		cc.QuitAction(userID, conn, impacket)
	case fbsCommon.CommandClubClubMessageNotify:
		cc.SendMessageAction(userID, conn, impacket)
	case fbsCommon.CommandClubClubMessageListNotify:
		cc.MessageListAction(userID, conn, impacket)

	// 来自游戏服的消息
	case fbsCommon.CommandClubG2CReloadRoomPush: // 重载房间
		gc.ReloadRoomAction(conn, impacket)
	case fbsCommon.CommandClubG2CJoinRoomPush: // 加入房间
		gc.JoinRoomAction(conn, impacket)
	case fbsCommon.CommandClubG2CQuitRoomPush: // 退出房间
		gc.QuitRoomAction(conn, impacket)
	case fbsCommon.CommandClubG2CDismissRoomPush: // 解散房间
		gc.DismissRoomAction(conn, impacket)
	case fbsCommon.CommandClubG2CStartRoomPush: // 房间开始
		gc.StartRoomAction(conn, impacket)
	case fbsCommon.CommandClubG2CRoomActiveRequest: // 房间活跃检测
		gc.RoomActiveAction(conn, impacket)

	// 来自info的消息
	case fbsCommon.CommandClubI2CRoomListRequest: // 获取房间列表
		sc.RoomListAction(conn, impacket)
	default:
		core.Logger.Error("未支持的消息id:%d", messageID)
	}
}
