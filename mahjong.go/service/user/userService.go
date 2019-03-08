package user

import (
	"fmt"

	//	"github.com/garyburd/redigo/redis"
	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	userModel "mahjong.go/model/user"
	configService "mahjong.go/service/config"
	logService "mahjong.go/service/log"
)

// 定义redis连接
// 根据id读取用户信息
func GetUser(userId int) *config.User {
	// TODO 从cache中读取

	// 从db中读取
	user := userModel.GetUser(userId)

	return user
}

// 获取用户房间id在缓存中的cache key
func getRoomIdCacheKey(userId int) string {
	return fmt.Sprintf(config.CACHE_KEY_USER_ROOM_ID, userId)
}

// 从redis中读取用户的roomId，判断用户是否在房间中
func GetRoomId(userId int) int64 {
	value, _ := core.RedisDoInt64(core.RedisClient2, "get", getRoomIdCacheKey(userId))

	return value
}

// 保存用户当前加入的房间号
func SetRoomId(userId int, roomId int64) {
	core.RedisDo(core.RedisClient2, "set", getRoomIdCacheKey(userId), roomId)
	// 为了防止用户不再上线，会导致redis中的此值不过期，所以设置一下过期时间(默认给3天)
	core.RedisDo(core.RedisClient2, "expire", getRoomIdCacheKey(userId), 259200)
}

// 删除用户已加入的房间号
// 只有当用户当前所在的房间id，与要删除的房间id相同时，才做删除操作
// 如果传入的roomId是0，表示强制清除
func DelRoomId(userId int, roomId int64) {
	if roomId == 0 || roomId == GetRoomId(userId) {
		core.RedisDo(core.RedisClient2, "del", getRoomIdCacheKey(userId))
	}
}

// 判断用户金额是否足够
func CheckMoneyEnough(userId int, amount int, user *config.User) bool {
	if user == nil {
		user = GetUser(userId)
	}
	if userModel.GetMoney(user) >= amount {
		return true
	} else {
		return false
	}
}

// 更新用户钻石, amount为负，表示扣钻石
// 优先扣除免费钻石，免费钻石不够扣的时候，再扣付费钻石
func UpdateMoney(ormObj orm.Ormer, user *config.User, amount int, changeType string, consumeType int) *core.Error {
	// 判断用户金额是否足够支付
	if !CheckMoneyEnough(user.UserId, amount*-1, user) {
		return core.NewError(-204, amount)
	}

	// 更新DB
	m, chargeErr := userModel.UpdateMoney(ormObj, user, amount)
	if chargeErr != nil {
		core.Logger.Error("更新用户钻石错误, userId: %d, amount: %d, err:%s", user.UserId, amount, chargeErr.Error())
		return chargeErr
	}

	// 钻石数量
	money := m[config.ENTITY_MODULE_DIAMOND]
	// 免费钻石数量
	giftMoney := m[config.ENTITY_MODULE_DIAMOND_FREE]
	// 正负值
	symbal := 1
	if amount < 0 {
		symbal = -1
	}

	// 记录操作日志
	sn := logService.GenSn(user.UserId)
	if err := logService.LogMoney(user.UserId, money*symbal, giftMoney*symbal, changeType, sn); err != nil {
		core.Logger.Error("记录消费日志错误: %s.", err.Error())
	}

	if amount > 0 {
		// 记录收入日志
		if money > 0 {
			logService.LogUserTransInfo(0, user.UserId, money, sn, 1, config.DIAMOND_TYPE_MONEY)
		}
		if giftMoney > 0 {
			logService.LogUserTransInfo(0, user.UserId, giftMoney, sn, 1, config.DIAMOND_TYPE_GIFT_MONEY)
		}
	} else {
		// 记录消耗日志
		logService.LogConsumeInfo(user.UserId, money, giftMoney, sn, "房费", consumeType)
	}

	return nil
}

func getRandomPayedCacheKey(userId int) string {
	return fmt.Sprintf(config.CACHE_KEY_RANDOM_PAYED, userId, util.GetYMD())
}

