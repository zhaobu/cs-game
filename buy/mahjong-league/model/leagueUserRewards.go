package model

import (
	"mahjong-league/config"
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(new(LeagueUserReward))
}

// LeagueUserReward 比赛奖励表
type LeagueUserReward struct {
	Id         int64 `orm:"pk"`
	UserId     int
	RaceId     int64
	Rank       int
	RewardType int
	Rewards    int //  非实物填写的具体数值
	Status     int
}

func (r *LeagueUserReward) TableName() string {
	return config.TABLE_LEAGUE_USER_REWARDS
}

// 插入用户奖励
func InsertLeagueUserReward(userId int, raceId int64, rank int, leagueRewards *LeagueReward) *LeagueUserReward {
	rewards := &LeagueUserReward{
		UserId:     userId,
		RaceId:     raceId,
		Rank:       rank,
		RewardType: leagueRewards.RewardType,
		Rewards:    leagueRewards.Rewards,
		Status:     0,
	}
	id, err := core.GetWriter().Insert(rewards)
	if err != nil {
		return nil
	}
	rewards.Id = id
	return rewards
}

// InsertUserRewardsList 批量插入排名数据
func InsertUserRewardsList(rewardsList []LeagueUserReward) (int64, error) {
	return core.GetWriter().InsertMulti(len(rewardsList), rewardsList)
}
