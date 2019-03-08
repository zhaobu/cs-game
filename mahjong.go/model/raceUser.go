package model

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

func init() {
	orm.RegisterModel(new(RaceUser))
}

// RaceUser 联赛用户表
type RaceUser struct {
	Id         int64 `orm:"pk"`
	RaceId     int64 // 比赛id
	UserId     int   // 用户id
	Round      int   // 轮次
	Status     int   // 用户状态, 0: 游戏中（已报名）;1: 已退赛; 2: 已淘汰; 5: 已解散
	Score      int   // 用户分数
	GiveupTime int64 // 退赛时间
	SignTime   int64 //  报名时间
	FailTime   int64 // 淘汰时间
	Rank       int   `orm:"-"` // 用户排名
}

// TableName RaceUser的数据库表名
func (ru *RaceUser) TableName() string {
	return "league_race_user"
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

// GetRaceUserRank 获取用户排名
func GetRaceUserRank(raceId int64, userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId)
	// 读排名
	rank, _ := core.RedisDoInt(core.RedisClient0, "ZREVRANK", cacheKey, userId)
	return rank + 1
}

// GetRaceUserScore 获取用户排名
func GetRaceUserScore(raceId int64, userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_LEAGUE_RACE_SCORES, raceId)
	// 读排名
	score, _ := core.RedisDoInt(core.RedisClient0, "ZSCORE", cacheKey, userId)
	return score
}
