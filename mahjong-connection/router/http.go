package router

import (
	wc "mahjong-connection/controller/web"
	"net/http"
)

// HTTPDespatch 分发http请求
func HTTPDespatch() {
	http.HandleFunc("/privateMessage", wc.PrivateMessageAction)
	http.HandleFunc("/broadcastMessage", wc.BroadcastMessageAction)
}
