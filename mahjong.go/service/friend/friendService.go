package friend

import (
	"fmt"

	"mahjong.go/config"
	"mahjong.go/library/core"
)

// GetFriends 获取用户的好友列表
func GetFriends(userId int) []int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_FRIEND, userId)
	friends, err := core.RedisDoInts(core.RedisClient0, "zrange", cacheKey, 0, -1)
	if err != nil {
		core.Logger.Error("[GetFriends]err:%v", err.Error())
	}
	return friends
}
