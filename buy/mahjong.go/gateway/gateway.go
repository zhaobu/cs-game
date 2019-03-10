package gateway

import (
	"fmt"

	"mahjong.go/library/core"

	"github.com/fwhappy/util"
)

// SendPrivateMessage 发送消息
func SendPrivateMessage(b []byte) {
	if len(core.AppConfig.GatewayRemotes) == 0 {
		core.Logger.Warn("[gateway.SendPrivateMessage]网关地址未设置")
		return
	}

	for _, remote := range core.AppConfig.GatewayRemotes {
		go func(remote string) {
			defer util.RecoverPanic()
			httpPost(fmt.Sprintf("http://%v/privateMessage", remote), b)
		}(remote)
	}
}

// SendBroadcastMessage 发送广播消息
func SendBroadcastMessage(b []byte) {
	if len(core.AppConfig.GatewayRemotes) == 0 {
		core.Logger.Warn("[gateway.SendBroadcastMessage]网关地址未设置")
		return
	}

	for _, remote := range core.AppConfig.GatewayRemotes {
		go func() {
			defer util.RecoverPanic()
			httpPost(fmt.Sprintf("http://%v/broadcastMessage", remote), b)
		}()
	}
}
