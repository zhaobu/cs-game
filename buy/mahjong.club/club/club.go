package club

import (
	"sync"

	"mahjong.club/message"
	"mahjong.club/room"
)

// Club 俱乐部信息
type Club struct {
	ID            int       // 俱乐部id
	Name          string    // 俱乐部名称
	Users         *sync.Map // 俱乐部在线用户信息
	RoomSet       *room.Set
	Mux           *sync.RWMutex // 俱乐部锁
	UMux          *sync.RWMutex // 用户锁
	LastMessageID uint64        // 最后消息id
	ML            *message.MList
}

// NewClub 生成一个新俱乐部
func NewClub(id int) *Club {
	c := &Club{ID: id, Mux: &sync.RWMutex{}}
	c.Users = &sync.Map{}
	c.RoomSet = room.NewSet()
	c.Mux = &sync.RWMutex{}
	c.UMux = &sync.RWMutex{}
	c.ML = message.NewMList()
	return c
}

// NextMessageID 生成下一条消息id
func (c *Club) NextMessageID() uint64 {
	c.LastMessageID++
	return c.LastMessageID
}
