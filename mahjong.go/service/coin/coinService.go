package coin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	configService "mahjong.go/service/config"
)

// Config 金币场配置
type Config struct {
	BaseCoin        int `json:"base_coin"`         // 底分
	RequireLowCoin  int `json:"require_low_coin"`  // 进入最低分
	RequireHighCoin int `json:"require_high_coin"` // 进入最高分
	ConsumeCoin     int `json:"consume_coin"`      // 入场金币消耗
	Status          int `json:"status"`            // 是否有效 0=有效；1：无效
}

func getConfigKey(coinType, gType int) string {
	// return fmt.Sprintf(config.CACHE_KEY_ROOM_COIN_CONFIG, gType, coinType)
	// 暂时不支持根据玩法选配置
	return fmt.Sprintf(config.CACHE_KEY_ROOM_COIN_CONFIG, 0, coinType)
}

// GetConfig 获取金币场配置
func GetConfig(coinType, gType int) *Config {
	var cfg Config
	jsonBytes, err := core.RedisDoBytes(core.RedisClient0, "get", getConfigKey(coinType, gType))
	if err != nil {
		core.Logger.Error("从redis读取自建金币场配置失败,coinType:%v,gType:%v,err:%v", coinType, gType, err.Error())
		return nil
	}
	err = json.Unmarshal(jsonBytes, &cfg)
	if err != nil {
		core.Logger.Error("解析自建金币场配置失败,coinType:%v,gType:%v,json data:%v, err:%v", coinType, gType, string(jsonBytes), err.Error())
		return nil
	}
	return &cfg
}

func getGameLogKey(userId, otherUserId int, createTime int64) string {
	Ymd := time.Unix(createTime, 0).Format("20060102")
	if userId < otherUserId {
		return fmt.Sprintf(config.CACHE_KEY_GAME_LOG, Ymd, userId, otherUserId)
	}
	return fmt.Sprintf(config.CACHE_KEY_GAME_LOG, Ymd, otherUserId, userId)
}

// HasTogetherGameLog 判断两个用户id是否可以同时游戏
func HasTogetherGameLog(userId, otherUserId int) bool {
	return false
	if configService.IsRobot(userId) || configService.IsRobot(otherUserId) {
		return false
	}
	exists, err := core.RedisDoBool(core.RedisClient0, "exists", getGameLogKey(userId, otherUserId, util.GetTime()))
	if err != nil {
		core.Logger.Warn("[HasTogetherGameLog]从redis获取数据失败,err:%v", err.Error())
	}
	return exists
}

// SetTogetherGameLog 设置两个用户的同时游戏记录
func SetTogetherGameLog(userId, otherUserId int, createTime int64) {
	if configService.IsRobot(userId) || configService.IsRobot(otherUserId) {
		return
	}
	cacheKey := getGameLogKey(userId, otherUserId, createTime)
	_, err := core.RedisDo(core.RedisClient0, "set", cacheKey, 1)
	if err == nil {
		// 设置过期时间
		core.RedisDo(core.RedisClient0, "expire", cacheKey, 86400)
	} else {
		core.Logger.Warn("[SetTogetherGameLog]写数据进redis失败,err:%v", err.Error())
	}
}

// IncrCoinUserCnt 增加金币场在线人数
func IncrCoinUserCnt(gType, coinType int) {
	core.RedisDo(core.RedisClient3, "hincrby", config.CACHE_KEY_COIN_USER_CNT, fmt.Sprintf("%v_%v", gType, coinType), 1)
}

// DecrCoinUserCnt 增加金币场在线人数
// 如果在线人数小于0，则数据出错，修正为在线人数为0
func DecrCoinUserCnt(gType, coinType int, cnt int) {
	currentValue, err := core.RedisDoInt(core.RedisClient3, "hincrby", config.CACHE_KEY_COIN_USER_CNT, fmt.Sprintf("%v_%v", gType, coinType), -1*cnt)
	if err == nil {
		if currentValue < 0 {
			core.Logger.Debug("[DecrCoinUserCnt]金币场在线人数异常, gtype:%v, coinType:%v,userCnt:%v", gType, coinType, currentValue)
			core.RedisDo(core.RedisClient3, "hset", config.CACHE_KEY_COIN_USER_CNT, fmt.Sprintf("%v_%v", gType, coinType), 0)
		}
	} else {
		core.Logger.Warn("[DecrCoinUserCnt]写数据进redis失败,err:%v", err.Error())
	}
}

// UpdateProvinceRank 更新用户的省排名
func UpdateProvinceRank(userId int, score int) {
	_, err := core.RedisDo(core.RedisClient0, "zadd", config.CACHE_KEY_COIN_RANK_PROVINCE, score, userId)
	if err != nil {
		core.Logger.Error("[coin.UpdateProvinceRank]error:%v", err.Error())
	}
}

// UpdateCityRank 更新用户的城市排名
func UpdateCityRank(userId, city int, score int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_COIN_RANK_CITY, city)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[coin.UpdateCityRank]error:%v", err.Error())
	}
}

// UpdateFriendRank 更新用户的好友排名
func UpdateFriendRank(userId int, score int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, userId)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[coin.UpdateFriendRank]error:%v", err.Error())
	}
}

// UpdateUserFriendRank 更新用户在好友排行中的score
func UpdateUserFriendRank(userId, friendUserId int, score int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, friendUserId)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[coin.UpdateFriendRank]error:%v", err.Error())
	}
}
