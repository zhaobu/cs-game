package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"
	"mahjong.push/config"
	"mahjong.push/library/core"
	"mahjong.push/library/util"
	pushService "mahjong.push/service/push"
)

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "qa", "env")
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	defer util.RecoverPanic()

	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 开启pprof监控
	go func() {
		http.ListenAndServe(":7777", nil)
	}()

	// 初始化基础配置
	core.LoadAppConfig(fmt.Sprintf("conf/%s/app.toml", *env))
	// 初始化日志配置
	core.LoadLoggerConfig(fmt.Sprintf("conf/%s/log.toml", *env))
	defer core.Logger.Flush()
	// 初始化Redis配置
	core.LoadRedisConfig(fmt.Sprintf("conf/%s/redis.toml", *env))

	// 开启服务
	run()
}

// 读取redis对象
func getRedisConn() redis.Conn {
	redisConn := core.RedisClient1.Get()
	redisConn.Do("select", core.RedisDBPUSH)

	return redisConn
}

// 启动服务
func run() {
	// 读取redis队列中的内容
	redisConn := getRedisConn()

	for {
		data, err := redisConn.Do("rpop", config.CACHE_KEY_PUSH_QUEUE_LIST)
		if err != nil {
			// 这里如果出错了，表示连接异常了
			core.Logger.Error("read push list error: %v", err)
			// 重新获取redis连接
			time.Sleep(time.Second)
			redisConn = getRedisConn()
			continue
		}

		// 队列为空，等候100ms
		if data == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		// 发送push
		go func() {
			defer util.RecoverPanic()
			if err := pushService.Send(data); err != nil {
				core.Logger.Error(err.Error())
			}
		}()
	}
}
