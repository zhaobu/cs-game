package main

import (
	"mahjong-connection/config"
	"mahjong-connection/protocal"
	"mahjong-connection/response"
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

// 消息号码生成器
type numberGenerator struct {
	value int
	mux   *sync.Mutex
}

var (
	mg                *numberGenerator
	heartbeatInterval int
)

func init() {
	mg = &numberGenerator{mux: &sync.Mutex{}}
}

// 生成一个消息号
func (mg *numberGenerator) getNumber() uint16 {
	mg.mux.Lock()
	defer mg.mux.Unlock()
	mg.value++
	return uint16(mg.value)
}

// 客户端向服务端发送握手协议
func c2sHandShake(args ...string) {
	id = getParamsInt(0, args)
	if id == 0 {
		showClientError("用户id缺失")
		return
	}

	user := make(map[string]interface{})
	user["token"] = util.GenToken(id, "latest", config.TOKEN_SECRET_KEY)
	js := simplejson.New()
	js.Set("user", user)
	message, _ := js.Encode()

	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	conn.Write(imPacket.Serialize())

	showClientDebug("send handShake")
}

// client给server发送握手成功
func c2sHandShakeAck() {
	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, nil)
	conn.Write(imPacket.Serialize())
	showClientDebug("send handShakeAck")
}

// client每3秒给server发送一个心跳消息
// 服务端如果超过6秒没有收到包，则认为客户端已离线
func c2sHeartBeat() {
	for {
		time.Sleep(time.Duration(heartbeatInterval) * time.Millisecond)
		conn.Write(response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT).Serialize())
		// _, err := conn.Write(response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT).Serialize())
		// showClientDebug("send c2sHeartBeat, err:%v", err)
	}
}
