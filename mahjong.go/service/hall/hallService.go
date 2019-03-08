package hall

import (
	"fmt"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 获取大厅房间集合的cachekey
func getHallRoomIdsCacheKey(remote string) string {
	return fmt.Sprintf(config.CACHE_KEY_HALL_ROOM_IDS, remote)
}

// 获取机器人房间列表的cachekey
func getHallRobotRoomListCacheKey(remote string) string {
	// return fmt.Sprintf(config.CACHE_KEY_HALL_ROBOT_ROOM_LIST, remote)
	return fmt.Sprintf(config.CACHE_KEY_HALL_ROBOT_ROOM_LIST)
}

// 从cache中删除大厅的统计信息
func DelRemote(remote string) {
	core.RedisDo(core.RedisClient3, "hdel", config.CACHE_KEY_REMOTE_USER_CNT, remote)
}

// 设置大厅的在线人数
func SetRemoteUserCnt(remote string, cnt int) {
	core.RedisDo(core.RedisClient3, "hset", config.CACHE_KEY_REMOTE_USER_CNT, remote, cnt)
}

// 设置大厅的最后活动时间
func SetRemoteActionTime(remote string, actionTime int64) {
	core.RedisDo(core.RedisClient3, "hset", config.CACHE_KEY_REMOTE_ACTION_TIME, remote, actionTime)
}

// 新增一个大厅房间到redis中
func AddHallRoom(remote string, roomId int64) {
	core.RedisDo(core.RedisClient3, "sadd", getHallRoomIdsCacheKey(remote), roomId)
}

// 从redis中删除一个大厅房间
func DelHallRoom(remote string, roomId int64) {
	core.RedisDo(core.RedisClient3, "srem", getHallRoomIdsCacheKey(remote), roomId)
}

// 清除redis中的大厅房间列表
func CleanHallRoom(remote string) {
	core.RedisDo(core.RedisClient3, "del", getHallRoomIdsCacheKey(remote))
}

// 清除redis中的机器人房间队列
func CleanHallRobotRoom(remote string) {
	core.RedisDo(core.RedisClient3, "del", getHallRobotRoomListCacheKey(remote))
}

// 添加一个房间到redis中的机器人房间队列
func AddHallRobotRoom(remote string, data string) {
	// core.RedisDo(core.RedisClient3, "lpush", getHallRobotRoomListCacheKey(remote), fmt.Sprintf("%v-%v-%v-%v", roomId, mType, cType, playerCnt))
	core.RedisDo(core.RedisClient3, "lpush", getHallRobotRoomListCacheKey(remote), data)
}

// 从redis中的机器人房间队列获取一个房间id
func GetHallRobotRoom(remote string) string {
	roomStr, _ := core.RedisDoString(core.RedisClient3, "rpop", getHallRobotRoomListCacheKey(remote))
	return roomStr

	/*
		if err == nil {
			roomInfo := strings.Split(roomStr, "-")
			roomId, _ := strconv.ParseInt(roomInfo[0], 10, 64)
			mType, _ := strconv.Atoi(roomInfo[1])
			cType, _ := strconv.Atoi(roomInfo[2])
			playerCnt, _ := strconv.Atoi(roomInfo[3])
			return roomId, mType, cType, playerCnt
		} else {
			return 0, 0, 0, 0
		}
	*/
}

// 判断用户房间是否存在
func IsRoomExists(remote string, roomId int64) bool {
	exists, _ := core.RedisDoBool(core.RedisClient3, "sismember", getHallRoomIdsCacheKey(remote), roomId)
	return exists
}

// 设置大厅的最后活动时间
func SetRemoteVersion(remote, version string) {
	core.RedisDo(core.RedisClient3, "hset", config.CACHE_KEY_REMOTE_VERSION, remote, version)
}

// IsRemoteActive 判断服务器是否活跃
func IsRemoteActive(remote string) bool {
	lastActionTime, _ := core.RedisDoInt64(core.RedisClient3, "hget", config.CACHE_KEY_REMOTE_ACTION_TIME, remote)
	return (util.GetTime() - lastActionTime) < int64(5)
}
