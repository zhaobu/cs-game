package rank

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/gateway"
	"mahjong.go/library/core"
	"mahjong.go/model"
	"mahjong.go/rank"
	configService "mahjong.go/service/config"
	userService "mahjong.go/service/user"
)

// UpdateProvinceRank 更新用户的省排名
func UpdateProvinceRank(seasonId, userId int, score int64) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_PROVINCE, seasonId)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[UpdateProvinceRank]error:%v", err.Error())
	}
}

// GetProvinceRank 获取用户排名的全省排名
func GetProvinceRank(seasonId, userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_PROVINCE, seasonId)
	rank, err := core.RedisDoInt(core.RedisClient0, "ZREVRANK", cacheKey, userId)
	core.Logger.Debug("[GetProvinceRank]seasonId:%v, userId:%v, rank:%v", seasonId, userId, rank)
	if err != nil {
		// core.Logger.Error("[GetProvinceRank]error:%v", err.Error())
	}
	return rank
}

// UpdateCityRank 更新用户的城市排名
func UpdateCityRank(seasonId, userId, city int, score int64) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_CITY, seasonId, city)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[UpdateCityRank]error:%v", err.Error())
	}
}

// UpdateFriendRank 更新用户的好友排名
func UpdateFriendRank(seasonId, userId int, score int64) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, seasonId, userId)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[UpdateFriendRank]error:%v", err.Error())
	}
}

// UpdateUserFriendRank 更新用户在好友排行中的score
func UpdateUserFriendRank(seasonId, userId, friendUserId int, score int64) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, seasonId, friendUserId)
	_, err := core.RedisDo(core.RedisClient0, "zadd", cacheKey, score, userId)
	if err != nil {
		core.Logger.Error("[UpdateFriendRank]error:%v", err.Error())
	}
}

// GetUserFriendRank 获取用户在好友中的排行
func GetUserFriendRank(seasonId, userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, seasonId, userId)
	rank, err := core.RedisDoInt(core.RedisClient0, "ZREVRANK", cacheKey, userId)
	if err != nil {
		core.Logger.Error("[GetUserFriendRank]error:%v", err.Error())
	}
	return rank
}

// GetFriendRankUser 获取某个排行的用户
func GetFriendRankUser(seasonId, userId, rank int) (frindId int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_FRIEND, seasonId, userId)
	userIds, err := core.RedisDoInts(core.RedisClient0, "ZREVRANGE", cacheKey, rank, rank)
	if err == nil {
		if len(userIds) > 0 {
			frindId = userIds[0]
		}
	} else {
		core.Logger.Error("[GetUserFriendRank]error:%v", err.Error())
	}
	return
}

// UpdateGradeSignCnt 更新段位报名人数
func UpdateGradeSignCnt(gradeId int, inc int) {
	core.RedisDo(core.RedisClient0, "hincrby", config.CACHE_KEY_SEASON_SIGN, gradeId, inc)
}

// 从缓存中取seasonUser
func getSeasonUserFromCache(userId, seasonId int) *model.SeasonUser {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_SEASON_USER, userId)
	jsonInfo, err := core.RedisDoStringMap(core.RedisClient0, "hgetall", cacheKey)
	if err != nil || len(jsonInfo) == 0 {
		return nil
	}

	id, _ := strconv.Atoi(jsonInfo["id"])
	gradeId, _ := strconv.Atoi(jsonInfo["grade_id"])
	gradeLevel, _ := strconv.Atoi(jsonInfo["grade_level"])
	star, _ := strconv.Atoi(jsonInfo["star_num"])
	exp, _ := strconv.Atoi(jsonInfo["exp"])
	times, _ := strconv.Atoi(jsonInfo["times"])
	lastCity, _ := strconv.Atoi(jsonInfo["last_city"])

	su := &model.SeasonUser{
		Id:         id,
		SeasonId:   seasonId,
		UserId:     userId,
		GradeId:    gradeId,
		GradeLevel: gradeLevel,
		StarNum:    star,
		Exp:        exp,
		Times:      times,
		LastCity:   lastCity,
	}
	return su
}

