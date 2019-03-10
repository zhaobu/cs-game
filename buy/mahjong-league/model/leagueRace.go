package model

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"sync"

	"github.com/astaxie/beego/orm"
)

// Race 比赛表
type Race struct {
	Id               int64          `orm:"pk"`
	LeagueId         int            //对应场次id
	SignupUserCount  int            `orm:"-"` // 已报名人数
	RequireUserMin   int            `orm:"-"` // 最少报名人数
	RequireUserCount int            `orm:"-"` // 最大报名人数
	SignTime         int64          // 报名开始时间
	GiveupTime       int64          `orm:"-"` // 放弃比赛截止时间
	StartTime        int64          // 比赛开始时间
	Status           int            // 比赛状态 0:报名中;1:排赛中;2:游戏中;3:结算中;4:已结束（正常结束）;5:已解散（人数不足解散）
	Round            int            //当前轮次
	PlanWait         chan int       `orm:"-" json:"-"` // 排赛等待系列
	PlanRooms        map[int64]bool `orm:"-" json:"-"`
	PlanRoomsMux     *sync.RWMutex  `orm:"-" json:"-"`
	RobotJoinStarted bool           `orm:"-" json:"-"` // 是否已经启动机器人加入程序
}

// TableName 数据库表名
func (r *Race) TableName() string {
	return config.TABLE_LEAGUE_RACE
}

// RaceMap 联赛列表
type RaceMap struct {
	Mux  *sync.RWMutex
	Data map[int64]*Race // key: RaceId
}

var (
	// RaceList 联赛列表
	RaceList *RaceMap
	// LeagueRacePlanChannel 待排赛队列
	LeagueRacePlanChannel chan int64
)

func init() {
	orm.RegisterModel(new(Race))
	RaceList = &RaceMap{
		Data: make(map[int64]*Race), //key: raceId, 比赛列表，包括报名中、比赛中、结算中的比赛
		Mux:  &sync.RWMutex{},
	}
	LeagueRacePlanChannel = make(chan int64, 100)
}

// Restore 从数据库恢复报名中的比赛列表
// 脚本启动时调用
func (rl *RaceMap) Restore() {
	var raceList []Race
	_, err := core.GetWriter().
		QueryTable(config.TABLE_LEAGUE_RACE).
		Filter("status__in", config.RACE_STATUS_SIGNUP, config.RACE_STATUS_PLAN, config.RACE_STATUS_PLAY, config.RACE_STATUS_SETTLEMENT).
		All(&raceList)
	if err != nil {
		core.Logger.Error("[Restore]league_race, 从数据库restore数据失败,err:%v", err.Error())
	}
	for i := 0; i < len(raceList); i++ {
		leagueInfo := LeagueList.Get(raceList[i].LeagueId)
		raceList[i].RequireUserMin = leagueInfo.RequireUserMin
		raceList[i].RequireUserCount = leagueInfo.RequireUserCount
		rl.Data[raceList[i].Id] = &raceList[i]
		core.Logger.Trace("[RestoreLeagueRaceList]id:%v, leagueId:%v, status:%v", raceList[i].Id, raceList[i].LeagueId, raceList[i].Status)
	}
	core.Logger.Info("[RestoreLeagueRaceList]completed, count:%v", len(rl.Data))
}

// Get 读取比赛信息
// 获取进行中的比赛，包括报名中、排赛中、比赛中、结算中的比赛
func (rl *RaceMap) Get(raceId int64) *Race {
	rl.Mux.Lock()
	defer rl.Mux.Unlock()
	return rl.Data[raceId]
}

// GetSignup 获取某个LeagueId对应的正在报名中的比赛
func (rl *RaceMap) GetSignup(leagueId int) *Race {
	rl.Mux.Lock()
	defer rl.Mux.Unlock()
	return rl.getSignup(leagueId)
}

// getSignup 获取某个LeagueId对应的正在报名中的比赛, 无锁版本
func (rl *RaceMap) getSignup(leagueId int) *Race {
	for _, raceInfo := range rl.Data {
		if leagueId == raceInfo.LeagueId && raceInfo.Status == config.RACE_STATUS_SIGNUP {
			return raceInfo
		}
	}
	return nil
}

// GetSignupList 获取报名中的比赛列表
// key: leagueId
func (rl *RaceMap) GetSignupList() map[int]*Race {
	rl.Mux.Lock()
	defer rl.Mux.Unlock()

	lists := make(map[int]*Race)
	for _, raceInfo := range rl.Data {
		if raceInfo.Status == config.RACE_STATUS_SIGNUP {
			lists[raceInfo.LeagueId] = raceInfo
		}
	}
	return lists
}

