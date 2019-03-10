package league

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

// League 联赛服务
type League struct {
	UserId       int
	Version      string
	Conn         *net.TCPConn
	ConnStatus   int                     // 连接状态
	ListenStatus bool                    // 连接的监控状态
	Queue        chan *protocal.ImPacket // 消息队列
}

// NewLeague 新建一个到联赛的连接
func NewLeague(userId int, version string) *League {
	l := &League{userId, version, nil, CONN_STATUS_UNCONNECTED, false, make(chan *protocal.ImPacket, 1024)}
	return l
}

// Connect 连接比赛服务
func (l *League) Connect() (bool, error) {
	result := selectserver.LeagueServer(l.UserId, l.Version)
	if result == nil || len(result.Remote) == 0 {
		return false, fmt.Errorf("select league remote error, userId:%v", l.UserId)
	}
	core.Logger.Info("[league.Connect]select server result, userId:%v, version:%v, remote:%v", l.UserId, l.Version, result.Remote)
	conn, err := net.DialTimeout("tcp", result.Remote, time.Second)
	if err != nil {
		return false, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return false, fmt.Errorf("[league.Connect]conn 转化成 tcp.Conn失败")
	}

	l.Conn = tcpConn
	l.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	return true, nil
}

// Handshake 握手
func (l *League) Handshake(token string) {
	user := make(map[string]interface{})
	// user["token"] = util.GenToken(id, "latest", config.TOKEN_SECRET_KEY)
	user["token"] = token
	user["no_heartbeat"] = 1
	user["from"] = "gateway"
	js := simplejson.New()
	js.Set("user", user)
	message, _ := js.Encode()

	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE, message)
	l.writeMessage(p)
	l.ConnStatus = CONN_STATUS_HANDSHAKE

	core.Logger.Info("[league]handshake, userId:%v, date:%#v", l.UserId, user)
}

// HandshakeAck 握手确认
func (l *League) HandshakeAck() {
	p := protocal.NewImPacket(protocal.PACKAGE_TYPE_HANDSHAKE_ACK, nil)
	l.writeMessage(p)
	l.ConnStatus = CONN_STATUS_SOCKET_CONNECTED
	// 开启消息推送
	go l.leagueMessageSend()
	core.Logger.Info("[league]handshakeAck, userId:%v", l.UserId)
}

// SendMessage 发送消息
func (l *League) AppendMessage(p *protocal.ImPacket) {
	l.Queue <- p
}

// writeMessage 发送消息
func (l *League) writeMessage(p *protocal.ImPacket) {
	_, err := l.Conn.Write(p.Serialize())
	ierror.MustNil(err)
}

func (l *League) leagueMessageSend() {
	defer util.RecoverPanic()
	defer func() {
		core.Logger.Info("[leagueMessageSend]关闭向联赛大厅的消息推送, userId:%v", l.UserId)
	}()
	core.Logger.Info("[leagueMessageSend]开启向联赛大厅的消息推送, userId:%v", l.UserId)
	for p := range l.Queue {
		l.writeMessage(p)
	}
}

// NeedReconnect 判断是否需要发起重连
func (l *League) NeedReconnect() bool {
	return l.ConnStatus == CONN_STATUS_UNCONNECTED
}

// Close 关闭连接
func (l *League) Close() {
	if l.ConnStatus != CONN_STATUS_CLOSED && l.ConnStatus != CONN_STATUS_UNCONNECTED {
		l.ConnStatus = CONN_STATUS_CLOSED
		close(l.Queue)
		l.Conn.Close()
		core.Logger.Info("[league]connection colsed, userId:%v", l.UserId)
	}
}