// GetSeasonUser 读取用户的赛季信息
func GetSeasonUser(userId, seasonId int) *model.SeasonUser {
	// 先从cache中取
	var su *model.SeasonUser

	su = getSeasonUserFromCache(userId, seasonId)

	// cache数据不存在，从db中取
	if su == nil {
		su = model.GetSeasonUser(userId, seasonId)
	}

	return su
}

// UpdateSeasonUserTimes 比赛结束后，更新用户的排位数据
func UpdateSeasonUserTimes(seasonId, userId, fromSlevel, finalSLevel, exp int) {
	// 更新db中的数据
	gradeId, gradeLevel, star := rank.ExplainSLevel(finalSLevel)
	fromGradeId, _, _ := rank.ExplainSLevel(fromSlevel)
	var params = orm.Params{
		"grade_id":    gradeId,
		"grade_level": gradeLevel,
		"star_num":    star,
		"times":       orm.ColValue(orm.ColAdd, 1),
		"exp":         exp,
	}
	if num, dberr := core.GetWriter().QueryTable("season_users").Filter("season_id", seasonId).Filter("user_id", userId).Update(params); num == 0 || dberr != nil {
		core.Logger.Error("[UpdateSeasonUserTimes]failed, seasonId:%v, userId:%v, sLevel:%v, effect rows:%v, err:%v", seasonId, userId, finalSLevel, num, dberr)
	}

	// cache中存在，需要刷新cache中的数据
	su := getSeasonUserFromCache(userId, seasonId)
	if su != nil {
		SetSeasonUserFromCache(userId, gradeId, gradeLevel, star, su.Times+1, exp)
	}
	core.Logger.Info("[UpdateSeasonUserTimes]seasonId:%v, userId:%v, sLevel:%v", seasonId, userId, finalSLevel)

	// 如果用户是刚升级到雀王，则清除之前的记录
	// 否则增加游戏次数
	if gradeId >= 5 {
		if fromGradeId == 4 {
			// 清除之前的记录
			DelDayPlayTimes(userId)
		} else {
			// 更新用户的每日游戏次数
			IncrDayPlayTimes(userId)
		}
	}
}

// SetSeasonUserFromCache 保存用户的排位赛数据
func SetSeasonUserFromCache(userId, gradeId, gradeLevel, star int, times, exp int) {
	data := []interface{}{fmt.Sprintf(config.CACHE_KEY_SEASON_USER, userId)}
	data = append(data, "grade_id", gradeId)
	data = append(data, "grade_level", gradeLevel)
	data = append(data, "star_num", star)
	data = append(data, "times", times)
	data = append(data, "exp", exp)
	core.RedisDo(core.RedisClient0, "hmset", data...)
	core.Logger.Info("[SetSeasonUserFromCache]userId:%v,data:%+v", userId, data)
}

// SaveRankUpgrade 存储用户的排名升级
func SaveRankUpgrade(userId, originUserId, rank int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_UPGRADE, userId)
	var data = map[string]interface{}{
		"friend_id": originUserId,
		"rank":      rank + 1,
	}

	str, err := util.InterfaceToJsonString(data)
	if err != nil {
		core.Logger.Error("[SaveRankUpgrade]userId:%v, originUserId:%v, rank:%v, err:%v", userId, originUserId, rank, err.Error())
	}
	core.RedisDo(core.RedisClient0, "set", cacheKey, str)
	core.Logger.Debug("[SaveRankUpgrade]userId:%v, originUserId:%v, rank:%v", userId, originUserId, rank)
}

// IncrDayPlayTimes 增加用户每日游戏次数
func IncrDayPlayTimes(userId int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_PLAY_TIMES, util.GetYMD(), userId)
	_, err := core.RedisDo(core.RedisClient0, "incr", cacheKey)
	if err == nil {
		// 设置过期时间
		core.RedisDo(core.RedisClient0, "expire", cacheKey, 86400)
	} else {
		core.Logger.Warn("[IncrDayPlayTimes]写数据进redis失败,err:%v", err.Error())
	}
}

