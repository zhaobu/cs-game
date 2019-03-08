package mq

import (
	"mahjong-connection/core"
	"mahjong-connection/protocal"
	"net"
	"sync"

	"github.com/fwhappy/util"
)

const (
	MQ_STATUS_NOT_OPEN = 0 // 未开启
	MQ_STATUS_OK       = 1 // 正常
	MQ_STATUS_PAUSE    = 2 // 暂停
	MA_STATUS_CLOSE    = 3 // 已终止
)

// MsgQueue 消息队列
type MsgQueue struct {
	UserId       int                     // 用户Id
	Conn         *net.TCPConn            // socket连接
	queue        chan *protocal.ImPacket // 消息队列
	pauseChan    chan int                // 暂停队列
	continueChan chan int                // 继续队列
	status       int                     // 状态，0:未开启；1:正常；2：暂停; 3:已中止
	closeOnce    *sync.Once
}

// NewMsgQueue 新建消息队列
func NewMsgQueue(userId int, conn *net.TCPConn) *MsgQueue {
	q := &MsgQueue{
		UserId:       userId,
		Conn:         conn,
		queue:        make(chan *protocal.ImPacket, 1024),
		pauseChan:    make(chan int, 1),
		continueChan: make(chan int, 1),
		status:       MQ_STATUS_NOT_OPEN,
		closeOnce:    &sync.Once{},
	}
	return q
}

// Append 追加一条消息到队列
func (q *MsgQueue) Append(imPacket *protocal.ImPacket) {
	if q.status != MQ_STATUS_OK && q.status != MQ_STATUS_PAUSE {
		// core.Logger.Warn("[mq.append]status error, UserID:%v, status:%v", q.UserID, q.status)
		return
	}
	select {
	case q.queue <- imPacket:
	default:
		core.Logger.Warn("[mq.append]queue is full, UserID:%v", q.UserId)
	}
}

// Send 直接发送消息
func (q *MsgQueue) Send(imPacket *protocal.ImPacket) {
	if _, err := q.Conn.Write(imPacket.Serialize()); err != nil {
		core.Logger.Error("user.Conn.Write error, UserID:%#v, packageId:%d,messageId:%d,messageType:%d,length:%d,err:%v",
			q.UserId, int(imPacket.GetPackage()), int(imPacket.GetMessageId()), int(imPacket.GetMessageType()), len(imPacket.Serialize()), err.Error())
	}
	if q.UserId == 104593 {
		core.Logger.Error("Send, UserID:%#v, packageId:%d,messageId:%d,messageType:%d,length:%d",
			q.UserId, int(imPacket.GetPackage()), int(imPacket.GetMessageId()), int(imPacket.GetMessageType()), len(imPacket.Serialize()))
	}
}

// Start 开始推送
func (q *MsgQueue) Start() {
	// 捕获异常
	defer util.RecoverPanic()

	if q.status > MQ_STATUS_NOT_OPEN {
		core.Logger.Debug("[mq.start]消息推送已开启，忽略本次操作, UserID:%v", q.UserId)
		return
	}
	// 设置为开启状态
	q.status = MQ_STATUS_OK
	core.Logger.Debug("[mq.start]UserID:%v", q.UserId)
	closed := false
	for {
		select {
		case imPacket := <-q.queue:
			if imPacket == nil {
				closed = true
				break
			}
			q.Send(imPacket)
		case v := <-q.pauseChan: // 暂停消息
			if v == 0 {
				closed = true
				break
			}
			core.Logger.Debug("[mq.pause]UserID:%v", q.UserId)
			<-q.continueChan // 等候继续消息
			core.Logger.Debug("[mq.continue]UserID:%v", q.UserId)
			break
		}
		if closed {
			break
		}
	}
	core.Logger.Debug("[mq.stop]UserID:%v", q.UserId)
}

// Close 关闭消息队列
func (q *MsgQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.queue)
		close(q.pauseChan)
		close(q.continueChan)
		core.Logger.Debug("[mq.close]UserID:%v", q.UserId)
	})
}

// Pause 暂停
func (q *MsgQueue) Pause() {
	if q.status != MQ_STATUS_OK {
		core.Logger.Warn("[mq.pause]status error, UserID:%v, status:%v", q.UserId, q.status)
		return
	}
	select {
	case q.pauseChan <- 1:
		q.status = MQ_STATUS_PAUSE
	default:
		core.Logger.Warn("[mq.pause]pauseChan is full, UserID:%v", q.UserId)
	}
}

// Continue 继续
func (q *MsgQueue) Continue() {
	if q.status != MQ_STATUS_PAUSE {
		core.Logger.Warn("[mq.continue]status error, UserID:%v, status:%v", q.UserId, q.status)
		return
	}
	select {
	case q.continueChan <- 1:
		q.status = MQ_STATUS_OK
	default:
		core.Logger.Warn("[mq.continue]continueChan is full, UserID:%v", q.UserId)
	}
}

// GetStatus 获取当前队列状态
func (q *MsgQueue) GetStatus() int {
	return q.status
}

// WasStarted 消息队列是否已开启
func (q *MsgQueue) WasStarted() bool {
	return q.status > MQ_STATUS_NOT_OPEN
}

// WasPaused 是否处于暂停状态
func (q *MsgQueue) WasPaused() bool {
	return q.status == MQ_STATUS_PAUSE
}
