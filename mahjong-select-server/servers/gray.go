package servers

import (
	"mahjong-select-server/config"
	"mahjong-select-server/core"
	"strconv"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

// GrapRule 灰度发布规则
type GrapRule struct {
	remote       string
	versionRules []string
	percent      int
	whiteList    []int
}

// 灰度规则列表
var grapRules []*GrapRule

// 游戏服务器最后加载时间
var grayRulesLastLoadTime int64

// loadGrayRules 加载游戏服列表
func loadGrayRules() {
	if util.GetTime()-grayRulesLastLoadTime >= config.GRAY_TIMEOUT {
		core.Logger.Debug("[servers.GrayRules]灰度服务器配置数据超时，重新从DB读取")
		grayRulesLastLoadTime = util.GetTime()
		loadGrayRulesFromDB()
	}
}

// 从数据库读取灰度
func loadGrayRulesFromDB() {
	tmpGrayRules := make([]*GrapRule, 0)
	defer func() {
		grapRules = tmpGrayRules
	}()

	data, err := core.RedisDoStringMap(core.RedisClient4, "hgetall", config.CACHE_KEY_GAME_GRAY)
	if err != nil {
		core.Logger.Error("[servers.GrapRules]从redis读取灰度服务器配置失败,err:%v", err.Error())
		return
	}
	if len(data) == 0 {
		core.Logger.Debug("[servers.GrapRules]灰度服务器未设置")
		return
	}

	for _, v := range data {
		rule, err := simplejson.NewJson([]byte(v))
		if err != nil {
			core.Logger.Error("[servers.GrapRules]解析灰度规则出错:%v", err.Error())
			continue
		}
		serverStr, err := rule.Get("serverList").String()
		if err != nil {
			core.Logger.Error("[servers.GrapRules]解析灰度规则[serverList]出错:%v", err.Error())
			continue
		}
		percentStr, err := rule.Get("percent").String()
		if err != nil {
			core.Logger.Error("[servers.GrapRules]解析灰度规则[percent]出错:%v", err.Error())
			continue
		}
		versionRuleStr, err := rule.Get("versionRule").String()
		if err != nil {
			core.Logger.Error("[servers.GrapRules]解析灰度规则[versionRule]出错:%v", err.Error())
			continue
		}
		whiteListStr, err := rule.Get("whiteList").String()
		if err != nil {
			core.Logger.Error("[servers.GrapRules]解析灰度规则[versionRule]出错:%v", err.Error())
			continue
		}
		servers := strings.Split(serverStr, ",")
		versionRules := strings.Split(versionRuleStr, ",")
		percent, _ := strconv.Atoi(percentStr)
		whteLists := make([]int, 0)
		for _, idStr := range strings.Split(whiteListStr, ",") {
			id, _ := strconv.Atoi(idStr)
			whteLists = append(whteLists, id)
		}

		for _, server := range servers {
			grayRule := &GrapRule{
				remote:       server,
				versionRules: versionRules,
				percent:      percent,
				whiteList:    whteLists,
			}
			tmpGrayRules = append(tmpGrayRules, grayRule)
		}
	}
	core.Logger.Debug("[servers.GrapRules]灰度列表读取成功:")
	for _, r := range tmpGrayRules {
		core.Logger.Debug("[servers.GrapRules]rule:%#v", r)
	}
}

// CheckGray 检查用户能否进入灰度规则
func CheckGray(userId int, validServers map[string]bool, version string) string {
	loadGrayRules()

	if len(grapRules) > 0 {
		keys := util.ShuffleSliceInt(util.GenRangeInt(len(grapRules), 0))
		for _, key := range keys {
			gr := grapRules[key]

			// 跳过不匹配的服务器
			if _, exists := validServers[gr.remote]; !exists {
				continue
			}
			// 版本不匹配无法进入
			if !util.InStringSlice(version, gr.versionRules) {
				delete(validServers, gr.remote)
				continue
			}
			// 在白名单中，直接进入
			if util.IntInSlice(userId, gr.whiteList) {
				core.Logger.Debug("[CheckGray]用户处于灰度发布白名单中，进入灰度服务器,userId:%v,remote:%v", userId, gr.remote)
				return gr.remote
			}
			// 通过概率运算，进入灰度发布
			percent := util.RandIntn(100)
			if percent < gr.percent {
				core.Logger.Debug("[CheckGray]用户通过灰度概率计算，进入灰度服务器,userId:%v,remote:%v,require:%v,rand:%v",
					userId, gr.remote, gr.percent, percent)
				return gr.remote
			}
			delete(validServers, gr.remote)
		}
	}
	return ""
}

// InitGrayRules 开局初始化灰度服务器配置
func InitGrayRules() {
	grayRulesLastLoadTime = util.GetTime()
	loadGrayRulesFromDB()
	core.Logger.Debug("[InitGrayRules]启动初始化灰度服务器配置完成:")
	for _, r := range grapRules {
		core.Logger.Debug("[servers.GrayRules]gray rule:%#v", r)
	}
}