// GetDayPlayTimes 获取用户每日游戏次数
func GetDayPlayTimes(userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_PLAY_TIMES, util.GetYMD(), userId)
	times, _ := core.RedisDoInt(core.RedisClient0, "get", cacheKey)
	core.Logger.Debug("[GetDayPlayTimes]userId:%v, cacheKey:%v, times:%v", userId, cacheKey, times)
	return times
}

// DelDayPlayTimes 删除用户每日游戏次数
func DelDayPlayTimes(userId int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_PLAY_TIMES, util.GetYMD(), userId)
	core.RedisDoInt(core.RedisClient0, "del", cacheKey)
}

func getGameLogKey(userId, otherUserId int, createTime int64) string {
	Ymd := time.Unix(createTime, 0).Format("20060102")
	if userId < otherUserId {
		return fmt.Sprintf(config.CACHE_KEY_RANK_TOGETHER_TIMES, Ymd, userId, otherUserId)
	}
	return fmt.Sprintf(config.CACHE_KEY_RANK_TOGETHER_TIMES, Ymd, otherUserId, userId)
}

// GetTogetherGameTimes 读取两个用户id的同时游戏次数
func GetTogetherGameTimes(userId, otherUserId int) int {
	if configService.IsRobot(userId) || configService.IsRobot(otherUserId) {
		return 0
	}
	times, _ := core.RedisDoInt(core.RedisClient0, "get", getGameLogKey(userId, otherUserId, util.GetTime()))
	return times
}

// AddTogetherGameTimes 增加两个用户的同时游戏次数
func AddTogetherGameTimes(userId, otherUserId int, createTime int64) {
	if configService.IsRobot(userId) || configService.IsRobot(otherUserId) {
		return
	}
	cacheKey := getGameLogKey(userId, otherUserId, createTime)
	_, err := core.RedisDo(core.RedisClient0, "incr", cacheKey)
	if err == nil {
		// 设置过期时间
		core.RedisDo(core.RedisClient0, "expire", cacheKey, 86400)
	} else {
		core.Logger.Warn("[AddTogetherGameTimes]写数据进redis失败,err:%v", err.Error())
	}
}

// GetSeasonUserWinningStreak 获取用户当前连胜次数
func GetSeasonUserWinningStreak(seasonId, userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_WINNING_STREAK, seasonId)
	times, _ := core.RedisDoInt(core.RedisClient0, "HGET", cacheKey, userId)
	return times
}

// SetSeasonUserWinningStreak 保存当前连胜次数
func SetSeasonUserWinningStreak(seasonId, userId, times int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_WINNING_STREAK, seasonId)
	core.RedisDo(core.RedisClient0, "HSET", cacheKey, userId, times)

	SetSeasonUserMaxWinningStreak(seasonId, userId, times)
}

// SetSeasonUserMaxWinningStreak 保存用户最高的连胜记录
func SetSeasonUserMaxWinningStreak(seasonId, userId, times int) {
	core.GetWriter().Raw("UPDATE season_users SET win_streak = ? where season_id= ? and user_id = ? and win_streak < ?",
		times, seasonId, userId, times).Exec()
}

// DelSeasonUserWinningStreak 清空当前连胜次数
func DelSeasonUserWinningStreak(seasonId, userId int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_RANK_WINNING_STREAK, seasonId)
	core.RedisDo(core.RedisClient0, "HDEL", cacheKey, userId)
}

// GetWinningStreakRewards 获取连胜次数奖励
func GetWinningStreakRewards(winningStreak int) float64 {
	rate, _ := core.RedisDoFloat64(core.RedisClient0, "HGET", config.CACHE_KEY_RANK_SEASON_GAME_WIN, winningStreak)
	return rate
}

// NotifyWinningStreakRewards 通知游戏服发放连胜奖励
func NotifyWinningStreakRewards(userId int, winningStreak int) {
	url := fmt.Sprintf("%v?user_id=%v&wins=%v", core.AppConfig.RankWinStreakRewardsURL, userId, winningStreak)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[NotifyWinningStreakRewards]het.Get, userId:%v, url:%v, error:%v", userId, url, err.Error())
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			core.Logger.Error("[NotifyWinningStreakRewards]het.Get read body, raceId:%v, url:%v, error:%v", url, err.Error())
		} else {
			core.Logger.Info("[NotifyWinningStreakRewards]success, userId:%v, url:%v, result:%v", userId, url, string(body))
		}
	}
}

