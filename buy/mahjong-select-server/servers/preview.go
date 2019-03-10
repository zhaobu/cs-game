package servers

import (
	"encoding/json"
	"mahjong-select-server/config"
	"mahjong-select-server/core"
	"strconv"
	"strings"

	"github.com/fwhappy/util"
	"github.com/garyburd/redigo/redis"
)

// PreviewRule 预发布规则
type PreviewRule struct {
	Enable          string // 是否启用
	Remote          string `json:"server"` // 服务器地址 ""
	WhiteList       string // 白名单
	whiteListParsed []int  // 白名单解析成slice
}

// PRule 预发规则
var PRule *PreviewRule
var previewLastLoadTime int64

func init() {
	PRule = &PreviewRule{
		whiteListParsed: make([]int, 0),
	}
}

func (pr *PreviewRule) clean() {
	pr.Enable = "0"
	pr.Remote = ""
	pr.WhiteList = ""
	pr.whiteListParsed = []int{}
}

func (pr *PreviewRule) parse() {
	pr.whiteListParsed = []int{}
	if pr.WhiteList != "" {
		ids := strings.Split(pr.WhiteList, ",")
		for _, idString := range ids {
			id, _ := strconv.Atoi(idString)
			pr.whiteListParsed = append(pr.whiteListParsed, id)
		}
	}
}

// loadPreviewRule 获取预发布配置
func loadPreviewRule() {
	if util.GetTime()-previewLastLoadTime >= config.PREVIEW_TIMEOUT {
		core.Logger.Debug("[servers.preview]预发布规则超时，重新从DB读取")
		previewLastLoadTime = util.GetTime()
		loadReleaseRuleFromDB()
	}
}

// 从数据库读取预发布配置
func loadReleaseRuleFromDB() {
	data, err := core.RedisDoString(core.RedisClient4, "get", config.CACHE_KEY_GAME_PREVIEW)
	if err != nil && err != redis.ErrNil {
		core.Logger.Error("[servers.preview]从数据库读取预发布配置失败,err:%v", err.Error())
	}
	if data == "" {
		core.Logger.Debug("[servers.preview]预发布规则未配置")
		PRule.clean()
		return
	}

	json.Unmarshal([]byte(data), PRule)
	PRule.parse()

	core.Logger.Debug("[servers.preview]预发布规则读取成功:%#v", PRule)
}

// CheckPreview 检查是否可以匹配到预发环境
func (pr *PreviewRule) CheckPreview(userId int, validServers map[string]bool) string {
	loadPreviewRule()
	if pr.Remote != "" {
		if _, exists := validServers[pr.Remote]; !exists {
			return ""
		}
		if util.IntInSlice(userId, pr.whiteListParsed) {
			return pr.Remote
		}
		delete(validServers, pr.Remote)
	}
	return ""
}

// InitPreviewRule 开局初始化预发布配置
func InitPreviewRule() {
	previewLastLoadTime = util.GetTime()
	loadReleaseRuleFromDB()

	core.Logger.Debug("[InitPreviewRule]启动初始化预发布配置完成:%#v", PRule)
}
