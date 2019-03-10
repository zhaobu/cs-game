package room

import (
	"sync"

	"github.com/fwhappy/mahjong/setting"
	"github.com/fwhappy/util"
)

// Room 房间信息
type Room struct {
	ID           int64             // 房间id
	Number       string            // 房间号
	GameType     int               // 游戏玩法
	Setting      *setting.MSetting // 扩展玩法
	CreateTime   int64             // 创建时间
	StartTime    int64             // 开始时间
	Status       int               // 房间状态 0:等人中;1:已开始
	CurrentRound int               // 当前局数
	Round        int               // 比赛局数
	CType        int               // 创建类型
	H5Create     bool              // 是否从h5创建

	Users *sync.Map
	Mux   *sync.RWMutex

	LastActiveTime int64 // 房间最后活跃检测时间
}

// NewRoom 新建一个房间
func NewRoom(id int64,
	number string,
	gameType int,
	roomSetting []byte,
	status int,
	createTime,
	startTime int64,
	round int,
	cType int,
	currentRound int) *Room {
	ms := setting.NewMSetting()
	s := make([]int, 0, len(roomSetting))
	for _, v := range roomSetting {
		s = append(s, int(v))
	}
	ms.SetSetting(s)
	room := &Room{
		ID:             id,
		Number:         number,
		GameType:       gameType,
		Setting:        ms,
		Status:         status,
		CreateTime:     createTime,
		StartTime:      startTime,
		Round:          round,
		CurrentRound:   currentRound,
		CType:          cType,
		H5Create:       false,
		Users:          &sync.Map{},
		Mux:            &sync.RWMutex{},
		LastActiveTime: util.GetTime(),
	}
	return room
}

// AddUser 添加房间用户
func (r *Room) AddUser(ru *User) {
	r.Users.Store(ru.ID, ru)
}

// DelUser 删除房间用户
func (r *Room) DelUser(userID int) {
	r.Users.Delete(userID)
}

// HasUser 用户是否在房间内
func (r *Room) HasUser(userID int) bool {
	_, ok := r.Users.Load(userID)
	return ok
}

// GetUserList 获取房间用户列表
func (r *Room) GetUserList() map[int]*User {
	users := make(map[int]*User)
	r.Users.Range(func(k, v interface{}) bool {
		userID := k.(int)
		users[userID] = v.(*User)
		return true
	})
	return users
}

// Start 设置游戏开始
func (r *Room) Start(round int) {
	if round == 1 {
		r.StartTime = util.GetTime()
		r.Status = 1
	}
	r.CurrentRound = round
}
