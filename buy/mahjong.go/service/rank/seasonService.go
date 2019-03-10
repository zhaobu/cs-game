package rank

import (
	"fmt"

	"mahjong.go/config"
	"mahjong.go/library/core"
)

// GetConsumeCard 获取参赛卡配置
func GetConsumeCard(seasonId, gradeId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_SEASON_CARD_CONSUME, seasonId)
	consume, err := core.RedisDoInt(core.RedisClient0, "hget", cacheKey, gradeId)
	if err != nil {
		core.Logger.Error("[GetConsumeCard]seasonId:%v, gradeId:%v, err:%v", seasonId, gradeId, err.Error())
	}
	return consume
}
