package model

import (
	"encoding/json"
	"mahjong-league/config"
	"mahjong-league/core"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

// LeagueList 联赛模板列表
var LeagueList *LeagueMap

func init() {
	orm.RegisterModel(new(League))
	LeagueList = &LeagueMap{
		Mux:  &sync.RWMutex{},
		Data: make(map[int]*League),
	}
}

// League 联赛模板
type League struct {
	Id                int `orm:"pk"`
	Name              string
	Img               string
	Icon              int
	GameType          int     // 游戏类型
	Setting           string  // 游戏设置
	Rounds            string  // 比赛流程,[[晋级类型，晋级人数，设定局数],...]
	RequireUserMin    int     // 最少报名人数
	RequireUserCount  int     // 需要报名人数
	RequireVerMin     string  // 版本要求
	PriceEntityId     int     // 价格对应的entityId
	Price             int     // 报名费用
	LeagueType        int     // 比赛类型
	StartCondition    int     // 开始条件，1：非定时塞；2：定时赛
	Cycle             int     // 循环模式，0=非循环赛 1=每天  2=每周  3=每月
	CycleVal          int     // 循环值
	Status            int     // 是否有效，0=无效 1=有效
	OpenTime          int64   // 开放时间戳
	CloseTime         int64   // 关闭时间戳
	UpdateTime        int64   // 最后修改时间
	SignTime          int64   // 报名开始时间，距当天0点秒数
	StartTime         int64   // 比赛开始时间，距当天0点秒数
	Weight            int     // 排名权重
	EnableRobot       int     // 是否支持机器人,0=不支持;1=支持
	RobotJoinInterval int     // 机器人加入间隔
	FwScores          string  // 机器人氛围积分配置
	SkipRank          int     // 跳过多少个排名，即前多少名不参与排名，永远不会真的发奖励
	Category          int     // 比赛场类别
	fwScores          [][]int `orm:"-"` // 机器人氛围积分配置，setting的解析版
	setting           []int   `orm:"-"` // setting的解析版
	rounds            [][]int `orm:"-"` // rounds的展开版

	// 模拟相关参数
	SimulationUserCount int `orm:"-"` // 上次模拟数量
}

// TableName 数据库表名
func (l *League) TableName() string {
	return config.TABLE_LEAGUE_LIST
}

// LeagueMap 联赛配置列表
type LeagueMap struct {
	Mux  *sync.RWMutex
	Data map[int]*League
}

// IsOpen 判断联赛是否处于开放状态
func (l *League) IsOpen(t int64) bool {
	if l.Status == config.LEAGUE_STATUS_CLOSE {
		return false
	}
	if t == 0 {
		t = util.GetTime()
	}
	if l.OpenTime > t || // 未开始
		(l.CloseTime > 0 && l.CloseTime < t) { // 已结束
		return false
	}
	return true
}

// GetOpenList 获取所有开放中的联赛列表
func (list *LeagueMap) GetOpenList() map[int]*League {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	openList := make(map[int]*League)
	for id, leagueInfo := range list.Data {
		if leagueInfo.IsOpen(0) {
			openList[id] = leagueInfo
		}
	}
	return openList
}

// Restore 重新读取联赛模板
// 用于重启后恢复数据
func (list *LeagueMap) Restore() {
	data := make(map[int]*League)
	result := LoadLeagueListFromDB()
	for i := 0; i < len(result); i++ {
		data[result[i].Id] = &result[i]
		core.Logger.Info("[RestoreLeagueList]id:%v", result[i].Id)
	}
	list.Data = data
	core.Logger.Info("[RestoreLeagueList]completed, count:%v", len(list.Data))
}

// LoadLeagueListFromDB 从数据库读取所有的联赛列表
func LoadLeagueListFromDB() []League {
	var result []League
	_, err := core.GetWriter().QueryTable(config.TABLE_LEAGUE_LIST).All(&result)
	if err != nil {
		core.Logger.Error("[LoadLeagueListFromDB]从数据库回复league_list数据失败,err:%v", err.Error())
	}
	return result
}

// Get 获取比赛模板信息
func (list *LeagueMap) Get(leagueId int) *League {
	list.Mux.Lock()
	defer list.Mux.Unlock()
	if leagueInfo, ok := list.Data[leagueId]; ok {
		return leagueInfo
	}
	return nil
}

// Set 添加比赛模板信息
func (list *LeagueMap) Set(l *League) {
	list.Mux.Lock()
	defer list.Mux.Unlock()
	list.Data[l.Id] = l
}

// CalcLeagueRaceTime 计算比赛的下次时间
// 非定时赛，全是0
// 定时赛，如果当前循环周期比赛未开始，返回当前循环周期的时间
// 如果已开始，则返回下个循环周期的时间
func (l *League) CalcLeagueRaceTime() (signupTime, giveupTime, startTime int64) {
	// 非定时赛
	if l.StartCondition == config.LEAGUE_START_CONDITION_TEMP {
		return
	}

	// 当前时间
	t := util.GetTime()

	// 获取当前循环时间
	var currentLoopTime int64
	// 偏移时间
	var offsetTime int64
	year, month, day := time.Now().Date()
	weekday := util.GetChinaWeekDay()
	// weekday := util.GetTime()
	switch l.Cycle {
	case config.LEAGUE_CYCLE_DAY:
		currentLoopTime = time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
		offsetTime = int64(86400)
	case config.LEAGUE_CYCLE_WEEK:
		currentLoopTime = time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
		currentLoopTime -= int64((weekday - 1) * 86400)
		currentLoopTime += int64((l.CycleVal - 1) * 86400)
		offsetTime = int64(7 * 86400)
	case config.LEAGUE_CYCLE_MONTH:
		thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		currentLoopTime = thisMonth.Unix() + int64((l.CycleVal-1)*86400)
		offsetTime = thisMonth.AddDate(0, 1, 0).Unix() - thisMonth.Unix()
	}
	// 比赛开始时间
	signupTime = currentLoopTime + l.SignTime
	startTime = currentLoopTime + l.StartTime
	giveupTime = startTime - config.RACE_FOBBIDEN_GIVEUP_SECOND

	// 比赛已开始，则加上偏移值，返回下个周期的时间
	if t >= startTime {
		signupTime += offsetTime
		giveupTime += offsetTime
		startTime += offsetTime
	}

	// core.Logger.Trace("计算比赛的时间, 报名时间:%v, 取消报名截止时间:%v, 开始时间:%v", util.FormatUnixTime(signupTime), util.FormatUnixTime(giveupTime), util.FormatUnixTime(startTime))

	return
}

// GetSetting 获取游戏配置
func (l *League) GetSetting() []int {
	if l.setting == nil || len(l.setting) == 0 {
		l.setting = make([]int, 0)
		json.Unmarshal([]byte(l.Setting), &l.setting)
	}
	return l.setting
}

// GetRounds 获取晋级配置
func (l *League) GetRounds() [][]int {
	if l.rounds == nil || len(l.rounds) == 0 {
		l.rounds = make([][]int, 0)
		err := json.Unmarshal([]byte(l.Rounds), &l.rounds)
		if err != nil {
			core.Logger.Error("[l.GetRounds]err:%v", err.Error())
		}
	}
	return l.rounds
}

// GetRoundRequireUserCount 获取round对应的需要人数
func (l *League) GetRoundRequireUserCount(round int) int {
	return l.GetRounds()[round-1][1]
}

// GetGameRound 获取游戏局数
func (l *League) GetGameRound(round int) int {
	return l.GetRounds()[round-1][2]
}

// GetTotalRound 获取总轮次
func (l *League) GetTotalRound() int {
	return len(l.GetRounds())
}

// GetRoomUserCount 获取房间用户
func (l *League) GetRoomUserCount() int {
	return l.GetSetting()[9]
}

// EnableAutoApply 是否支持自动加入
func (l *League) EnableAutoApply() bool {
	// return true
	return l.EnableRobot > 0 && l.StartCondition == config.LEAGUE_START_CONDITION_FIXED
}

// EnableSimulationUserCount 是否支持自动加入
func (l *League) EnableSimulationUserCount() bool {
	// return true
	return l.EnableRobot > 0 && l.StartCondition == config.LEAGUE_START_CONDITION_TEMP
}

// CheckVersion 版本是否可以进入
func (l *League) CheckVersion(version string) bool {
	return strings.Compare(version, l.RequireVerMin) > -1
}

// GetRobotFWScores 获取比赛的机器人氛围积分
func (l *League) GetRobotFWScores() [][]int {
	if l.GetTotalRound() == 1 || l.EnableRobot == 0 {
		return [][]int{}
	}

	if len(l.FwScores) == 0 {
		return [][]int{}
	}

	if l.fwScores == nil || len(l.fwScores) == 0 {
		l.fwScores = make([][]int, 0)

		err := json.Unmarshal([]byte(l.FwScores), &l.fwScores)
		if err != nil {
			core.Logger.Error("[l.GetRobotFWScores]err:%v", err.Error())
		}
	}

	return l.fwScores
}
