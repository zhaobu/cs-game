package model

import (
	"github.com/astaxie/beego/orm"
	"mahjong.go/library/core"
)

func init() {
	orm.RegisterModel(new(LeagueReward))
}

// LeagueReward 比赛奖励表
type LeagueReward struct {
	Id            int64  `orm:"pk"`
	LeagueId      int    //  比赛场id
	CalcStartRank int    //  计算排名的开始
	CalcEndRank   int    //  计算排名的结束
	RewardType    int    // 奖品类型
	Rewards       int    //  非实物填写的具体数值
	Rank          string //  名次内容
	Content       string // 奖品内容
}

// TableName 数据库真实表名
func (lr *LeagueReward) TableName() string {
	return "league_rewards"
}

// 从数据库读取某个联赛的奖励列表
func loadLeagueRewardsForceDB(leagueId int) []LeagueReward {
	var result []LeagueReward
	_, err := core.GetWriter().
		QueryTable("league_rewards").
		Filter("league_id", leagueId).
		OrderBy("id").
		All(&result)
	if err != nil {
		core.Logger.Error("[loadLeagueRewardsForceDB]从数据库读取league_rewards失败, leagueId:%v, err:%v", leagueId, err.Error())
	}
	return result
}

// 展开联赛奖励列表，按排名
func expandLeagueRewards(rewards []LeagueReward) []*LeagueReward {
	rewardsList := make([]*LeagueReward, 0)
	for i := 0; i < len(rewards); i++ {
		rewardsInfo := rewards[i]
		for i := rewardsInfo.CalcStartRank; i <= rewardsInfo.CalcEndRank; i++ {
			rewardsList = append(rewardsList, &rewardsInfo)
		}
	}
	return rewardsList
}

// GetFrontLeagueRewards 获取前N条比赛奖励
func GetFrontLeagueRewards(leagueId int) []string {
	rewards := make([]string, 0)

	loopTimes := 0
	for _, rewardInfo := range expandLeagueRewards(loadLeagueRewardsForceDB(leagueId)) {
		rewards = append(rewards, rewardInfo.Rank, rewardInfo.Content)
		loopTimes++
		if loopTimes >= 3 {
			break
		}
	}

	return rewards
}
