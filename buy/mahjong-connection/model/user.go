package model

import (
	"mahjong-connection/club"
	"mahjong-connection/core"
	"mahjong-connection/game"
	"mahjong-connection/league"
	"mahjong-connection/mq"
	"mahjong-connection/protocal"
	"net"
	"sync"
)

// User 连接用户信息
type User struct {
	UserId  int    // 用户id
	Version string // 客户端版本
	Token   string // 连接时的token

	// 重连用到的客户端当前游戏参数
	ClientRoomId int64 // 客户端房间id
	ClientRound  int   // 客户端局数
	ClientSeq    int   // 客户端最后消息id

	GameConn *net.TCPConn // 游戏服连接

	ClientConn             *net.TCPConn // 客户端连接
	ClientMQ               *mq.MsgQueue
	ClientConnectTime      int64 // 连接时间
	ClientHandshakeTime    int64 // 握手时间
	ClientHandshakeAckTime int64 // 握手完成时间
	ClientHeartBeatTime    int64 // 最后心跳时间

	League *league.League
	Club   *club.Club
	Game   *game.Game

	// 关闭标志
	QuitOnce *sync.Once
}

// NewUser 新建一个用户
func NewUser(userId int, clientConn *net.TCPConn) *User {
	return &User{
		UserId:     userId,
		QuitOnce:   &sync.Once{},
		ClientConn: clientConn,
		ClientMQ:   mq.NewMsgQueue(userId, clientConn),
	}
}

// SendMessageToClient 监听用户消息队列，依次给用户发消息
func (u *User) SendMessageToClient(p *protocal.ImPacket) {
	// 需要用户确认连接之后，才开始推送消息
	if u.ClientHandshakeAckTime > 0 {
		u.ClientMQ.Append(p)
	} else {
		core.Logger.Warn("[SendMessageToClient]用户未handshakeAck，不允许添加消息到消息队列, userId:%v", u.UserId)
	}
}

// WriteMessageToClient 写消息至socket连接
// 某些需要直接给用户发消息的业务，可以跳过消息队列，通过此函数直接发送
func (u *User) WriteMessageToClient(p *protocal.ImPacket) {
	u.ClientMQ.Send(p)
}