// 判断用户今日是否已付费随机房间
func IsRandomPayed(userId int) bool {
	value, _ := core.RedisDoInt64(core.RedisClient2, "get", getRandomPayedCacheKey(userId))
	return value > 0
}

// 设置用户当日随机房间已付费
func SetRandomPayed(userId int) {
	success, _ := core.RedisDoBool(core.RedisClient2, "setnx", getRandomPayedCacheKey(userId), 1)
	if success {
		// 过期时间1天
		core.RedisDo(core.RedisClient2, "expire", getRandomPayedCacheKey(userId), 86400)
	}
}

// 删除用户的分享回放记录
func DelSharePlayFlag(userId int) {
	core.RedisDo(core.RedisClient0, "del", fmt.Sprintf(config.CACHE_KEY_H5_SHARE, userId))
}

// GetUserAvatar 获取用户头像
func GetUserAvatar(userId int) string {
	avatarURL := core.AppConfig.UserAvatarUrl
	secret := util.Md5Sum(fmt.Sprintf("didao2016%v", userId))
	// core.Logger.Debug("secret:%v,0:%v,1:%v", secret, secret[:1], secret[1:2])
	return fmt.Sprintf(avatarURL, secret[:1], secret[1:2], userId) + "?" + util.GetYMD()
}

// SaveUserLastGame 存储用户最后游戏记录
func SaveUserLastGame(userIds []int, roomId int64) {
	if len(userIds) == 0 {
		return
	}
	for _, userId := range userIds {
		vs, _ := util.InterfaceToJsonString(
			map[string]interface{}{
				"room_id": roomId,
				"user_id": userId,
			})
		_, err := core.RedisDo(core.RedisClient2, "lpush", config.CACHE_KEY_USER_LAST_PLAY, vs)
		if err != nil {
			core.Logger.Warn("[SaveUserLastGame]error,userId:%v, roomId:%v, err:%v", userId, roomId, err.Error())
		}
	}
}

// SaveRoomResultUnread 保留用户的游戏结果未查看标志
func SaveRoomResultUnread(userIds []int, roomId int64) {
	if len(userIds) == 0 {
		return
	}
	for _, userId := range userIds {
		if configService.IsRobot(userId) {
			continue
		}
		cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_UNREAD_GAME_RESULT, userId)
		_, err := core.RedisDo(core.RedisClient5, "set", cacheKey, roomId)
		if err != nil {
			core.Logger.Warn("[SaveRoomResultUnread]error,userId:%v, roomId:%v, err:%v", userId, roomId, err.Error())
			return
		}
		core.RedisDo(core.RedisClient5, "expire", cacheKey, config.CACHE_KEY_ROOM_RESULT_EXPIRE)
		core.Logger.Debug("[SaveRoomResultUnread]userId:%v, roomId:%v", userId, roomId)
	}
}

// RemoveRoomResultUnread 删除用户的游戏结果未查看标志
func RemoveRoomResultUnread(userId int) {
	if configService.IsRobot(userId) {
		return
	}
	cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_UNREAD_GAME_RESULT, userId)
	_, err := core.RedisDo(core.RedisClient5, "del", cacheKey)
	if err != nil {
		core.Logger.Warn("[RemoveRoomResultUnread]error,userId:%v, err:%v", userId, err.Error())
		return
	}
}

// GetUserAvatarBox 获取用户头像框
func GetUserAvatarBox(userId int) int {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_AVATAR_BOX, userId)
	v, _ := core.RedisDoInt(core.RedisClient0, "get", cacheKey)
	return v
}

// GetUserMemberLevel 获取用户会员等级以及经验加成
func GetUserMemberLevel(userId int) (level, addExp int) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_USER_MEMBER_LEVEL, userId)
	level, _ = core.RedisDoInt(core.RedisClient0, "get", cacheKey)
	if level > 0 {
		addExp, _ = core.RedisDoInt(core.RedisClient0, "hget", config.CACHE_KEY_USER_MEMBER_LEVEL_ADD_EXP, level)
	}

	core.Logger.Debug("[GetUserMemberLevel]userId:%v, level:%v, addExp:%v", userId, level, addExp)
	return
}
