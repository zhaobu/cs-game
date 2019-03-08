package room

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/fwhappy/util"
	"github.com/garyburd/redigo/redis"
	"mahjong.go/config"
	"mahjong.go/library/core"
	configService "mahjong.go/service/config"
)

// 生成根据number查找id的cachekey
func getRoomNumberIdCacheKey(number string) string {
	return fmt.Sprintf(config.CACHE_KEY_ROOM_NUMBER_ID, number)
}

// 生成根据id查找number的cachekey
func getRoomIdNumberCacheKey(roomId int64) string {
	return fmt.Sprintf(config.CACHE_KEY_ROOM_ID_NUMBER, roomId)
}

// 获取房间所在线路的cachekey
func getRoomRemoteCacheKey(roomId int64) string {
	return fmt.Sprintf(config.CACHE_KEY_ROOM_REMOTE, roomId)
}

// 创建房间成功之后, 需保存数据至cache中
// fixme 这里有点啰嗦，写入次数比较多，后续优化
func SaveRoom(roomId int64, number string, remote string) bool {
	redisConn := core.RedisClient1.Get()
	defer redisConn.Close()

	// 保存房间number与房间id的对应关系, 如果对应关系存在，返回失败
	success, _ := redis.Bool(redisConn.Do("setnx", getRoomNumberIdCacheKey(number), roomId))
	if !success {
		return false
	}

	// 保存房间id与number的对应关系
	redisConn.Do("set", getRoomIdNumberCacheKey(roomId), number)

	// 保存房间number所在的remote
	redisConn.Do("set", getRoomRemoteCacheKey(roomId), remote)

	return true
}

// 房间正常完成或者异常结束之后，需要清除房间关联的数据
func CleanRoom(roomId int64, number string) {
	redisConn := core.RedisClient1.Get()
	defer redisConn.Close()

	// 删除房间number与房间id的对应关系
	redisConn.Do("del", getRoomNumberIdCacheKey(number))

	// 删除房间id与number的对应关系
	redisConn.Do("del", getRoomIdNumberCacheKey(roomId))

	// 删除房间number所在的remote
	redisConn.Do("del", getRoomRemoteCacheKey(roomId))
}

// 设置房间指令的缓存有效期
func SetRoomExpire(roomId int64, number string, expire int) {
	redisConn := core.RedisClient1.Get()
	defer redisConn.Close()
	redisConn.Do("expire", getRoomNumberIdCacheKey(number), expire)
	redisConn.Do("expire", getRoomIdNumberCacheKey(roomId), expire)
	redisConn.Do("expire", getRoomRemoteCacheKey(roomId), expire)
}

// 根据房间号获取房间id
func GetRoomIdByNumber(number string) int64 {
	roomId, _ := core.RedisDoInt64(core.RedisClient1, "get", getRoomNumberIdCacheKey(number))
	return roomId
}

// 根据房间id获取房间号
func GetRoomNumberById(roomId int64) string {
	number, _ := core.RedisDoString(core.RedisClient1, "get", getRoomIdNumberCacheKey(roomId))
	return number
}

// 判断房间号是否已存在
func IsExists(number string) bool {
	// 连接redis对象
	redisConn := core.RedisClient1.Get()
	defer redisConn.Close()
	exists, _ := redis.Bool(redisConn.Do("exists", getRoomNumberIdCacheKey(number)))

	return exists
}

// 创建一个房间id，自增不重复
func GenRoomId() int64 {
	// TODO 加锁，从数据库恢复目前最大值

	// 连接redis对象
	redisConn := core.RedisClient1.Get()
	defer redisConn.Close()
	roomId, _ := redis.Int64(redisConn.Do("incr", config.CACHE_KEY_ROOM_BUILDER))

	return roomId
}

// 找到一个可用的房间号
func GetRoomNumber() string {
	var number string
	for {
		number = genRoomNumber()
		if !IsExists(number) {
			break
		}

		// TODO 判断最多执行次数
	}

	return number
}

// 生成一个房间号
func genRoomNumber() string {
	var buffer bytes.Buffer

	for i := 0; i < config.ROOM_NUMBER_LENGTH; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		buffer.WriteString(strconv.Itoa(r.Intn(10)))
	}

	return buffer.String()
}

// 读取房间id所在的remote
func GetRoomRemote(roomId int64) string {
	remote, _ := core.RedisDoString(core.RedisClient1, "get", getRoomRemoteCacheKey(roomId))

	return remote
}

// StatGrowth 记录房间关联的俱乐部的成长值
func StatGrowth(clubId int) {
	data := make(map[string]interface{})
	data["userId"] = clubId
	data["type"] = 4
	data["num"] = 1
	str, _ := util.InterfaceToJsonString(data)
	core.RedisDo(core.RedisClient4, "rpush", config.CACHE_KEY_ROOM_CLUB_GROWTH, str)
}

// AppendClubCreateRoomPush 将俱乐部创建消息写入消息推送队列
func AppendClubCreateRoomPush(clubId int, roomId int64, number string, creator int, settingSlice []int) {
	data := make(map[string]interface{})
	data["push_type"] = "create_room"
	data["club_id"] = clubId
	data["room_id"] = roomId
	data["number"] = number
	data["creator"] = creator
	data["setting"] = settingSlice
	byt, err := util.InterfaceToJsonString(data)
	if err != nil {
		core.Logger.Warn("[AppendClubCreateRoomPush]failure, error:%v", err.Error())
		return
	}
	core.RedisDo(core.RedisClient3, "lpush", config.CACHE_KEY_CLUB_PUSH_LIST, byt)
}

// SaveResult 存储游戏结果
func SaveResult(roomId int64, data []byte) {
	cacheKey := fmt.Sprintf(config.CACHE_KEY_ROOM_RESULT, roomId)
	core.RedisDo(core.RedisClient5, "set", cacheKey, data)
	core.RedisDo(core.RedisClient5, "expire", cacheKey, config.CACHE_KEY_ROOM_RESULT_EXPIRE)
	core.Logger.Info("[room.SaveResult]roomId:%v", roomId)
}

// GetDrawEffectExtraRate 获取房间用户的好牌率
func GetDrawEffectExtraRate(userId, cType, gradeId int, leagueId, coinType int) int {
	var userType string
	if configService.IsRobot(userId) {
		userType = "ROBOT"
	} else {
		userType = "USER"
	}

	var hashKey string
	switch cType {
	case config.ROOM_TYPE_LEAGUE:
		hashKey = fmt.Sprintf("%v:%v:%v", userType, cType, leagueId)
	case config.ROOM_TYPE_RANK:
		hashKey = fmt.Sprintf("%v:%v:%v", userType, cType, gradeId)
	case config.ROOM_TYPE_COIN:
		hashKey = fmt.Sprintf("%v:%v:%v", userType, cType, coinType)
	default:
		hashKey = fmt.Sprintf("%v:%v:%v", userType, cType, 0)
	}
	v, _ := core.RedisDoInt(core.RedisClient0, "HGET", config.CACHE_KEY_DRAW_EXPECT_EXTRA_RATE, hashKey)
	core.Logger.Debug("[roomService.GetDrawEffectExtraRate]userId:%v, hashKey:%v, rate:%v", userId, hashKey, v)

	return v
}