// GetProvinceRankRewards 获取排位赛排名奖励
func GetProvinceRankRewards(rank int) int {
	rewards, _ := core.RedisDoInt(core.RedisClient0, "hget", config.CACHE_KEY_RANK_SEASON_PROVINCE_RANK_REWARDS, rank)
	return rewards
}

// GetGradeRewards 获取排位赛段位奖励
func GetGradeRewards(grade int) int {
	rewards, _ := core.RedisDoInt(core.RedisClient0, "hget", config.CACHE_KEY_RANK_SEASON_GRADE_REWARDS, grade)
	return rewards
}

// BuildMesgFlag 构建消息标记
func BuildMesgFlag(title string, userId int) string {
	return fmt.Sprintf("%v:%v", title, userId)
}

// CheckMsgSent 检查某个消息是否已发送
func CheckMsgSent(flag string) bool {
	cacheKey := fmt.Sprintf("RANK:MESSAGE:SENT:%v:%v", flag, util.GetYMD())
	exists, _ := core.RedisDoBool(core.RedisClient2, "exists", cacheKey)
	return exists
}

// SetMsgSent 标记某个消息是否已发送
func SetMsgSent(flag string) {
	cacheKey := fmt.Sprintf("RANK:MESSAGE:SENT:%v:%v", flag, util.GetYMD())
	core.RedisDo(core.RedisClient2, "set", cacheKey, 1)
	core.RedisDo(core.RedisClient2, "expire", cacheKey, 86400)
}

// SendGradeUpMessage 发放段位升级消息
func SendGradeUpMessage(userId int, grade int) {
	// 雀王和雀神才发
	if grade < 5 {
		return
	}

	// 检查今天有没有发给用户
	flag := BuildMesgFlag("GRADE", userId)
	if CheckMsgSent(flag) {
		return
	}

	// 读取奖励内容
	rewards := GetGradeRewards(grade)
	if rewards <= 0 {
		return
	}

	var gradeName string
	if grade == 5 {
		gradeName = "雀王"
	} else {
		gradeName = "雀神"
	}
	u := userService.GetUser(userId)
	nickname := u.Nickname
	avatar := userService.GetUserAvatar(userId)
	avatarBox := userService.GetUserAvatarBox(userId)
	content := fmt.Sprintf("恭喜【[[a]] %v】荣升%v，获得瓜分%v元奖金资格！", nickname, gradeName, rewards)
	p := gateway.PrivateMessagePush(userId, fbsCommon.MessageIdRANK_GRADE_UP, content, avatar, avatarBox)
	gateway.SendBroadcastMessage(p.Serialize())

	core.Logger.Debug("[SendGradeUpMessage]userId:%v, grade:%v, content:%v", userId, grade, content)

	// 记录今天发放标记
	SetMsgSent(flag)
}

// SendProvinceRankMessage 发放排名奖励信息
func SendProvinceRankMessage(userId int, rank int) {
	// 只有前10名，才发
	if rank > 10 {
		return
	}
	// 检查今天有没有发给用户
	flag := BuildMesgFlag("RANK", userId)
	if CheckMsgSent(flag) {
		return
	}
	// 读取奖励内容
	rewards := GetProvinceRankRewards(rank)
	if rewards <= 0 {
		return
	}

	u := userService.GetUser(userId)
	nickname := u.Nickname
	avatar := userService.GetUserAvatar(userId)
	avatarBox := userService.GetUserAvatarBox(userId)
	content := fmt.Sprintf("恭喜【[[a]] %v】晋升全省第%v名，获得领取%v元排名奖金资格！", nickname, rank, rewards)
	p := gateway.PrivateMessagePush(userId, fbsCommon.MessageIdRANK_PROVICE_RANK, content, avatar, avatarBox)
	gateway.SendBroadcastMessage(p.Serialize())

	core.Logger.Debug("[SendProvinceRankMessage]userId:%v, rank:%v, content:%v", userId, rank, content)

	// 记录今天发放标记
	SetMsgSent(flag)
}
