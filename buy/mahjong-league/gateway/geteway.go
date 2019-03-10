package gateway

import (
	"fmt"
	"mahjong-league/core"

	"github.com/fwhappy/util"
)

// SendPrivateMessage 发送消息
func SendPrivateMessage(b []byte) {
	if len(core.GetAppConfig().GatewayRemotes) == 0 {
		core.Logger.Warn("[gateway.SendPrivateMessage]网关地址未设置")
		return
	}

	for _, remote := range core.GetAppConfig().GatewayRemotes {
		go func() {
			defer util.RecoverPanic()
			httpPost(fmt.Sprintf("http://%v/privateMessage", remote), b)
		}()
	}
}
