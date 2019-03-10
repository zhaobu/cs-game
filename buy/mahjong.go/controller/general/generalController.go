package general

import (
	"net"

	simplejson "github.com/bitly/go-simplejson"
	"mahjong.go/mi/protocal"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
)

// GeneralRequestAction 通用请求协议
func GeneralRequestAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	request := fbsCommon.GetRootAsGeneralRequest(impacket.GetBody(), 0)
	act := string(request.Act())
	t, _ := simplejson.NewJson(request.DataBytes())
	d, _ := t.Map()
	core.Logger.Debug("[GeneralRequestAction]act:%v,data:%v", act, d)
}

// GeneralNotifyAction 通用通知协议
func GeneralNotifyAction(conn *net.TCPConn, userId int, impacket *protocal.ImPacket) {
	request := fbsCommon.GetRootAsGeneralNotify(impacket.GetBody(), 0)
	act := string(request.Act())
	t, _ := simplejson.NewJson(request.DataBytes())
	d, _ := t.Map()
	core.Logger.Debug("[GeneralRequestAction]act:%v,data:%v", act, d)
}
