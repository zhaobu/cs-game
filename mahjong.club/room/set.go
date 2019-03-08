package room

import (
	"sync"

	"github.com/fwhappy/util"
)

// Set 房间集合
type Set struct {
	Mux   *sync.RWMutex
	Rooms *sync.Map
}

// NewSet 新建一个房间集合
func NewSet() *Set {
	set := &Set{}
	set.Mux = &sync.RWMutex{}
	set.Rooms = &sync.Map{}
	return set
}

// Add 添加或替换一个房间
func (s *Set) Add(room *Room) {
	s.Rooms.Store(room.ID, room)
}

// Get 读取一个俱乐部房间
func (s *Set) Get(roomId int64) (*Room, bool) {
	r, ok := s.Rooms.Load(roomId)
	if ok {
		return r.(*Room), ok
	}
	return nil, ok
}

// Del 删除一个房间
func (s *Set) Del(id int64) {
	s.Rooms.Delete(id)
}

// HasRoom 房间是否已存在
func (s *Set) HasRoom(id int64) bool {
	_, ok := s.Rooms.Load(id)
	return ok
}

// Len 取房间数量
func (s *Set) Len() int {
	return util.SMapLen(s.Rooms)
}
