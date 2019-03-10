package game

import (
	"net"

	"github.com/bitly/go-simplejson"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"

	gameService "mahjong.go/service/game"
)

func SystemHandlerAction(conn *net.TCPConn, impacket *protocal.ImPacket) {
	// 解析参数
	js, _ := simplejson.NewJson(impacket.GetMessage())
	systemKey, _ := js.Get("systemKey").String()
	// 检查systemkey
	if systemKey != config.SYSTEM_KEY {
		err := core.NewError(-5)
		conn.Write(gameService.GenJsonError(protocal.PACKAGE_TYPE_SYSTEM, err).Serialize())
		return
	}

	var err *core.Error
	err = nil

	act, _ := js.Get("act").String()
	core.Logger.Debugf("[SystemHandlerAction]act:%s", act)
	switch act {
	case "roomInfo": // 拉取房间信息
		roomId, _ := js.Get("roomId").Int64()
		err = gameService.GetRoomInfo(conn, roomId)
	case "gameDetail": // 游戏明细，测试使用
		roomId, _ := js.Get("roomId").Int64()
		err = gameService.GetGameDetail(conn, roomId)
	case "stat": // 游戏统计信息
		v, _ := js.Get("v").Int()
		gameService.Stat(conn, v)
	case "jr": // 加入房间
		err = gameService.H5JoinRoom(conn, js)
	case "create": // 创建犯贱
		err = gameService.H5CreateRoom(conn, js)
	case "dismiss": // 创建犯贱
		err = gameService.H5DismissRoom(conn, js)
	case "hf": // 心跳日志
		err = gameService.HeartBeatFlag(conn, js)
	default:
		err = core.NewError(-1)
	}

	if err != nil {
		core.Logger.Error("SystemHandlerAction:%s", err.Error())
		conn.Write(gameService.GenJsonError(protocal.PACKAGE_TYPE_SYSTEM, err).Serialize())
	}
}
