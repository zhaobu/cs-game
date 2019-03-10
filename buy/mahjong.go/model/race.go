package model

import (
	"github.com/astaxie/beego/orm"
	"mahjong.go/library/core"
)

func init() {
	orm.RegisterModel(new(Race))
}

// Race 比赛表
type Race struct {
	Id               int64 `orm:"pk"`
	LeagueId         int   //对应场次id
	SignupUserCount  int   `orm:"-"` // 已报名人数
	RequireUserCount int   `orm:"-"` // 需求人数
	RequireUserMin   int   `orm:"-"` // 最小报名人数
	SignTime         int64 // 报名开始时间
	GiveupTime       int64 `orm:"-"` // 放弃比赛截止时间
	StartTime        int64 // 比赛开始时间
	Status           int   // 比赛状态 0:报名中;1:排赛中;2:游戏中;3:结算中;4:已结束（正常结束）;5:已解散（人数不足解散）
	Round            int   //当前轮次
}

// TableName 数据库表名
func (r *Race) TableName() string {
	return "league_race"
}

// GetRace 获取比赛信息
func GetRace(id int64) *Race {
	r := &Race{Id: id}
	core.GetWriter().Read(r)

	// 读取比赛报名人数
	// cnt, _ := core.GetWriter().QueryTable("league_race_user").Filter("race_id", r.Id).Filter("round", 1).Count()
	cnt, _ := core.GetWriter().QueryTable("league_race_user").Filter("race_id", r.Id).Count()
	r.SignupUserCount = int(cnt)

	return r
}