// RestoreLeagueRacePlanChannel 恢复待排赛队列
// 脚本启动时调用
// 必须要在RaceList调用之后
func RestoreLeagueRacePlanChannel() {
	for _, raceInfo := range RaceList.Data {
		if raceInfo.Status == config.RACE_STATUS_PLAN {
			LeagueRacePlanChannel <- raceInfo.Id
			core.Logger.Trace("[RestoreLeagueRacePlanChannel]重新将排赛中的比赛塞入到队列中, raceId:%v", raceInfo.Id)
		}
	}
	core.Logger.Info("[RestoreLeagueRacePlanChannel]completed")
}

// NewRace 生成一个新的比赛
func NewRace(leagueId int) *Race {
	leagueInfo, exists := LeagueList.Data[leagueId]
	if !exists {
		core.Logger.Error("[GetSignupRaceNS]生成新的比赛失败，比赛模板未找到")
		return nil
	}
	// 计算比赛场的三个时间
	signupTime, giveupTime, startTime := leagueInfo.CalcLeagueRaceTime()
	race := &Race{
		LeagueId:         leagueInfo.Id,
		SignupUserCount:  0,
		RequireUserMin:   leagueInfo.RequireUserMin,
		RequireUserCount: leagueInfo.RequireUserCount,
		SignTime:         signupTime,
		GiveupTime:       giveupTime,
		StartTime:        startTime,
		Status:           config.RACE_STATUS_SIGNUP,
		Round:            1,
		RobotJoinStarted: false,
	}
	id, err := core.GetWriter().Insert(race)
	if err != nil {
		core.Logger.Error("[NewRace]插入表失败,err:%v", err.Error())
	}
	race.Id = id
	return race
}

// GetSignupNS 获取某个LeagueId对应的正在报名中的比赛
// 如果比赛不存在，则生成一条新的
func (rl *RaceMap) GetSignupNS(leagueId int) (*Race, bool) {
	race := rl.getSignup(leagueId)
	newFlag := false
	if race == nil {
		if race = NewRace(leagueId); race == nil {
			return nil, newFlag
		}
		newFlag = true
	}
	rl.Data[race.Id] = race
	core.Logger.Debug("[GetSignupRaceNS]生成一个新的比赛, raceId:%v, leagueId:%v", race.Id, race.LeagueId)

	return race, newFlag
}

// IsFull 判断比赛报名人数是否已达上限
func (r *Race) IsFull() bool {
	return r.SignupUserCount >= r.RequireUserCount
}

// IsEnough 判断比赛报名人数是否已足够开赛
func (r *Race) IsEnough() bool {
	return r.SignupUserCount >= r.RequireUserMin
}

// Update LeagueReace 表更新操作
func (r *Race) Update(o orm.Ormer, cols ...string) (int64, error) {
	if o == nil {
		o = core.GetWriter()
	}
	num, err := o.Update(r, cols...)
	if err != nil {
		core.Logger.Debug("[leagueRace.Update]leagueId:%v, raceId:%v, error:%v", r.LeagueId, r.Id, err.Error())
	}
	return num, err
}

// GetPlanWaitLen 获取当前等候排赛的房间数
func (r *Race) GetPlanWaitLen() int {
	r.PlanRoomsMux.RLock()
	defer r.PlanRoomsMux.RUnlock()
	return len(r.PlanRooms)
}

// DelPlanWait 删除一个等候排赛中的房间，如果全部删完了，则认为排赛完成了
func (r *Race) DelPlanWait(raceRoomId int64) {
	r.PlanRoomsMux.Lock()
	defer r.PlanRoomsMux.Unlock()
	delete(r.PlanRooms, raceRoomId)
	// 所有房间完成
	if len(r.PlanRooms) == 0 {
		r.PlanWait <- 1
	}
}

// IsCompleted 是否已完成
func (r *Race) IsCompleted() bool {
	return r.Status == config.RACE_STATUS_SETTLEMENT &&
		r.Round == LeagueList.Get(r.LeagueId).GetTotalRound()
}

// IsLastRound 是否最后一轮
func (r *Race) IsLastRound() bool {
	return r.Round == LeagueList.Get(r.LeagueId).GetTotalRound()
}

// IsRunning 是否正在进行中的比赛
func (r *Race) IsRunning() bool {
	return r.Status == config.RACE_STATUS_SIGNUP ||
		r.Status == config.RACE_STATUS_PLAN ||
		r.Status == config.RACE_STATUS_PLAY ||
		r.Status == config.RACE_STATUS_SETTLEMENT
}
