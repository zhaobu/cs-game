package game

import (
	"fmt"

	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 获取保存回放版本的cacheKey
func (p *Playback) getVersionCacheKey() string {
	return fmt.Sprintf(config.CACHE_KEY_PLAYBACK_VERSION, p.roomId, p.round)
}

// 获取保存回放数据的cacheKey
func (p *Playback) getDataCacheKey(isIntact bool) string {
	if isIntact {
		return fmt.Sprintf(config.CACHE_KEY_PLAYBACK_DATA_INTACT, p.roomId, p.round)
	}
	return fmt.Sprintf(config.CACHE_KEY_PLAYBACK_DATA, p.roomId, p.round)
}

// 获取保存回放数据的cacheKey的有效期
func (p *Playback) getDataCacheExpire(isIntact bool) int {
	if isIntact {
		return config.GAME_PLAYBACK_DATA_INTACT_EXPIRE
	}
	return config.GAME_PLAYBACK_DATA_EXPIRE
}

// 保存回放的版本去redis
func (p *Playback) saveVersionToRedis() {
	redisConn := core.RedisClient5.Get()
	defer redisConn.Close()

	// 设置回放对应的版本号
	redisConn.Do("set", p.getVersionCacheKey(), GameVersion)
	redisConn.Do("expire", p.getVersionCacheKey(), config.GAME_PLAYBACK_DATA_EXPIRE)
}

// 保存回放数据
// 回放数据会有一些新旧版本兼容的问题，所以在保存回放数据的时候，顺便保留一份版本数据，用于判断版本是否匹配
func (p *Playback) saveToRedis(data []byte, isIntact bool) {
	redisConn := core.RedisClient5.Get()
	defer redisConn.Close()

	// 设置回放对应的版本号
	redisConn.Do("set", p.getVersionCacheKey(), GameVersion)
	redisConn.Do("expire", p.getVersionCacheKey(), config.GAME_PLAYBACK_DATA_EXPIRE)

	// 设置回放数据
	cacheKey := p.getDataCacheKey(isIntact)
	redisConn.Do("set", cacheKey, data)
	redisConn.Do("expire", cacheKey, p.getDataCacheExpire(isIntact))

	core.Logger.Info("[savePlayback][redis]roomId:%v,round:%v,isIntact", p.roomId, p.round, isIntact)
}
