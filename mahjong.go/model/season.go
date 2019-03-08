package model

import (
	"encoding/json"
	"strconv"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// func init() {
// 	orm.RegisterModel(new(Season))
// }

// Season 参赛卡消耗日志
type Season struct {
	Id        int   `json:"id"`
	GType     int   `json:"game_type"`
	Setting   []int `json:"setting"`
	Rounds    int   `json:"rounds"`
	IsFree    int   `json:"is_free"`
	Status    int   `json:"status"`
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

// IsOpen 是否上架
func (s *Season) IsOpen() bool {
	t := util.GetTime()
	return s.Status == 1 && t >= s.StartTime && t < s.EndTime
}

// // TableName 数据库表名
// func (s *Season) TableName() string {
// 	return "season_list"
// }

// GetSeason 读取当前赛季
func GetSeason() *Season {
	jsonInfo, err := core.RedisDoStringMap(core.RedisClient0, "hgetall", config.CACHE_KEY_SEASON_INFO)
	if err != nil {
		core.Logger.Error("从redis当前赛季数据失败,err:%v", err.Error())
		return nil
	}
	id, _ := strconv.Atoi(jsonInfo["id"])
	gType, _ := strconv.Atoi(jsonInfo["game_type"])
	rounds, _ := strconv.Atoi(jsonInfo["rounds"])
	isFree, _ := strconv.Atoi(jsonInfo["is_free"])
	status, _ := strconv.Atoi(jsonInfo["status"])
	startTime, _ := strconv.ParseInt(jsonInfo["start_time"], 10, 64)
	endTime, _ := strconv.ParseInt(jsonInfo["end_time"], 10, 64)

	season := &Season{
		Id:        id,
		GType:     gType,
		Rounds:    rounds,
		IsFree:    isFree,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
	}
	json.Unmarshal([]byte(jsonInfo["setting"]), &season.Setting)

	core.Logger.Debug("[GetSeason]season:%#v", season)

	return season
}
