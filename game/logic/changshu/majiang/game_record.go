package majiang

import (
	pbgame_logic "cy/game/pb/game/mj/changshu"
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
	RankInfo  []*RankCell       //总分排行
	UserScore []map[int32]int32 //战绩流水
}

func (self *GameRecord) Init(config *pbgame_logic.CreateArg) {
	self.RankInfo = make([]*RankCell, config.PlayerCount)
	for i := int32(0); i < config.PlayerCount; i++ {
		self.RankInfo[i] = &RankCell{ChairId: i}
	}
	self.UserScore = make([]map[int32]int32, 0, config.RInfo.LoopCnt)
}

func (self *GameRecord) AddGameEnd(info map[int32]int32) {
	self.UserScore = append(self.UserScore, info)
	for k, v := range info {
		self.RankInfo[k].TotalScore += v
		if v > 0 {
			self.RankInfo[k].WinTimes++
		}
	}
	//重新排名
	sort.Slice(self.RankInfo, func(i, j int) bool {
		if self.RankInfo[i].TotalScore == self.RankInfo[j].TotalScore {
			return self.RankInfo[i].ChairId < self.RankInfo[j].ChairId
		}
		return self.RankInfo[i].TotalScore > self.RankInfo[j].TotalScore
	})
}
