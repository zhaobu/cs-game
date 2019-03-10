package club

import (
	"fmt"
	"mahjong-connection/core"
	"mahjong-connection/ierror"
	"mahjong-connection/protocal"
	"mahjong-connection/selectserver"
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

const (
	CONN_STATUS_UNCONNECTED      = 0   // 未连接
	CONN_STATUS_SOCKET_CONNECTED = 10  // 底层socket连接完成
	CONN_STATUS_HANDSHAKE        = 20  // 执行逻辑握手
	CONN_STATUS_CLOSED           = 30  // 已关闭
	CONN_STATUS_CONNECTED        = 100 // 正常连接
)

// Club 俱乐部服务
type Club struct {
	UserId       int
	Version      string
	Conn         *net.TCPConn
	ConnStatus   int                     // 连接状态
	ListenStatus bool                    // 连接的监控状态
	Queue        chan *protocal.ImPacket // 消息队列
}

// NewClub 新建一个到联赛的连接
func NewClub(userId int, version string) *Club {
	l := &Club{userId, version, nil, CONN_STATUS_UNCONNECTED, false, make(chan *protocal.ImPacket, 1024)}
	return l
}

// Connect 连接比赛服务
func (c *Club) Connect() (bool, error) {
	result := selectserver.Club(c.UserId, c.Version)
	if result == nil || result.Code < 0 || len(result.Remote) == 0 {
		var code int
		if result != nil {
			code = result.Code
		}
		return false, fmt.Errorf("select server error, code:%v", code)
	}
	core.Logger.Info("[club.Connect]select server result, userId:%v, version:%v, result:%+v", c.UserId, c.Version, result)
	conn, err := net.DialTimeout("tcp", result.Remote, time.Second)
	if err != nil {
		return false, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return false, fmt.Errorf("[league.Connect]conn 转化成 tcp.Conn失败")
	}

	c.Conn = tcpConn
	c.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	return true, nil
}

// Handshake 握手
func (c *Club) Handshake(token string) {
	user := make(map[string]interface{})
	user["token"] = token
	user["no_heartbeat"] = 1
	user["from"] = "gateway"
	js := simplejson.New()
	js.Set("user", user)
	message, _ := js.Encode()

	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	c.writeMessage(p)
	c.ConnStatus = CONN_STATUS_HANDSHAKE

	core.Logger.Info("[club]handshake, userId:%v, data:%#v", c.UserId, user)
}

// HandshakeAck 握手确认
func (c *Club) HandshakeAck() {
	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, nil)
	c.writeMessage(p)
	c.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	// 开启消息推送
	go c.messageSend()
	core.Logger.Info("[club]handshakeAck, userId:%v", c.UserId)
}

// AppendMessage 发送消息
func (c *Club) AppendMessage(p *protocal.ImPacket) {
	c.Queue <- p
}

// writeMessage 发送消息
func (c *Club) writeMessage(p *protocal.ImPacket) {
	_, err := c.Conn.Write(p.Serialize())
	ierror.MustNil(err)
}

func (c *Club) messageSend() {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[club.messageSend]关闭向俱乐部的消息推送, userId:%v", c.UserId)
	}()
	core.Logger.Info("[club.messageSend]开启向俱乐部的消息推送, userId:%v", c.UserId)
	for p := range c.Queue {
		c.writeMessage(p)
	}
}

// NeedReconnect 判断是否需要发起重连
func (c *Club) NeedReconnect() bool {
	return c.ConnStatus == CONN_STATUS_UNCONNECTED
}

// Close 关闭连接
func (c *Club) Close() {
	if c.ConnStatus != CONN_STATUS_CLOSED && c.ConnStatus != CONN_STATUS_UNCONNECTED {
		c.ConnStatus = CONN_STATUS_CLOSED
		close(c.Queue)
		c.Conn.Close()
		core.Logger.Info("[club]连接断开, userId:%v", c.UserId)
	}
	core.Logger.Info("[club]未连接，无需断开, userId:%v, connect status:%v", c.UserId, c.ConnStatus)
}
