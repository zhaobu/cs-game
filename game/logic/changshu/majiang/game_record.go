package majiang

import (
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
	"sort"
)

//GameRecord文件写战绩回放

//一个人的总分
type RankCell struct {
	WinTimes   int32 //胜局
	ChairId    int32 //座位号
	TotalScore int32 //累计总分
}

type GameRecord struct {
	TotalInning uint32
	RankInfo    []*RankCell       //总分排行
	UserScore   []map[int32]int32 //战绩流水
}

func (self *GameRecord) Init(config *pbgame_logic.CreateArg) {
	self.RankInfo = make([]*RankCell, config.PlayerCount)
	for i := int32(0); i < config.PlayerCount; i++ {
		self.RankInfo[i] = &RankCell{ChairId: i}
	}
	self.UserScore = make([]map[int32]int32, 0, config.RInfo.LoopCnt)
}

//记录游戏战绩
func (self *GameRecord) AddGameRecord(info map[int32]int32) {
	self.UserScore = append(self.UserScore, info)
	for k, v := range info {
		self.RankInfo[k].TotalScore += v
		if v > 0 {
			self.RankInfo[k].WinTimes++
		}
	}
	self.TotalInning++
	//重新排名
	sort.Slice(self.RankInfo, func(i, j int) bool {
		if self.RankInfo[i].TotalScore == self.RankInfo[j].TotalScore {
			return self.RankInfo[i].ChairId < self.RankInfo[j].ChairId
		}
		return self.RankInfo[i].TotalScore > self.RankInfo[j].TotalScore
	})
}

//查询游戏记录
func (self *GameRecord) GetGameRecord() *pbgame_logic.S2CGameRecord {
	msg := &pbgame_logic.S2CGameRecord{TotalInning: self.TotalInning}
	rankInfo := make([]*pbgame_logic.GameRecordRank, 0, len(self.RankInfo))
	for _, v := range self.RankInfo {
		rankInfo = append(rankInfo, &pbgame_logic.GameRecordRank{ChairId: v.ChairId, TotalScore: v.TotalScore, WinTimes: v.WinTimes})
	}
	msg.RankInfo = rankInfo
	gameRecord := &pbgame_logic.Json_GameRecord{}
	gameRecord.UserScore = make([]*pbgame_logic.Json_GameRecord_InningInfo, 0, len(self.UserScore))
	for _, v := range self.UserScore {
		tmp := &pbgame_logic.Json_GameRecord_InningInfo{Score: v}
		gameRecord.UserScore = append(gameRecord.UserScore, tmp)
	}
	msg.JsonRecordInfo = util.PB2JSON(gameRecord, false)
	return msg
}
