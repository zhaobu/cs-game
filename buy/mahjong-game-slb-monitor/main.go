package main

import (
	"flag"
	"mahjong-game-slb-monitor/core"
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	ping "github.com/sparrc/go-ping"

	"github.com/fwhappy/mail/exmail"
	"github.com/fwhappy/util"
)

var (
	// 配置文件夹, 最好是绝对路劲, /Home/xh/goim/etc/local
	confDir = flag.String("confDir", "etc/local", "config dir path")
)

// RETRY_TIMES 重试次数
const RETRY_TIMES = 5

// 缓存键值配置
// CACHE_KEY_SLB_SERVER_LISTS 游戏服务器列表
// CACHE_KEY_SLBS_INVALID 异常SLB列表
const (
	CACHE_KEY_SLBS_INVALID     = "SLBS:INVALID"
	CACHE_KEY_SLB_SERVER_LISTS = "SERVER:LISTS"
)

// slb 对象
type slb struct {
	ip           string
	status       bool
	failTimes    int
	successTimes int
}

// slb列表
// 0: 成功次数; 1:失败次数
// 在线状态下，连续失败RETRY_TIMES次，则认为服务器已离线
// 离线状态下，连续成功RETRY_TIMES次，则认为服务器已上线
var pool map[string]*slb

func init() {
	flag.Parse()
	pool = make(map[string]*slb)
}

func main() {
	defer util.RecoverPanic()

	// 加载app配置
	core.LoadAppConfig(core.GetConfigFile("app.toml", *confDir))
	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *confDir))
	defer core.Logger.Flush()
	// 初始化Redis配置
	core.LoadRedisConfig(core.GetConfigFile("redis.toml", *confDir))

	startMonitor()
}

// 开始监控
func startMonitor() {
	lastLoadPoolTime := util.GetTime()
	for {
		// 每秒执行一次
		time.Sleep(time.Second)

		// 每10秒加载一次redis配置
		currentTime := util.GetTime()
		if currentTime-lastLoadPoolTime >= 10 || len(pool) == 0 {
			loadPool()
			lastLoadPoolTime = currentTime
		}

		monitorPool()
	}
}

// 执行监控
func monitorPool() {
	if len(pool) == 0 {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(pool))
	for _, s := range pool {
		// go monitorServer(s, wg)
		monitorServer(s, wg)
	}
	wg.Wait()
}

// ping 服务器
func monitorServer(s *slb, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	pinger, err := ping.NewPinger(s.ip)
	if err != nil {
		core.Logger.Error("NewPinger ERROR: %s", err.Error())
		return
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		// core.Logger.Debug("%d bytes from %s: icmp_seq=%d time=%v", pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		// core.Logger.Debug("--- %s ping statistics ---", stats.Addr)
		// core.Logger.Debug("%d packets transmitted, %d packets received, %v%% packet loss", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		core.Logger.Debug("round-trip min/avg/max/stddev = %v/%v/%v/%v", stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)

		if stats.PacketsRecv == 0 {
			core.Logger.Debug("ping %v fail", s.ip)
		}

		// 收到响应
		if stats.PacketsRecv >= 1 {
			s.successTimes++
			s.failTimes = 0
		} else {
			s.failTimes++
			s.successTimes = 0
		}

		// 检测到服务器异常，下线
		if s.status && s.failTimes >= RETRY_TIMES {
			offlineServer(s)
		}
		// 检测到服务器恢复，上线
		if !s.status && s.successTimes >= RETRY_TIMES {
			onlineServer(s)
		}
	}

	core.Logger.Debug("PING %s (%s):", pinger.Addr(), pinger.IPAddr())

	pinger.Count = 1
	pinger.Interval = time.Second
	pinger.Timeout = time.Second
	pinger.SetPrivileged(false)
	pinger.Run()
}

// 加载slb池
func loadPool() {
	gameServers, err := core.RedisDoStringMap("hgetall", CACHE_KEY_SLB_SERVER_LISTS)
	if err != nil {
		core.Logger.Error("读取game server list 失败,%v", err.Error)
		return
	}

	slbIPList := []string{}
	for _, v := range gameServers {
		server, _ := simplejson.NewJson([]byte(v))
		if err != nil {
			core.Logger.Error("解析game server数据失败,%v", err.Error)
			continue
		}

		slbs, err := server.Get("slbs").StringArray()
		if err != nil {
			core.Logger.Error("解析game server slbs 数据失败,%v", err.Error)
		}

		for _, ip := range slbs {
			if !util.InStringSlice(ip, slbIPList) {
				slbIPList = append(slbIPList, ip)
			}
		}
	}

	if len(slbIPList) == 0 {
		core.Logger.Error("未找到slb配置")
		return
	}

	// 从pool中移除已删除的slb配置
	for k, s := range pool {
		if !util.InStringSlice(s.ip, slbIPList) {
			delete(pool, k)
			core.Logger.Info("移除监控服务器:%v", s.ip)
		}
	}

	// 添加新的slb配置进pool
	for _, ip := range slbIPList {
		if _, exists := pool[ip]; !exists {
			pool[ip] = &slb{ip: ip, status: isOnline(ip), successTimes: 0, failTimes: 0}
			core.Logger.Info("新增监控服务器:%v, status:%v", ip, pool[ip].status)
		}
	}

	// core.Logger.Debug("加载slb pool 完成:%v", pool)
	core.Logger.Debug("加载slb pool 完成")
}

func onlineServer(s *slb) {
	_, err := core.RedisDo("srem", CACHE_KEY_SLBS_INVALID, s.ip)
	if err != nil {
		core.Logger.Error("上线线slb失败:%v", s.ip)
		return
	}
	s.status = true

	core.Logger.Info("onlineServer:%v", s.ip)

	body := ""
	body += "SLB报警：<br />"
	body += "IP：" + s.ip + "<br />"
	body += "类型: 恢复服务<br />"
	body += "时间:" + util.GetTimestamp() + "<br />"

	// 短信通知、邮件通知
	go sendAlarm(body)
}

func offlineServer(s *slb) {
	_, err := core.RedisDo("sadd", CACHE_KEY_SLBS_INVALID, s.ip)
	if err != nil {
		core.Logger.Error("下线slb失败:%v", s.ip)
		return
	}
	s.status = false

	core.Logger.Info("offlineServer:%v", s.ip)

	// 短信通知、邮件通知
	body := ""
	body += "SLB报警：<br />"
	body += "IP：" + s.ip + "<br />"
	body += "类型: 下线报警<br />"
	body += "时间:" + util.GetTimestamp() + "<br />"

	sendAlarm(body)
}

func isOnline(ip string) bool {
	isOnline, _ := core.RedisDoBool("sismember", CACHE_KEY_SLBS_INVALID, ip)
	return !isOnline
}

func sendAlarm(body string) {
	host := core.AppConfig["alarm_host"].(string)
	port := int(core.AppConfig["alarm_port"].(int64))
	email := core.AppConfig["alarm_email"].(string)
	password := core.AppConfig["alarm_password"].(string)
	addressList := core.AppConfig["alarm_address_list"].(string)
	from := core.AppConfig["alarm_from"].(string)
	subject := core.AppConfig["alarm_subject"].(string)

	// 发送邮件
	sendErrors := exmail.Send(host, port, email, password, addressList, from, subject, body)
	if len(sendErrors) > 0 {
		for _, err := range sendErrors {
			core.Logger.Error("Send Mail error:%v", err.Error)
		}
	}

	// TODO 发送短信
}
