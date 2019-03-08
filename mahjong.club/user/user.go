package user

import (
	"net"
	"sync"

	"github.com/fwhappy/mahjong/protocal"
	"github.com/fwhappy/util"
	"mahjong.club/core"
)

// User 用户
type User struct {
	ID            int
	HandshakeTime int64                   // 握手时间
	HeartBeatTime int64                   // 心跳时间
	Info          *Info                   // 附加信息
	Mux           *sync.RWMutex           // 用户锁
	Conn          *net.TCPConn            // 活动连接
	Mq            chan *protocal.ImPacket // 消息队列
	MStatus       bool                    // 消息状态
	Clubs         []int                   // 所属俱乐部id
	QuitOnce      *sync.Once
	NoHeartbeat   int    // 是否监听用户心跳
	From          string // 用户来源
}

// NewUser 创建一个新用户
func NewUser(id int) *User {
	user := &User{}
	user.ID = id
	user.Info = NewInfo(id)
	user.Mux = &sync.RWMutex{}
	user.Mq = make(chan *protocal.ImPacket, 1024)
	user.MStatus = true
	user.Clubs = make([]int, 0, 1)
	user.QuitOnce = &sync.Once{}
	return user
}

// SendMessage 监听用户消息队列，依次给用户发消息
func (u *User) SendMessage() {
	// 捕获异常
	defer util.RecoverPanic()

	for imPacket := range u.Mq {
		u.WriteMessage(imPacket)
	}
}

// WriteMessage 写消息至socket连接
// 某些需要直接给用户发消息的业务，可以跳过消息队列，通过此函数直接发送
func (u *User) WriteMessage(imPacket *protocal.ImPacket) {
	if _, err := u.Conn.Write(imPacket.Serialize()); err != nil {
		core.Logger.Error("消息发送失败, userId:%v, packageId:%v,messageId:%v,messageType:%v,length:%v",
			u.ID, imPacket.GetPackage(), imPacket.GetMessageId(), imPacket.GetMessageType(), len(imPacket.Serialize()))
		core.Logger.Error("u.Conn.Write error: %s.", err.Error())
	} else {
		// core.Logger.Trace("消息发送成功, userId:%v, packageId:%v,messageId:%v,messageType:%v,length:%v",
		// 	u.ID, imPacket.GetPackage(), imPacket.GetMessageId(), imPacket.GetMessageType(), len(imPacket.Serialize()))
	}
}

// AppendMessage 添加一条消息到消息队列
func (u *User) AppendMessage(imPacket *protocal.ImPacket) {
	if u.MStatus {
		u.Mq <- imPacket
	}
}
