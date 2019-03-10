// 获取麻将排序数据
package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"

	fbsCommon "mahjong.go/fbs/Common"
)

// 读取创建牌局价格
func getMahjongRoundPrice(round int) int {
	var price int
	var exists bool

	if isDragonBoatFestival() {
		price, exists = config.MahjongRoundPriceDragonBoatFestival[round]
	} else {
		price, exists = config.MahjongRoundPrice[round]
	}

	if exists {
		return price
	}
	return 0
}

// 读取比赛牌局价格
func getMatchPrice(gType int) int {
	if price, exists := config.MahjongMatchPrice[gType]; exists {
		return price
	}
	return 0
}

// 读取创建牌局价格
func getLDPrice(round int) int {
	if price, exists := config.MahjongLDRoundPrice[round]; exists {
		return price
	}
	return 0
}

// getCreatePrice 读取自建类型的房间价格（创建、电视端、俱乐部）
func getCreatePrice(gType, round int) (int, error) {
	var price int
	priceString, err := core.RedisDoString(core.RedisClient4, "hget", config.CACHE_KEY_ROOM_PRICE, gType)
	if err != nil {
		core.Logger.Error("从redis读取自建类型的价格失败,gType:%v,err:%v", gType, err.Error())
		return price, err
	}
	// 解析出
	var priceList map[string]interface{}
	err = json.Unmarshal([]byte(priceString), &priceList)
	if err != nil {
		core.Logger.Error("解析自建类型的价格配置失败,gType:%v, json data:%v,err:%v", gType, priceString, err.Error())
		return price, err
	}
	// 解析具体的值
	if v, exists := priceList[strconv.Itoa(round)]; exists {
		switch v.(type) {
		case int:
			price = v.(int)
		case int64:
			price = int(v.(int64))
		case float64:
			price = int(v.(float64))
		case string:
			price, _ = strconv.Atoi(v.(string))
		default:
			core.Logger.Error("自建类型局数对应的价格无法解析,gType:%v, round:%v, v:%#v", gType, round, v)
			err = fmt.Errorf("自建类型局数对应的价格无法解析,gType:%v, round:%v, v:%#v", gType, round, v)
			break
		}
	} else {
		core.Logger.Error("自建类型局数对应的价格未设置, gType:%v, round:%v", gType, round)
		err = fmt.Errorf("自建类型局数对应的价格未设置, gType:%v, round:%v", gType, round)
	}
	return price, err
}

// GetGamePrice 读取牌局创建价格
func GetGamePrice(gType, cType, round int) int {
	price := 0
	var err error
	if IsRandomRoom(cType) {
		// 随机组局
		price = config.MahjongRandomPrice
	} else if IsMatchRoom(cType) {
		price = getMatchPrice(gType)
	} else {
		price, err = getCreatePrice(gType, round)
		// 如果价格读取失败，则做一次容错
		if err != nil {
			// 自主创建
			if gType == fbsCommon.GameTypeMAHJONG_STT {
				// 两丁拐价格独立计算
				price = getLDPrice(round)
			} else {
				price = getMahjongRoundPrice(round)
			}
		}
	}
	return price
}

// IsRobot 判断用户是否机器人
// 范围: [2000, 10000), [50000, 100000)
func IsRobot(userId int) bool {
	if userId >= 2000 && userId <= 10000 {
		return true
	}
	if userId >= 50000 && userId < 100000 {
		return true
	}
	return false
}

// CheckRandomRoomGameType 检查随机房间是否支持此类型
func CheckRandomRoomGameType(gameType int) bool {
	if util.IntInSlice(gameType, config.RandomGameTypeList) {
		return true
	}
	return false
}

// GetRandomRoomDefaultSetting 读取游戏类型的随机组局默认配置
func GetRandomRoomDefaultSetting(gameType int) []int {
	return config.RandomRoomDefaultSettingList[gameType]
}

// CheckMatchRoomGameType 检查随机房间是否支持此类型
func CheckMatchRoomGameType(gameType int) bool {
	if util.IntInSlice(gameType, config.MatchGameTypeList) {
		return true
	}

	return false
}

// GetMatchRoomDefaultSetting 读取游戏类型的比赛配置
// 如果取不到，认为是数据错误，读默认值
func GetMatchRoomDefaultSetting(gameType int) []int {
	if setting, exists := config.MatchRoomDefaultSettingList[gameType]; exists {
		return setting
	}
	return nil
}

// GetCoinRoomDefaultSetting 获取金币场的默认玩法
// 如果取不到，认为是数据错误，不能开房间
func GetCoinRoomDefaultSetting(gameType int) []int {
	if setting, exists := config.CoinRoomDefaultSettingList[gameType]; exists {
		return setting
	}
	return nil
}

// 读取积分倍数
func GetScoreMultiple(gameType int) int {
	if multiple, exists := config.MatchScoreMultipleList[gameType]; exists {
		return multiple
	}
	return 1
}

// IsNoticeId
func IsNoticeId(chatId int16) bool {
	switch chatId {
	// 实时语音
	case config.CHAT_ID_YY_JOIN:
		fallthrough
	case config.CHAT_ID_YY_QUIT:
		fallthrough
	// 信号强度
	case config.CHAT_ID_SIGNAL_VERY_STRONGER:
		fallthrough
	case config.CHAT_ID_SIGNAL_STRONGER:
		fallthrough
	case config.CHAT_ID_SIGNAL_NORMAL:
		fallthrough
	case config.CHAT_ID_SIGNAL_WEAK:
		fallthrough
	case config.CHAT_ID_SIGNAL_VERY_WEAK:
		fallthrough
	case config.CHAT_ID_VOICE_ID:
		return true
	default:
		break
	}
	return false
}

// isDragonBoatFestival
// 2016-6-14~2016-7-14 09:00:00，钻石价格减半
func isDragonBoatFestival() bool {
	currentTime := util.GetTime()
	return currentTime >= int64(1497369600) && currentTime < int64(1499994000)
}

// IsCreateRoom 是否是自建类型的房间
func IsCreateRoom(cType int) bool {
	return cType == config.ROOM_TYPE_CREATE
}

// IsTVRoom 是否是电视端类型的房间
func IsTVRoom(cType int) bool {
	return cType == config.ROOM_TYPE_TV
}

// IsRandomRoom 是否是随机类型的房间
func IsRandomRoom(cType int) bool {
	return cType == config.ROOM_TYPE_RAND
}

// IsMatchRoom 是否是雀王榜类型的房间
func IsMatchRoom(cType int) bool {
	return cType == config.ROOM_TYPE_MATCH
}

// IsClubRoom 是否是俱乐部类型的房间
func IsClubRoom(cType int) bool {
	return cType == config.ROOM_TYPE_CLUB
}

// IsClubMatchRoom 是否是俱乐部淘汰赛类型的房间
func IsClubMatchRoom(cType int) bool {
	return cType == config.ROOM_TYPE_CLUB_MATCH
}

// IsCoinRoom 是否是金币场房间
func IsCoinRoom(cType int) bool {
	return cType == config.ROOM_TYPE_COIN
}

// IsLeagueRoom 是否联赛房间
func IsLeagueRoom(cType int) bool {
	return cType == config.ROOM_TYPE_LEAGUE
}

// IsRankRoom 是否是排位赛类型的房间
func IsRankRoom(cType int) bool {
	return cType == config.ROOM_TYPE_RANK
}
