package model

import (
	"encoding/json"
	"mahjong-league/config"
	"mahjong-league/core"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

var (
	// RaceRoomsList 所有比赛队员的房间列表
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
	ActiveTime   int64  `orm:"-"` // 房间最后活跃时间
}

// TableName 数据库真实表名
func (rr *RaceRoom) TableName() string {
	return config.TABLE_LEAGUE_RACE_ROOM
}

// RaceRoomsMap 某比赛对应的房间列表
type RaceRoomsMap struct {
	Mux  *sync.RWMutex
	Data map[int64]*RaceRoom
}

// GetRaceRooms 获取比赛的房间列表
func GetRaceRooms(raceId int64) *RaceRoomsMap {
	if v, ok := RaceRoomsList.Load(raceId); ok {
		return v.(*RaceRoomsMap)
	}
	return nil
}

// GetRaceRoom 获取比赛的房间
func GetRaceRoom(raceId, raceRoomId int64) *RaceRoom {
	if v, ok := RaceRoomsList.Load(raceId); ok {
		return v.(*RaceRoomsMap).Get(raceRoomId)
	}
	return nil
}

// NewRaceRooms 获取比赛的房间列表 如果列表不存在，则新加一条
func NewRaceRooms(raceId int64) *RaceRoomsMap {
	raceRooms := &RaceRoomsMap{
		Mux:  &sync.RWMutex{},
		Data: make(map[int64]*RaceRoom),
	}
	return raceRooms
}

// New 新建一个联赛房间
func (raceRooms *RaceRoomsMap) New(raceInfo *Race, userIds []int, o orm.Ormer) *RaceRoom {
	userString, _ := util.InterfaceToJsonString(userIds)
	raceRoom := &RaceRoom{
		RaceId:     raceInfo.Id,
		Round:      raceInfo.Round,
		Status:     config.RACE_ROOM_STATUS_NORMAL,
		CreateTime: util.GetTime(),
		UpdateTime: util.GetTime(),
		Users:      userString,
	}
	id, err := o.Insert(raceRoom)
	if err != nil {
		core.Logger.Error("[RaceRoomsMap.New]写入league_room表失败,raceId:%v, err:%v", raceInfo.Id, err.Error())
	}
	raceRoom.Id = id
	raceRooms.Data[id] = raceRoom

	return raceRoom
}

// RestoreRaceRooms 启动时恢复进行中的比赛房间信息
func RestoreRaceRooms() {
	for _, race := range RaceList.Data {
		// 新建一个比赛房间列表
		raceRooms := NewRaceRooms(race.Id)
		// 从数据库读取RaceRooms
		rooms := loadRaceRoomsFromDB(race.Id)
		for i := 0; i < len(rooms); i++ {
			raceRooms.Data[rooms[i].Id] = &rooms[i]
		}
		RaceRoomsList.Store(race.Id, raceRooms)
		core.Logger.Debug("[RestoreRaceRooms]raceId:%v, room count:%v", race.Id, len(raceRooms.Data))
	}
	core.Logger.Debug("[RestoreRaceRooms]completed, race count:%v", len(RaceList.Data))
}

func loadRaceRoomsFromDB(raceId int64) []RaceRoom {
	var rooms []RaceRoom
	_, err := core.GetWriter().QueryTable(config.TABLE_LEAGUE_RACE_ROOM).Filter("race_id", raceId).All(&rooms)
	if err != nil {
		core.Logger.Error("[loadRaceRoomsFromDB]从数据库读取league_room失败, raceId:%v, err:%v", raceId, err.Error())
	}
	return rooms
}

// GetUsers 获取房间用户
func (rr *RaceRoom) GetUsers() []int {
	users := []int{}
	json.Unmarshal([]byte(rr.Users), &users)
	return users
}

// Get 获取比赛用户信息
func (raceRooms *RaceRoomsMap) Get(raceRoomId int64) *RaceRoom {
	raceRooms.Mux.RLock()
	defer raceRooms.Mux.RUnlock()
	return raceRooms.Data[raceRoomId]
}

// Set 获取比赛用户信息
func (raceRooms *RaceRoomsMap) Set(rr *RaceRoom) {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	raceRooms.Data[rr.Id] = rr
}

// IsCompleted 判断所有比赛是否都已经结束
func (raceRooms *RaceRoomsMap) IsCompleted() bool {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	for _, raceRoom := range raceRooms.Data {
		if raceRoom.Status == config.RACE_ROOM_STATUS_NORMAL {
			return false
		}
	}
	return true
}

// PlayingCount 获取进行中的房间数量
func (raceRooms *RaceRoomsMap) PlayingCount() int {
	raceRooms.Mux.Lock()
	defer raceRooms.Mux.Unlock()
	cnt := 0
	for _, raceRoom := range raceRooms.Data {
		if raceRoom.Status == config.RACE_ROOM_STATUS_NORMAL {
			cnt++
		}
	}
	return cnt
}

// IsActive 判断房间是否活跃
// 3分钟内算活跃
func (rr *RaceRoom) IsActive() bool {
	return rr.ActiveTime > int64(0) && util.GetTime()-rr.ActiveTime <= int64(180)
}
