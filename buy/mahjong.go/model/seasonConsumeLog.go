package model

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(new(SeasonConsumeLog))
}

// SeasonConsumeLog 参赛卡消耗日志
type SeasonConsumeLog struct {
	Id         int `orm:"pk"`
	SeasonId   int
	UserId     int
	RoomId     int64
	Num        int // 游戏类型
	Grade      int
	Level      int
	Stars      int
	CreateTime int64
}

// TableName 数据库表名
func (s *SeasonConsumeLog) TableName() string {
	return "season_consume_log"
}
