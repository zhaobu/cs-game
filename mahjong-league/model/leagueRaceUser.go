package model

import (
	"encoding/json"
	"fmt"
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/robot"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

var (
	// RaceUsersList 所有比赛的用户列表的集合, key: raceId, value: RaceUsers
	RaceUsersList *sync.Map
	// UserRaceContainer 用户的比赛id对照表, key: userId, value: raceId
	UserRaceContainer *sync.Map
)

func init() {
	orm.RegisterModel(new(RaceUser))
	RaceUsersList = &sync.Map{}
	UserRaceContainer = &sync.Map{}
}

// RaceUser 联赛用户表
type RaceUser struct {
	Id           int64       `orm:"pk"`
	RaceId       int64       // 比赛id
	UserId       int         // 用户id
	Round        int         // 轮次
	Status       int         // 用户状态, 0: 游戏中（已报名）;1: 已退赛; 2: 已淘汰; 5: 已解散
	Score        int         // 用户分数
	Price        int         // 消耗钻石
	GiveupTime   int64       // 退赛时间
	SignTime     int64       //  报名时间
	FailTime     int64       // 淘汰时间
	FwScores     string      `orm:"column(fw_score)"` // 氛围积分
	Rank         int         `orm:"-"`                // 用户排名
	RoomId       int64       `orm:"-"`                // 用户房间
	GiveupStatus int         `orm:"-"`                // 是否允许用户退赛, 0: 不允许; 允许
	Consume      map[int]int `orm:"-"`                // 支付金额
	fwScores     []int       `orm:"-"`                // 机器人氛围积分配置，解析版
}

// TableName RaceUser的数据库表名
func (ru *RaceUser) TableName() string {
	return config.TABLE_LEAGUE_RACE_USER
}

// RaceUsers 比赛用户列表
type RaceUsers struct {
	RaceId int64 // 比赛id
	Mux    *sync.RWMutex
	Users  map[int]*RaceUser // key: userId
}

// NewRaceUser 插入用户比赛信息
func NewRaceUser(userId int, raceId int64, price int) (*RaceUser, error) {
	raceUser := &RaceUser{
		RaceId:     int64(raceId),
		UserId:     userId,
		Round:      1,
		Status:     config.RACE_USER_STATUS_SIGNUP,
		Score:      config.RACE_USER_SCORE_BASE,
		SignTime:   util.GetTime(),
		GiveupTime: 0,
		FailTime:   0,
		Price:      price,
	}
	id, err := core.GetWriter().Insert(raceUser)
	if err != nil {
		core.Logger.Error("[NewRaceUser]插入表失败,err:%v", err.Error())
	}
	raceUser.Id = id
	return raceUser, err
}

// New 添加用户比赛信息
// 同时会插入数据库
func (raceUsers *RaceUsers) New(userId int, raceId int64, price int) (*RaceUser, error) {
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()

	raceUser, _ := NewRaceUser(userId, raceId, price)
	raceUsers.Users[userId] = raceUser
	SetUserRace(userId, raceId)
	return raceUser, nil
}

// Update LeagueReace 表更新操作
func (ru *RaceUser) Update(o orm.Ormer, cols ...string) (int64, error) {
	if o == nil {
		o = core.GetWriter()
	}
	num, err := o.Update(ru, cols...)
	if err != nil {
		core.Logger.Debug("[leagueRaceUser.Update]raceId:%v, userId:%v, error:%v", ru.Id, ru.UserId, err.Error())
	}
	return num, err
}

// ToDismiss 解散比赛用户
func (raceUsers *RaceUsers) ToDismiss(raceId int64) {
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	for _, raceUser := range raceUsers.Users {
		// 更改用户状态
		raceUser.Status = config.RACE_USER_STATUS_DISMISS
		// 删除用户比赛对应关系
		DelUserRace(raceUser.UserId)
	}
	// 更新数据库
	num, err := core.GetWriter().
		QueryTable(config.TABLE_LEAGUE_RACE_USER).
		Filter("race_id", raceId).Update(
		orm.Params{
			"status": config.RACE_USER_STATUS_DISMISS,
		})
	if err != nil {
		core.Logger.Error("[RaceUsers.ToDismiss]更新失败,raceId:%v,  err:%v", raceId, err)
	} else {
		core.Logger.Info("[RaceUsers.ToDismiss]比赛取消，更新用户的报名状态,raceId:%v, num:%v", raceId, num)
	}
}

// Del 移除用户报名信息
// 同时会删除数据库
func (raceUsers *RaceUsers) Del(userId int, id int64) error {
	// 删除数据库数据
	_, err := core.GetWriter().Delete(&RaceUser{Id: id})
	if err != nil {
		return err
	}

	// 删除内存中的报名星系
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	delete(raceUsers.Users, userId)

	return nil
}

// Add 添加比赛信息
func (raceUsers *RaceUsers) Add(u *RaceUser) {
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	raceUsers.Users[u.UserId] = u
}

// Get 从比赛用户列表中读取用户的比赛信息
func (raceUsers *RaceUsers) Get(userId int) *RaceUser {
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	return raceUsers.Users[userId]
}

// Len 取报名用户人数
func (raceUsers *RaceUsers) Len() int {
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	return len(raceUsers.Users)
}

// GetRaceUsers 获取比赛用户列表
func GetRaceUsers(raceId int64) *RaceUsers {
	list, ok := RaceUsersList.Load(raceId)
	if ok {
		return list.(*RaceUsers)
	}
	return nil
}

// AddRaceUsers 添加一个比赛用户列表
func AddRaceUsers(raceId int64, leagueId int) *RaceUsers {
	list := &RaceUsers{
		RaceId: raceId,
		Mux:    &sync.RWMutex{},
		Users:  make(map[int]*RaceUser),
	}
	RaceUsersList.Store(raceId, list)
	core.Logger.Info("[AddRaceUserList]生成一个新的比赛用户列表,raceId:%v, leagueId:%v", raceId, leagueId)
	return list
}

// GetRaceUserInfo 获取用户的某个比赛的报名信息
func GetRaceUserInfo(raceId int64, userId int) *RaceUser {
	list := GetRaceUsers(raceId)
	if list == nil {
		return nil
	}
	return list.Get(userId)
}

// SetUserRace 获取用户的已报名比赛信息
func SetUserRace(userId int, raceId int64) {
	UserRaceContainer.Store(userId, raceId)
}

// GetUserRace 获取用户的已报名比赛信息
func GetUserRace(userId int) (raceId int64) {
	if v, ok := UserRaceContainer.Load(userId); ok {
		raceId = v.(int64)
	}
	return
}

// DelUserRace 删除用户已报名比赛信息
func DelUserRace(userId int) {
	UserRaceContainer.Delete(userId)
}

// DelUserRaceSpecied 删除用户已报名比赛信息
func DelUserRaceSpecied(userId int, raceId int64) {
	if raceId == int64(0) || raceId == GetUserRace(userId) {
		UserRaceContainer.Delete(userId)
	}
}

// RestoreLeagueRaceUser 恢复比赛用户数据
// 循环进行中的比赛列表 恢复数据
func RestoreLeagueRaceUser() {
	o := core.GetWriter()
	for _, raceInfo := range RaceList.Data {
		var raceUserList []RaceUser
		raceUsers := AddRaceUsers(raceInfo.Id, raceInfo.LeagueId)

		_, err := o.QueryTable(config.TABLE_LEAGUE_RACE_USER).
			Filter("race_id", raceInfo.Id).
			All(&raceUserList)
		if err != nil {
			core.Logger.Error("[RestoreLeagueRaceUser]从league_race_user读取列表失败, raceId:%v, err:%v", raceInfo.Id, err.Error())
			continue
		}
		for i := 0; i < len(raceUserList); i++ {
			// 记录报名人数
			raceInfo.SignupUserCount++
			raceUsers.Users[raceUserList[i].UserId] = &raceUserList[i]
			// 游戏中的用户，需要回复用户当前的比赛状态
			if raceUserList[i].Status == config.RACE_USER_STATUS_SIGNUP {
				SetUserRace(raceUserList[i].UserId, raceUserList[i].RaceId)
				// 如果是机器人，需要记录机器人被占用
				if robot.IsRobot(raceUserList[i].UserId) {
					robot.Occupied(raceUserList[i].UserId)
				}
			}
			core.Logger.Trace("[RestoreLeagueRaceUser]恢复用户报名数据, userId:%v, raceId:%v, leagueId:%v", raceUserList[i].UserId, raceInfo.Id, raceInfo.LeagueId)
		}
		core.Logger.Info("[RestoreLeagueRaceUser]raceId:%v, leagueId:%v, user count:%v", raceInfo.Id, raceInfo.LeagueId, len(raceUsers.Users))
	}
	core.Logger.Info("[RestoreLeagueRaceUser]completed, race count:%v", len(RaceList.Data))
}

// InitRaceUserRank初始化用户排名
func InitRaceUserRank(raceId int64, userIds []int) {
	params := make([]interface{}, 0, len(userIds)+1)
	params = append(params, fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId))
	for _, userId := range userIds {
		// 初始积分1000
		params = append(params, config.RACE_USER_SCORE_BASE, userId)
	}
	core.RedisDo(core.RedisClient0, "ZADD", params...)
	core.RedisDo(core.RedisClient0, "expire", 3*86400)
}

