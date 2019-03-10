package model

import (
	"github.com/astaxie/beego/orm"
	"mahjong.go/library/core"
)

func init() {
	orm.RegisterModel(new(SeasonUser))
}

// SeasonUser 赛季用户信息
type SeasonUser struct {
	Id         int `orm:"pk"`
	SeasonId   int // 赛季id
	UserId     int // 用户id
	GradeId    int // 段位
	GradeLevel int // 段位等级
	StarNum    int // 等级星星数
	Times      int // 总参赛次数
	LastCity   int // 用户所在城市
	Exp        int // 用户经验值
}

// TableName 数据库表名
func (su *SeasonUser) TableName() string {
	return "season_users"
}

// GetSeasonUser 读取用户的赛季信息
func GetSeasonUser(userId, seasonId int) *SeasonUser {
	su := &SeasonUser{}
	if dberr := core.GetWriter().QueryTable("season_users").Filter("season_id", seasonId).Filter("user_id", userId).One(su); dberr != nil {
		// core.Logger.Error("[GetSeasonUser]sql error: %s", dberr.Error())
		return nil
	}
	return su
}
