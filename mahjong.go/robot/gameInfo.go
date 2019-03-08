package robot

import (
	"encoding/json"
)

// 启动机器人配置
type GameInfo struct {
	Remote     string `json:"remote"`      // 远程服务器地址
	RoomId     int64  `json:"room_id"`     // 房间id
	RobotId    int    `json:"robot_id"`    // 机器人id
	AILevel    int    `json:"ai_level"`    // 启动机器人级别,如果=0,默认是最高级的机器人
	LeagueId   int    `json:"league_id"`   // 联赛id
	RaceId     int64  `json:"race_id"`     // 比赛id
	GType      int    `json:"gType"`       // 游戏类型
	CType      int    `json:"cType"`       // 创建类型
	RequireCnt int    `json:"requireCnt"`  // 需要机器人数量
	CoinType   int    `json:"coin_type"`   // 金币场id
	GradeId    int    `json:"grade_id"`    // 段位id
	TRound     int    `json:"total_round"` // 总局数
}

func NewGameInfo(remote string, roomId int64, gType, cType, requireCnt, tRound int) *GameInfo {
	gameInfo := &GameInfo{
		Remote:     remote,
		RoomId:     roomId,
		GType:      gType,
		CType:      cType,
		RequireCnt: requireCnt,
		TRound:     tRound,
	}
	return gameInfo
}

func (i *GameInfo) String() string {
	v, _ := json.Marshal(i)
	return string(v)
}

func DeserializeGameInfo(s string) *GameInfo {
	var gameInfo GameInfo
	json.Unmarshal([]byte(s), &gameInfo)
	return &gameInfo
}
