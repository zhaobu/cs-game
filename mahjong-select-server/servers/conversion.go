package servers

import (
	"mahjong-select-server/config"
	"mahjong-select-server/core"
	"strings"
	"sync"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
)

// IPConversionMap ip转换地址map
type IPConversionMap struct {
	mux *sync.Mutex
	ips map[string][]string
}

// IPMap slb转换库
var IPMap *IPConversionMap

// slb地址转换最后读取时间
var ipConversionLastLoadTime int64

func init() {
	IPMap = &IPConversionMap{
		mux: &sync.Mutex{},
		ips: make(map[string][]string),
	}
}

// Conversion 将服务器ip，转换成slb ip
func (im *IPConversionMap) Conversion(remote string) string {
	im.mux.Lock()
	defer im.mux.Unlock()

	s := strings.Split(remote, ":")
	if len(s) != 2 {
		core.Logger.Debug("[servers.Conversion]remote格式错误,直接返回原值,remote:%v", remote)
		return remote
	}
	ip, port := s[0], s[1]

	// 超时重读
	if util.GetTime()-ipConversionLastLoadTime >= config.CONVERSION_TIMEOUT {
		core.Logger.Debug("[servers.Conversion]slb转换库数据超时，重新从数据库读取")
		ipConversionLastLoadTime = util.GetTime()
		im.loadIPConversionMapFromDB()
	}
	if ipList, exists := im.ips[ip]; exists && len(ipList) > 0 {
		remote = ipList[0] + ":" + port
		core.Logger.Debug("[servers.Conversion]找到slb对应关系,origin:%v,to:%v,转换后的remote:%v", ip, ipList[0], remote)
		return remote
	}
	core.Logger.Error("[servers.Conversion]转换失败,返回原remote:%v", remote)
	return remote
}

// 读取所有不健康的slb地址列表
func getInvalidSlbs() []string {
	invalidSlbs, err := core.RedisDoStrings(core.RedisClient4, "smembers", config.CACHE_KEY_SLBS_INVALID)
	if err != nil {
		core.Logger.Error("[servers.conversion]getInvalidSlbs失败,err:%v", err.Error())
	}
	return invalidSlbs
}

// 重新从数据库中加载ip转换库
func (im *IPConversionMap) loadIPConversionMapFromDB() {
	invalidSlbs := getInvalidSlbs()
	core.Logger.Debug("[servers.conversion]读取无效slb列表, invalidSlbs:%v", invalidSlbs)

	// 读取服务器配置
	servers, err := core.RedisDoStringMap(core.RedisClient4, "hgetall", config.CACHE_KEY_SERVER_LISTS)
	if err != nil {
		core.Logger.Error("[servers.conversion]从redis读取slb转换配置失败,err:%v", err.Error())
	}
	core.Logger.Debug("[servers.conversion]读取到slb转换列表:%v", servers)

	// 清空原转换规则
	im.ips = make(map[string][]string)
	for _, v := range servers {
		server, err := simplejson.NewJson([]byte(v))
		if err != nil {
			core.Logger.Error("[servers.conversion]解析server失败,err:%v", err.Error())
			continue
		}
		serverIP, _ := server.Get("ip").String()
		slbs, err := server.Get("slbs").StringArray()
		if err != nil {
			core.Logger.Error("[servers.conversion]解析server slbs失败,err:%v", err.Error())
		}
		if len(slbs) == 0 {
			continue
		}

		slbIPs := make([]string, 0)
		for _, ip := range slbs {
			// 过滤掉异常的slb
			if len(invalidSlbs) > 0 && util.InStringSlice(ip, invalidSlbs) {
				continue
			}
			slbIPs = append(slbIPs, ip)
		}
		// 必须保留一个slb
		if len(slbIPs) == 0 {
			slbIPs = append(slbIPs, slbs[0])
		}
		im.ips[serverIP] = slbIPs
	}
	core.Logger.Debug("[server.conversion]生成slb转换列表成功,ips:%#v", im.ips)
}
