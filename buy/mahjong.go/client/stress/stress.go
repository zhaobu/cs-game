package main

import (
	"flag"
	"fmt"
	"time"

	"mahjong.go/library/core"
	"mahjong.go/robot"
)

// 命令行参数
var (
	// 服务器ip、监听端口、环境
	h   = flag.String("h", "127.0.0.1", "host")
	p   = flag.String("p", "9090", "port")
	env = flag.String("env", "local", "env")
	// h = flag.String("h", "10.26.239.72", "host")
	// p = flag.String("p", "8990", "port")
	// env = flag.String("env", "qa", "env")
	// 参与测试牌用户数
	cnt = flag.Int("c", 4, "参与测试用户数")
	// 开始的用户id
	fromUserId = flag.Int("f", 1, "初始用户id")
)

// 设置
var (
	debugMode      = true                     // 是否输出日志
	createInterval = time.Millisecond * 60000 // 每分1个用户
	replyInterval  = 1                        // 回应间隔 replyInterval
	remote         string                     // 远程连接
)

func init() {
	flag.Parse()
	remote = *h + ":" + *p
}

func main() {
	fmt.Printf("压测程序启动, remote:%v, from user id :%v, total user:%v \n", remote, fromUserId, cnt)
	// 初始化机器人配置
	core.LoadRobotConfig(fmt.Sprintf("conf/env/%s/robot.toml", *env))
	userId := *fromUserId
	totalUser := *cnt
	for i := 0; i < totalUser; i++ {
		gameInfo := robot.NewGameInfo(remote, 0, 5, 1, 1, 4)
		gameInfo.AILevel = 2 // 高级ai

		fmt.Printf("启动机器人, robot id :%v, progress:%v/%v \n", userId, userId, *cnt)
		go createRobot(userId, gameInfo)
		userId++

		time.Sleep(createInterval)
	}

	// 挂起主进程
	time.Sleep(time.Hour * 24)
}

// 创建机器人
func createRobot(userId int, gameInfo *robot.GameInfo) {
	user := robot.NewRobot(userId)
	user.RoomId = gameInfo.RoomId
	user.CType = gameInfo.CType
	user.MType = gameInfo.GType
	user.DismissInterval = 1
	user.DismissRandom = 1
	user.GameInfo = gameInfo
	user.TRound = gameInfo.TRound
	user.AILevel = gameInfo.AILevel
	user.Run()
}