// GetRaceUserRanks 读取比赛用户的排名
func GetRaceUserRanks(raceId int64) []int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId)
	ranks, _ := core.RedisDoInts(core.RedisClient0, "ZREVRANGE", cacheKey, 0, -1, "WITHSCORES")
	return ranks
}

// DelRaceUserRank 从排名集合中删除
func DelRaceUserRank(raceId int64, userIds []int) {
	params := make([]interface{}, 0, len(userIds)+1)
	params = append(params, fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId))
	for _, userId := range userIds {
		params = append(params, userId)
	}
	core.RedisDo(core.RedisClient0, "ZREM", params...)
}

// UpdateRaceUserScoreAndRank 更新用户积分和排名
func UpdateRaceUserScoreAndRank(raceId int64, userId int, scoreChange int) bool {
	// 更新积分
	var params = orm.Params{}
	params["score"] = orm.ColValue(orm.ColAdd, scoreChange)
	core.GetWriter().QueryTable("league_race_user").Filter("race_id", raceId).Filter("user_id", userId).Update(params)
	// 更新排名
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId)
	core.RedisDo(core.RedisClient0, "ZINCRBY", cacheKey, scoreChange, userId)

	core.Logger.Debug("[UpdateRaceUserScoreAndRank]raceId:%v, userId:%v, score:%v", raceId, userId, scoreChange)

	return true
}

// GetRobotFWScore 获取比赛的机器人氛围积分
func (ru *RaceUser) GetRobotFWScore(round int) int {
	if ru.fwScores == nil || len(ru.fwScores) == 0 {
		if len(ru.FwScores) == 0 {
			return 0
		}
		ru.fwScores = make([]int, 0)
		err := json.Unmarshal([]byte(ru.FwScores), &ru.fwScores)
		if err != nil {
			core.Logger.Error("[ru.GetRobotFWScores]FwScores:%#v, err:%v", ru.FwScores, err.Error())
		}
	}
	if round > len(ru.fwScores) {
		return 0
	}
	return ru.fwScores[round-1]
}
