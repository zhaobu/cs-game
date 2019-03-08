package model

import (
	"mahjong-league/core"
)

// SetRemoteUserCnt 统计服务器实时人数
func SetRemoteUserCnt(remote string, cnt int) {
	core.RedisDo(core.RedisClient3, "hset", "HALL:REMOTE:USER:CNT", remote, cnt)
}

// SetRemoteActionTime 记录服务器的最后活动时间
func SetRemoteActionTime(remote string, actionTime int64) {
	core.RedisDo(core.RedisClient3, "hset", "HALL:REMOTE:ACTION:TIME", remote, actionTime)
}

// SetRemoteVersion 设置服务器版本的最后活动时间
func SetRemoteVersion(remote, version string) {
	core.RedisDo(core.RedisClient3, "hset", "HALL:REMOTE:VERSION", remote, version)
}
