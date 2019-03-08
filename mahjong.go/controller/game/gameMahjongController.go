package game

import (
	"net"

	"mahjong.go/mi/protocal"
	"mahjong.go/library/core"

	fbsCommon "mahjong.go/fbs/Common"
	gameService "mahjong.go/service/game"
)

// 用户回应
func MahjongUserOperationAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	// 解析出body中的数据
	request := fbsCommon.GetRootAsUserOperationNotify(impacket.GetBody(), 0)
	op := request.Op(nil)

	err := gameService.UserOperationAction(userId, op)
	if err != nil {
		core.Logger.Debug(err.Error())
	}
}
