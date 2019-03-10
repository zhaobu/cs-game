package servers

import (
	"encoding/json"
	"mahjong-select-server/config"
	"mahjong-select-server/core"

	"github.com/fwhappy/util"
)

// GameServer 游戏服务器配置
type GameServer struct {
	IP         string
	Port       string
	Remote     string
	Domain     string
	Enable     string // ”0“清人状态，"1"正常状态
	ServerType string `json:"type"`
}

// 游戏服列表
var gameServers []*GameServer

// 游戏服务器最后加载时间
var gameServersLastLoadTime int64

func init() {
	gameServers = make([]*GameServer, 0)
}

// GetGameServers 返回游戏服列表
func GetGameServers() []*GameServer {
	if util.GetTime()-gameServersLastLoadTime >= config.GAME_SERVERS_TIMEOUT {
		core.Logger.Debug("[servers.GameServers]游戏服列表数据超时，重新从DB读取")
		gameServersLastLoadTime = util.GetTime()
		loadGameServersFromDB()
	}
	return gameServers
}

// 从数据库读取游戏服列表
func loadGameServersFromDB() {
	data, err := core.RedisDoStringMap(core.RedisClient4, "hgetall", config.CACHE_KEY_GAME_SERVERS)
	if err != nil {
		core.Logger.Error("[servers.GameServers]从redis读取游戏服列表配置失败,err:%v", err.Error())
		gameServers = make([]*GameServer, 0)
		return
	}
	if len(data) == 0 {
		core.Logger.Debug("[servers.GameServers]后台未配置游戏服列表")
		gameServers = make([]*GameServer, 0)
		return
	}

	tmpGameServers := make([]*GameServer, 0, len(data))
	for _, server := range data {
		gameServer := &GameServer{}
		json.Unmarshal([]byte(server), gameServer)
		gameServer.Remote = gameServer.IP + ":" + gameServer.Port
		tmpGameServers = append(tmpGameServers, gameServer)
	}
	gameServers = tmpGameServers
	core.Logger.Debug("[servers.GameServers]游戏服列表读取成功:")
	for _, s := range gameServers {
		core.Logger.Debug("[servers.GameServers]server:%#v", s)
	}
}

// InitGameServers 加载游戏服列表
func InitGameServers() {
	gameServersLastLoadTime = util.GetTime()
	loadGameServersFromDB()

	core.Logger.Debug("[InitGameServers]启动初始化游戏服配置完成:")
	for _, s := range gameServers {
		core.Logger.Debug("[servers.GameServers]server:%#v", s)
	}
}
