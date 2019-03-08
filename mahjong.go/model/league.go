package model

import (
	"encoding/json"

	"github.com/astaxie/beego/orm"
	"mahjong.go/library/core"
)

func init() {
	orm.RegisterModel(new(League))
}

// League 联赛模板
type League struct {
	Id               int `orm:"pk"`
	Name             string
	Img              string
	Icon             int
	GameType         int     // 游戏类型
	Setting          string  // 游戏设置
	Rounds           string  // 比赛流程,[[晋级类型，晋级人数，设定局数],...]
	RequireUserCount int     // 需要报名人数
	RequireUserMin   int     // 最小报名人数
	PriceEntityId    int     // 报名费用entityId
	Price            int     // 报名费用
	LeagueType       int     // 比赛类型
	StartCondition   int     // 开始条件，1：非定时塞；2：定时赛
	Cycle            int     // 循环模式，0=非循环赛 1=每天  2=每周  3=每月
	CycleVal         int     // 循环值
	Status           int     // 是否有效，0=无效 1=有效
	OpenTime         int64   // 开放时间戳
	CloseTime        int64   // 关闭时间戳
	UpdateTime       int64   // 最后修改时间
	SignTime         int64   // 报名开始时间，距当天0点秒数
	StartTime        int64   // 比赛开始时间，距当天0点秒数
	Category         int     // 比赛场类别
	setting          []int   `orm:"-"` // setting的解析版
	rounds           [][]int `orm:"-"` // rounds的展开版
}

// TableName 数据库表名
func (l *League) TableName() string {
	return "league_list"
}

// GetLeague 读取联赛信息
func GetLeague(id int) *League {
	l := &League{Id: id}
	err := core.GetWriter().Read(l)
	if err != nil {
		core.Logger.Error("[GetLeague]id:%v, err:%v", id, err.Error())
	}
	return l
}

// GetSetting 读取联赛设置
func (l *League) GetSetting() []int {
	if l.setting == nil || len(l.setting) == 0 {
		l.setting = make([]int, 0)
		json.Unmarshal([]byte(l.Setting), &l.setting)
	}
	return l.setting
}

// GetRounds 获取游戏晋级设置
func (l *League) GetRounds() [][]int {
	if l.rounds == nil || len(l.rounds) == 0 {
		l.rounds = make([][]int, 0)
		err := json.Unmarshal([]byte(l.Rounds), &l.rounds)
		if err != nil {
			core.Logger.Error("[l.GetRounds],l.rounds:%v, league:%#v, err:%v", err.Error(), l.rounds, l)
		}
	}
	return l.rounds
}

// GetGameRound 根据当前比赛轮次，获取房间局数
func (l *League) GetGameRound(round int) int {
	return l.GetRounds()[round-1][2]
}
