package mq

import (
	"net"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
)

// MsgQueue 消息队列
type MsgQueue struct {
	UserId       int                     // 用户id
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
		status:       0,
		closeOnce:    &sync.Once{},
	}
	return q
}

// Append 追加一条消息到队列
func (q *MsgQueue) Append(imPacket *protocal.ImPacket) {
	if q.status != 1 && q.status != 2 {
		core.Logger.Warn("[mq.append]status error, userId:%v, status:%v", q.UserId, q.status)
		return
	}
	select {
	case q.queue <- imPacket:
	default:
		core.Logger.Warn("[mq.append]queue is full, userId:%v", q.UserId)
	}
}

// Send 直接发送消息
func (q *MsgQueue) Send(imPacket *protocal.ImPacket) {
	if _, err := q.Conn.Write(imPacket.Serialize()); err != nil {
		core.Logger.Errorf("user.Conn.Write error, userId:%#v, packageId:%d,messageId:%d,messageType:%d,length:%d,err:%v",
			q.UserId, int(imPacket.GetPackage()), int(imPacket.GetMessageId()), int(imPacket.GetMessageType()), len(imPacket.Serialize()), err.Error())
	}
}

// Start 开始推送
func (q *MsgQueue) Start() {
	if q.status > 0 {
		core.Logger.Debug("[mq.start]消息推送已开启，忽略本次操作, userId:%v", q.UserId)
		return
	}
	// 捕获异常
	defer util.RecoverPanic()

	// 设置为开启状态
	q.status = 1
	core.Logger.Debug("[mq.start]userId:%v", q.UserId)
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
			core.Logger.Debug("[mq.pause]userId:%v", q.UserId)
			<-q.continueChan // 等候继续消息
			core.Logger.Debug("[mq.continue]userId:%v", q.UserId)
			break
		}
		if closed {
			break
		}
	}
	core.Logger.Debug("[mq.stop]userId:%v", q.UserId)
}

// Close 关闭消息队列
func (q *MsgQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.queue)
		close(q.pauseChan)
		close(q.continueChan)
		core.Logger.Debug("[mq.close]userId:%v", q.UserId)
	})
}

// Pause 暂停
func (q *MsgQueue) Pause() {
	if q.status != 1 {
		core.Logger.Warn("[mq.pause]status error, userId:%v, status:%v", q.UserId, q.status)
		return
	}
	select {
	case q.pauseChan <- 1:
		q.status = 2
	default:
		core.Logger.Warn("[mq.pause]pauseChan is full, userId:%v", q.UserId)
	}
}

// Continue 继续
func (q *MsgQueue) Continue() {
	if q.status != 2 {
		core.Logger.Warn("[mq.continue]status error, userId:%v, status:%v", q.UserId, q.status)
		return
	}
	select {
	case q.continueChan <- 1:
		q.status = 1
	default:
		core.Logger.Warn("[mq.continue]continueChan is full, userId:%v", q.UserId)
	}
}

// GetStatus 获取当前队列状态
func (q *MsgQueue) GetStatus() int {
	return q.status
}

// WasStarted 消息队列是否已开启
func (q *MsgQueue) WasStarted() bool {
	return q.status > 0
}

// WasPaused 是否处于暂停状态
func (q *MsgQueue) WasPaused() bool {
	return q.status == 2
}
