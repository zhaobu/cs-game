package model

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"sync"

	"github.com/astaxie/beego/orm"
)

// LeagueRewardsList 所有联赛的奖励列表
var LeagueRewardsList *LeagueRewardsMap

func init() {
	orm.RegisterModel(new(LeagueReward))
	LeagueRewardsList = &LeagueRewardsMap{
		Mux:  &sync.RWMutex{},
		Data: make(map[int][]*LeagueReward),
	}
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
	return config.TABLE_LEAGUE_REWARDS
}

// LeagueRewardsMap 所有比赛奖励
type LeagueRewardsMap struct {
	Mux  *sync.RWMutex
	Data map[int][]*LeagueReward
}

// Restore 重新从数据库恢复所有联赛奖励
// 脚本启动时调用
func (rl *LeagueRewardsMap) Restore() {
	for _, leagueInfo := range LeagueList.Data {
		rl.ReloadLeagueRewards(leagueInfo.Id)
	}
	core.Logger.Info("[RestoreLeagueRewards]completed, 恢复奖励的比赛个数:%v", len(rl.Data))
}

// ReloadLeagueRewards 从数据库恢复某个联赛的奖励
func (rl *LeagueRewardsMap) ReloadLeagueRewards(leagueId int) {
	rewards := loadLeagueRewardsForceDB(leagueId)
	expandRewards := expandLeagueRewards(rewards)

	rl.Mux.Lock()
	defer rl.Mux.Unlock()
	rl.Data[leagueId] = expandRewards

	core.Logger.Info("[ReloadLeagueRewards]恢复比赛奖励, leagueId:%v, 奖励条数:%v", leagueId, len(expandRewards))
}

// Get 读取某个联赛的奖励列表
func (rl *LeagueRewardsMap) Get(leagueId int) []*LeagueReward {
	rl.Mux.Lock()
	defer rl.Mux.Unlock()
	return rl.Data[leagueId]
}

// 从数据库读取某个联赛的奖励列表
func loadLeagueRewardsForceDB(leagueId int) []LeagueReward {
	var result []LeagueReward
	_, err := core.GetWriter().
		QueryTable(config.TABLE_LEAGUE_REWARDS).
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
	lastRank := 0
	for _, rewardInfo := range LeagueRewardsList.Get(leagueId) {
		if rewardInfo.CalcStartRank == lastRank {
			continue
		}
		rewards = append(rewards, rewardInfo.Rank, rewardInfo.Content)
		lastRank = rewardInfo.CalcStartRank
		loopTimes++
		if loopTimes >= 3 {
			break
		}

	}

	return rewards
}
