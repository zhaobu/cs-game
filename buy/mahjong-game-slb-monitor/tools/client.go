package main

import (
	"flag"
	"fmt"
	"mahjong-game-slb-monitor/core"

	"github.com/fwhappy/util"
	"github.com/garyburd/redigo/redis"
)

var (
	// 配置文件夹, 最好是绝对路劲, /Home/xh/goim/etc/local
	confDir = flag.String("confDir", "etc/local", "config dir path")
)

func init() {
	flag.Parse()
}

func main() {
	// 初始化Redis配置
	core.LoadRedisConfig(core.GetConfigFile("redis.toml", *confDir))

	servers := make(map[string]interface{})

	ip := "127.0.0.1"
	slb := []string{
		"114.55.227.47",   // qa
		"118.178.190.132", // g1
		"121.43.37.217",   // g2
		"118.178.127.24",  // g3
		"121.199.9.247",   // g4
		// "101.37.227.73",   // g4-slb
		// "172.217.27.142",  // google
		// "216.239.37.1",    // golang.org
		// "74.117.182.168",  // facebook
		// "31.13.84.4", // twitter

	}
	server := make(map[string]interface{})
	server["name"] = "g1"
	server["ip"] = ip
	server["ports"] = []string{"8990", "8991"}
	server["slbs"] = slb
	servers[ip], _ = util.InterfaceToJsonString(server)

	fmt.Println(fmt.Sprintf("server:%#v", servers))

	// 存储至redis
	// data, _ := util.InterfaceToJsonString(servers)
	core.RedisDo("del", "SERVER:LISTS")
	_, err := core.RedisDo("hmset", redis.Args{}.Add("SERVER:LISTS").AddFlat(servers)...)
	if err != nil {
		fmt.Println("存储servers失败", err.Error())
	} else {
		fmt.Println("存储servers成功")
	}
}
