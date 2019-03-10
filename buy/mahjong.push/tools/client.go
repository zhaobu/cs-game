package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"mahjong.push/config"
	"mahjong.push/library/core"
	"mahjong.push/library/util"
)

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "qa", "env")
)

func init() {
	// 解析url参数
	flag.Parse()
}

// 测试写入函数
func main() {
	defer util.RecoverPanic()

	// 初始化Redis配置
	core.LoadRedisConfig(fmt.Sprintf("conf/%s/redis.toml", *env))

	// testMI()
	testAPNS()
}

func testMI() {
	for {
		data := make(map[string]interface{})
		// 发送者信息
		data["senderId"] = 1
		data["senderNickname"] = "王虎"
		// 接受者信息
		data["device"] = "android"
		data["deviceToken"] = "Bp4y/7iVSGXoGYblk8rBwU7FnqCPt9AokrR2whtgbMk="
		// 语言包id
		data["langId"] = util.RandIntn(32) + 1
		// 发送时间
		data["time"] = util.GetTime()

		var byt, _ = json.Marshal(data)
		core.RedisDo(core.RedisClient1, core.RedisDBPUSH, "lpush", config.CACHE_KEY_PUSH_QUEUE_LIST, string(byt))
		fmt.Printf("test push:%v\n", data)

		time.Sleep(time.Second * 5)
	}
}

func testAPNS() {
	for {
		data := make(map[string]interface{})
		// 发送者信息
		data["senderId"] = 1
		data["senderNickname"] = "王虎"
		// 接受者信息
		data["device"] = "ios"
		data["deviceToken"] = "pjfU4WTj3+0sH7L4aTKx6aLoFsWpIFZo5B1vj0or9W8="
		data["deviceToken"] = "vVLh5906kzBnCSHsv//J829aeX76KRLit4QFLrMoF4w="
		// 语言包id
		data["langId"] = util.RandIntn(32) + 1
		// 发送时间
		data["time"] = util.GetTime()

		var byt, _ = json.Marshal(data)
		core.RedisDo(core.RedisClient1, core.RedisDBPUSH, "lpush", config.CACHE_KEY_PUSH_QUEUE_LIST, string(byt))
		fmt.Printf("test push:%v\n", data)

		time.Sleep(time.Second * 5)
	}
}
