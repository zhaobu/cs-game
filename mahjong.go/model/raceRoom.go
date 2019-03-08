package model

import (
	"encoding/json"
	"sync"

	"github.com/astaxie/beego/orm"
	"mahjong.go/library/core"
)

var (
	// RaceRoomsList 所有比赛的房间列表
	RaceRoomsList *sync.Map // key: raceId, value: *RaceRoomsMap
)

func init() {
	orm.RegisterModel(new(RaceRoom))
	RaceRoomsList = &sync.Map{}
}

// RaceRoom 联赛用户表
type RaceRoom struct {
	Id           int64  `orm:"pk"`
	RaceId       int64  // 比赛id
	RoomId       int64  // 房间id
	Round        int    // 轮次
	Users        string // json string, 参与用户列表, [1,2,3]
	Status       int    // 状态,0:游戏中;1:已结束;2:异常结束
	ServerRemote string // 房间所在游戏服
	CreateTime   int64  // 创建时间
	UpdateTime   int64  // 最后更新时间，可用作结束时间
}

// TableName 数据库真实表名
func (rr *RaceRoom) TableName() string {
	return "league_room"
}

// RaceRoomsMap 某比赛对应的房间列表
type RaceRoomsMap struct {
	Mux  *sync.RWMutex
	Data map[int64][]int // key: roomId; value: users
}

// GetRaceRooms 获取比赛的房间列表
func GetRaceRooms(raceId int64) *RaceRoomsMap {
	if v, ok := RaceRoomsList.Load(raceId); ok {
		return v.(*RaceRoomsMap)
	}
	return nil
}

// GetRaceRoomsNS 获取比赛的房间列表
func GetRaceRoomsNS(raceId int64) *RaceRoomsMap {
	raceRooms := GetRaceRooms(raceId)
	if raceRooms == nil {
		raceRooms = &RaceRoomsMap{
			Mux:  &sync.RWMutex{},
			Data: make(map[int64][]int),
		}
		RaceRoomsList.Store(raceId, raceRooms)
	}
	return raceRooms
}

// Get 获取比赛房间的用户列表
func (raceRooms *RaceRoomsMap) Get(roomId int64) []int {
	raceRooms.Mux.RLock()
	defer raceRooms.Mux.RUnlock()
	return raceRooms.Data[roomId]
}

// Del 删除比赛的房间
func (raceRooms *RaceRoomsMap) Del(roomId int64) {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	delete(raceRooms.Data, roomId)
}

// Set 获取比赛用户信息
func (raceRooms *RaceRoomsMap) Set(roomId int64, userIds []int) {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	raceRooms.Data[roomId] = userIds
}

// GetRoomUserIds 获取比赛用户信息
func (raceRooms *RaceRoomsMap) GetRoomUserIds() []int {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	users := make([]int, 0)
	for _, ids := range raceRooms.Data {
		users = append(users, ids...)
	}
	return users
}

// GetRaceRoom 获取比赛房间
func GetRaceRoom(id int64) *RaceRoom {
	rr := &RaceRoom{Id: id}
	core.GetWriter().Read(rr)
	if rr.RaceId == 0 {
		return nil
	}
	return rr
}

// GetUsers 读取房间参与用户
func (rr *RaceRoom) GetUsers() []int {
	users := make([]int, 0)
	json.Unmarshal([]byte(rr.Users), &users)
	return users
}
