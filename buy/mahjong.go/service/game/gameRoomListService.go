package game

import (
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
)

// RoomList 房间列表
type RoomList struct {
	Rooms *sync.Map
	Mux   *sync.RWMutex
}

// NewRoomList 生成一个大厅列表
func NewRoomList() *RoomList {
	roomMap := &RoomList{}
	roomMap.Rooms = &sync.Map{}
	roomMap.Mux = &sync.RWMutex{}

	return roomMap
}

// SetRoom 添加房间到大厅列表
func (list *RoomList) SetRoom(room *Room) {
	list.Mux.Lock()
	defer list.Mux.Unlock()
	list.Rooms.Store(room.RoomId, room)
}

// GetRoom 从map中获取房间信息
func (list *RoomList) GetRoom(roomId int64) (*Room, *core.Error) {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	if room, ok := list.Rooms.Load(roomId); ok {
		return room.(*Room), nil
	}
	return nil, core.NewError(-300, roomId)
}

// DelRoom 从map中移除用户信息
func (list *RoomList) DelRoom(roomId int64) {
	list.Mux.Lock()
	defer list.Mux.Unlock()
	list.Rooms.Delete(roomId)
}

// Len 从map中移除用户信息
func (list *RoomList) Len() int {
	return util.SMapLen(list.Rooms)
}
