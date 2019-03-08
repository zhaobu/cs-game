package message

import (
	"container/list"

	"mahjong.club/config"
)

// MList 消息列表
type MList struct {
	// 前进后出，即最新的消息在左边
	list *list.List
}

// NewMList 新建一个消息列表
func NewMList() *MList {
	return &MList{
		list: list.New(),
	}
}

// Size 消息列表长度
func (ml *MList) Size() int {
	return ml.list.Len()
}

// GetList 从列表中获取指定长度的消息
func (ml *MList) GetList(msgID uint64, limit int) []*Msg {
	msgList := make([]*Msg, 0, limit)
	for m := ml.list.Front(); m != nil; m = m.Next() {
		msg := m.Value.(Msg)
		if msgID == 0 || msg.MID < msgID {
			msgList = append(msgList, &msg)
		}
		if len(msgList) >= limit {
			break
		}
		if msg.IsTimeout() {
			break
		}
	}
	return msgList
}

// Add 添加一条消息到消息队列
func (ml *MList) Add(m *Msg) {
	// 如果队列条数已满，则删除最后一个元素
	if ml.list.Len() >= config.CLUB_MESSAGE_LIST_LENGTH {
		ml.list.Remove(ml.list.Back())
	}
	ml.list.PushFront(*m)
}
