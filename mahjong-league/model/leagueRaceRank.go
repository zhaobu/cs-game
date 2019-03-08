package model

import (
	"mahjong-league/config"
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(new(RaceRank))
}

// RaceRank 比赛排名表
type RaceRank struct {
	Id     int64 `orm:"pk"`
	RaceId int64 // 比赛id
	UserId int   // 用户id
	Rank   int   // 用户排名
	Round  int   // 最终轮次
	Score  int   // 用户分数
	Status int   // 用户状态, 0: 正常结束; 1: 退赛; 2: 已淘汰
}

// TableName RaceRank的真实表名
func (rr *RaceRank) TableName() string {
	return config.TABLE_LEAGUE_RACE_RANK
}

// InsertRaceRanks 批量插入排名数据
func InsertRaceRanks(rankList []RaceRank) (int64, error) {
	return core.GetWriter().InsertMulti(len(rankList), rankList)
}
