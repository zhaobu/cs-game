package user

import (
	"strconv"

	"mahjong.go/config"
	"mahjong.go/library/core"
	userModel "mahjong.go/model/user"
)

// 读取用户的扩展信息
func GetUserInfoList(userId int) config.UserInfoList {
	// TODO 从cache中获取

	userInfoList := userModel.GetUserInfoList(userId)

	return userInfoList
}

// 读取用户积分
func GetScore(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_SCORE)
}

// 读取用户随机组局累积积分
func GetScoreRandom(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_SCORE_RANDOM)
}

// 读取用户比赛累积积分
func GetScoreMatch(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_SCORE_MATCH)
}

// GetScoreCoin 读取金币场积分
// 读取用户比赛累积积分
func GetScoreCoin(ui config.UserInfoList) int {
	return getInfoInt(ui, config.USER_INFO_TYPE_SCORE_COIN)
}

// 读取用户性别
func GetGender(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_GENDER)
}

// 读取device_code
func GetDeviceCode(this config.UserInfoList) string {
	return getInfoString(this, config.USER_INFO_TYPE_DEVICE_CODE)
}

// 读取城市
func GetCity(this config.UserInfoList) string {
	return getInfoString(this, config.USER_INFO_TYPE_CITY)
}

// GetGameCity 读取城市
func GetGameCity(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_GAME_CITY)
}

// GetRankCards 读取用户排位赛参赛卡
func GetRankCards(this config.UserInfoList) int {
	return getInfoInt(this, config.USER_INFO_TYPE_RANK_CARDS)
}

func getInfoInt(this config.UserInfoList, infoType int) int {
	if userInfo, ok := this[infoType]; ok {
		val, _ := strconv.Atoi(userInfo.Info)
		return val
	} else {
		return 0
	}
}

func getInfoString(this config.UserInfoList, infoType int) string {
	if userInfo, ok := this[infoType]; ok {
		return userInfo.Info
	} else {
		return ""
	}
}

// 更新user_info表中value为整形且累加的字段
func updateAccumulative(userId, infoType, value int) *core.Error {
	userInfoList := GetUserInfoList(userId)

	var err *core.Error
	if userInfo, ok := userInfoList[infoType]; ok {
		// 已存在，更新
		currentScore, _ := strconv.Atoi(userInfo.Info)
		userInfo.Info = strconv.Itoa(currentScore + value)
		err = userModel.UpdateInfo(userInfo)
	} else {
		// 不存在，插入
		_, err = userModel.InsertInfo(userId, infoType, strconv.Itoa(value))
	}

	if err != nil {
		core.Logger.Error("[userinfoServer.updateAccumulative]userId:%v,infoType:%v,value:%v,Error: %s.", userId, infoType, value, err.Error())
	}

	return err
}

// 更新user_info表中覆盖value的字段
func updateCover(userId, infoType int, value string) *core.Error {
	userInfoList := GetUserInfoList(userId)

	var err *core.Error
	if userInfo, ok := userInfoList[infoType]; ok {
		// 已存在，更新
		userInfo.Info = value
		err = userModel.UpdateInfo(userInfo)
	} else {
		// 不存在，插入
		_, err = userModel.InsertInfo(userId, infoType, value)
	}
	if err != nil {
		core.Logger.Error("[userinfoServer.updateCover]userId:%v,infoType:%v,value:%v,Error: %s.", userId, infoType, value, err.Error())
	} else {
		core.Logger.Debug("[updateCover],userId:%v,infoType:%v,value:%v", userId, infoType, value)
	}

	return err
}

// UpdateScore 更新用户总分
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdateScore(userId, score int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_SCORE, score)
}

// UpdateScoreRandom 更新用户的随机积分
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdateScoreRandom(userId, score int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_SCORE_RANDOM, score)
}

// UpdateScoreMatch 更新用户的比赛积分
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdateScoreMatch(userId, score int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_SCORE_MATCH, score)
}

// UpdateScoreCoin 更新用户的金币数
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdateScoreCoin(userId, score int) *core.Error {
	return updateCover(userId, config.USER_INFO_TYPE_SCORE_COIN, strconv.Itoa(score))
}

// UpdatePunishmentTimes 更新用户的惩罚次数
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdatePunishmentTimes(userId int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_PUNISHMENT_TIMES, 1)
}

// UpdateRankCards 更新用户的惩罚次数
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdateRankCards(userId, amount int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_RANK_CARDS, amount)
}

// UpdatePunishmentFlag 更新用户的比赛积分
// 如果记录不存在，新增记录，若已存在，更新数值
func UpdatePunishmentFlag(userId int) *core.Error {
	return updateAccumulative(userId, config.USER_INFO_TYPE_PUNISHMENT_FLAG, 1)
}
