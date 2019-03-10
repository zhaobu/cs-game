package game

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

// Game 联赛服务
type Game struct {
	UserId       int
	Version      string
	SelectType   string
	Conn         *net.TCPConn
	ConnStatus   int                     // 连接状态
	ListenStatus bool                    // 连接的监控状态
	Queue        chan *protocal.ImPacket // 消息队列
}

// NewGame 新建一个到联赛的连接
func NewGame(userId int, version, selectType string) *Game {
	g := &Game{userId, version, selectType, nil, CONN_STATUS_UNCONNECTED, false, make(chan *protocal.ImPacket, 1024)}
	return g
}

// Connect 连接游戏服务器
func (g *Game) Connect() (int, error) {
	result := selectserver.SelectByType(g.UserId, g.Version, g.SelectType)
	if result == nil || result.Code < 0 {
		var code int
		if result != nil {
			code = result.Code
		}
		return code, fmt.Errorf("select server error, code:%v", code)
	}
	core.Logger.Info("[game.Connect]select server result, userId:%v, version:%v, result:%+v", g.UserId, g.Version, result)
	conn, err := net.DialTimeout("tcp", result.Remote, time.Second)
	if err != nil {
		return result.Code, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return result.Code, fmt.Errorf("[game.Connect]conn 转化成 tcp.Conn失败")
	}

	g.Conn = tcpConn
	g.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	return result.Code, nil
}

// Handshake 握手
func (g *Game) Handshake(token string, roomId int64, round, seq int) {
	user := make(map[string]interface{})
	// user["token"] = util.GenToken(id, "latest", config.TOKEN_SECRET_KEY)
	user["token"] = token
	user["no_heartbeat"] = 1
	user["from"] = "gateway"
	user["room_id"] = roomId
	user["round"] = round
	user["step"] = seq
	js := simplejson.New()
	js.Set("user", user)
	message, _ := js.Encode()

	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	g.writeMessage(p)
	g.ConnStatus = CONN_STATUS_HANDSHAKE

	core.Logger.Info("[game.Handshake]handshake, userId:%v, date:%#v", g.UserId, user)
}

// HandshakeAck 握手确认
func (g *Game) HandshakeAck() {
	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, nil)
	g.writeMessage(p)
	g.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	// 开启消息推送
	go g.gameMessageSend()
	core.Logger.Info("[game.HandshakeAck]handshakeAck, userId:%v", g.UserId)
}

// AppendMessage 向消息队列追加消息
func (g *Game) AppendMessage(p *protocal.ImPacket) {
	g.Queue <- p
}

// writeMessage 发送消息
func (g *Game) writeMessage(p *protocal.ImPacket) {
	_, err := g.Conn.Write(p.Serialize())
	ierror.MustNil(err)
}

func (g *Game) gameMessageSend() {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[game.gameMessageSend]关闭向游戏服的消息推送, userId:%v", g.UserId)
	}()
	core.Logger.Info("[game.gameMessageSend]开启向游戏服的消息推送, userId:%v", g.UserId)
	for p := range g.Queue {
		g.writeMessage(p)
	}
}

// NeedReconnect 判断是否需要发起重连
func (g *Game) NeedReconnect() bool {
	return g.ConnStatus == CONN_STATUS_UNCONNECTED
}

// IsClosed 连接是否处于关闭状态
func (g *Game) IsClosed() bool {
	return g.ConnStatus == CONN_STATUS_CLOSED
}

// IsConnected 连接是否处于连接状态
func (g *Game) IsConnected() bool {
	return g.ConnStatus == CONN_STATUS_SOCKET_CONNECTED
}

// Close 关闭连接
func (g *Game) Close() {
	if g.ConnStatus != CONN_STATUS_CLOSED && g.ConnStatus != CONN_STATUS_UNCONNECTED {
		g.ConnStatus = CONN_STATUS_CLOSED
		close(g.Queue)
		g.Conn.Close()
		core.Logger.Info("[game.Close]connection colsed, userId:%v", g.UserId)
	}
	// core.Logger.Info("[game.Close]未连接，无需断开, userId:%v, connect status:%v", g.UserId, g.ConnStatus)
}
