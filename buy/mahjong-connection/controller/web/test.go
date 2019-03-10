package web

import (
	"mahjong-connection/service"
	"net/http"
)

// TestAction ????
func TestAction(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// BroadcastMessageAction ????
func BroadcastMessageAction(w http.ResponseWriter, r *http.Request) {
	service.BroadcastMessage(w, r)
}

// PrivateMessageAction 私聊消息
func PrivateMessageAction(w http.ResponseWriter, r *http.Request) {
	service.PrivateMessage(w, r)
}
