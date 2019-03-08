package user

import (
	"mahjong-league/protocal"
	"net"
	"sync"

	"github.com/fwhappy/util"
)

// User 用户sdf
type User struct {
	ID               int
	HandshakeTime    int64         // 握手时间
	HandshakeAckTime int64         // 握手确认时间
	HeartBeatTime    int64         // 心跳时间
	Info             *Info         // 附加信息
	Mux              *sync.RWMutex // 用户锁
	Conn             *net.TCPConn  // 活动连接
	ConnectTime      int64         // 连接时间
	Mq               *MsgQueue     // 消息队列
	QuitOnce         *sync.Once
	NoHeartbeat      int    // 是否不监听心跳
	From             string // 来源
	Version          string
}

// NewUser 创建一个新用户
func NewUser(id int, conn *net.TCPConn) *User {
	user := &User{}
	user.ID = id
	user.Conn = conn
	user.Info = NewInfo(id)
	user.Mux = &sync.RWMutex{}
	user.Mq = NewMsgQueue(id, conn)
	user.QuitOnce = &sync.Once{}
	user.ConnectTime = util.GetTime()
	return user
}

// SendMessage 监听用户消息队列，依次给用户发消息
func (u *User) SendMessage(imPacket *protocal.ImPacket) {
	// 需要用户确认连接之后，才开始推送消息
	if u.HandshakeAckTime > 0 {
		// u.Mux.Lock()
		// defer u.Mux.Unlock()
		u.Mq.Append(imPacket)
	}
}

// WriteMessage 写消息至socket连接
// 某些需要直接给用户发消息的业务，可以跳过消息队列，通过此函数直接发送
func (u *User) WriteMessage(imPacket *protocal.ImPacket) {
	u.Mq.Send(imPacket)
}
