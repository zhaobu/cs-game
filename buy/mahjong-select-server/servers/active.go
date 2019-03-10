package servers

import (
	"mahjong-select-server/config"
	"mahjong-select-server/core"

	"github.com/fwhappy/util"
)

// 当前活跃服务器列表
var activeServers []string

// 活跃服务器最后加载时间
var activeServersLastLoadTime int64

func init() {
	activeServers = make([]string, 0)
}

// IsActive 当前服务器是否是激活服务器
func IsActive(remote string) bool {
	if !util.InStringSlice(remote, activeServers) {
		core.Logger.Error("[servers.active]IsActive:%v", false)
		return false
	}
	return true
	/*
		isExists, err := core.RedisDoBool(core.RedisClient3, "hexists", config.CACHE_KEY_REMOTE_USER_CNT, remote)
		if err != nil {
			core.Logger.Error("[servers.active]IsActive:%v", err.Error())
			return false
		}
		return isExists
	*/
}

// GetActiveServers 获取所有的有效服务器
func GetActiveServers() []string {
	if util.GetTime()-activeServersLastLoadTime >= config.ACTIVE_TIMEOUT {
		core.Logger.Debug("[servers.activeServers]当前活跃游戏服数据超时，重新从DB读取")
		activeServersLastLoadTime = util.GetTime()
		loadActiveServersFromDB()
	}
	return activeServers
}

func loadActiveServersFromDB() {
	// 读取当前活跃中的服务器
	currentActiveServers := make([]string, 0)
	servers, err := core.RedisDoStringMap(core.RedisClient3, "hgetall", config.CACHE_KEY_REMOTE_USER_CNT)
	if err != nil {
		core.Logger.Error("[servers.active]GetActiveServers:%v", err.Error())
		return
	}
	for remote := range servers {
		currentActiveServers = append(currentActiveServers, remote)
	}
	// 覆盖到当前变量
	activeServers = currentActiveServers
}

// InitActiveServers 启动app时，初始化一次有效服务器
func InitActiveServers() {
	activeServersLastLoadTime = util.GetTime()
	loadActiveServersFromDB()

	core.Logger.Debug("[InitActiveServers]启动初始化活跃游戏服信息完成:%#v", activeServers)
}
