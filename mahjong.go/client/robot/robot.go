package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"

	"mahjong.go/robot"
	hallService "mahjong.go/service/hall"
)

// 支持的命令行参数
var (
	// 服务器ip、端口、环境
	env = flag.String("env", "local", "env")
	// 开始的用户id
	fromUserId = flag.Int("f", 50000, "初始用户id")
	// 总用户数
	userCount = flag.Int("c", 1000, "总用户数")

	// 机器人回应解散操作的响应间隔
	dismissInterval = flag.Int("di", 1, "解散回应操作间隔")
	dismissRandom   = flag.Int("dr", 1, "解散回应操作随机间隔")
)

// 设置
var (
	users    map[int]int // 用户列表
	userLock *sync.Mutex // 用户列表锁
)

func init() {
	flag.Parse()
	users = make(map[int]int)
	userLock = &sync.Mutex{}
}

func main() {
	go func() {
		http.ListenAndServe(":7776", nil)
	}()
	fmt.Println("开启机器人")

	// 初始化机器人配置
	core.LoadRobotConfig(fmt.Sprintf("conf/env/%s/robot.toml", *env))
	// 初始化Redis配置
	core.LoadRedisConfig(fmt.Sprintf("conf/env/%s/redis.toml", *env))
	// 启动时，清空原来的机器人申请
	hallService.CleanHallRobotRoom("")

	// 从redis中读取一个游戏配置
	for {
		if newGame := hallService.GetHallRobotRoom(""); len(newGame) > 0 {
			// 每次都去判断一下测试文件
			if !core.IsProduct() {
				loadTestConfig()
			}
			gameInfo := robot.DeserializeGameInfo(newGame)
			debug("新的机器人游戏配置, gameInfo:%+v", gameInfo)
			if gameInfo.RoomId > 0 {
				go startRoom(gameInfo)
			} else if gameInfo.RaceId > 0 && gameInfo.LeagueId > 0 {
				go startLeague(gameInfo)
			} else if gameInfo.RobotId > 0 {
				go startRobot(gameInfo)
			} else {
				show("未支持的游戏配置, gameInfo:%+v", gameInfo)
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// 单独启动一个机器人
func startRobot(gameInfo *robot.GameInfo) {
	createRobot(gameInfo)
}

// 开始给房间添加机器人
func startRoom(gameInfo *robot.GameInfo) {
	show(fmt.Sprintf("发现一个新房间:%+v", gameInfo))
	// 开3个机器人给这个用户
	for i := 1; i < gameInfo.RequireCnt; i++ {
		robot.JoinSleep(gameInfo.GradeId)
		go createRobot(gameInfo)
	}
}

// 开始联赛添加机器人
func startLeague(gameInfo *robot.GameInfo) {
	show(fmt.Sprintf("发现一个新的联赛:%+v", gameInfo))
	// 开3个机器人给这个用户
	for i := 1; i < gameInfo.RequireCnt; i++ {
		robot.JoinSleep(gameInfo.GradeId)
		go createRobot(gameInfo)
	}
}

// 创建机器人
func createRobot(gameInfo *robot.GameInfo) {
	var robotId int
	if gameInfo.RobotId > 0 {
		robotId = gameInfo.RobotId
	} else {
		robotId = getUserId()
		if robotId == 0 {
			show("userId已用尽")
			return
		}
	}
	defer func() {
		delUser(robotId)
	}()

	// 设置机器人已被占用
	addUser(robotId)
	user := robot.NewRobot(robotId)
	user.RoomId = gameInfo.RoomId
	user.CType = gameInfo.CType
	user.MType = gameInfo.GType
	user.DismissInterval = *dismissInterval
	user.DismissRandom = *dismissRandom
	user.GameInfo = gameInfo
	user.TRound = gameInfo.TRound
	user.AILevel = gameInfo.AILevel
	user.Run()
}

func addUser(userId int) {
	userLock.Lock()
	defer userLock.Unlock()

	users[userId] = 1
}

func delUser(userId int) {
	userLock.Lock()
	defer userLock.Unlock()

	delete(users, userId)
}

// 获取一个可用的用户
// 策略是先随机10次，如果都找不到可以用的，那么就从头开始找，如果再找不到，则机器人已用完
func getUserId() int {
	userLock.Lock()
	defer userLock.Unlock()

	for i := 0; i < 20; i++ {
		id := *fromUserId + util.RandIntn(*userCount)
		if _, exists := users[id]; !exists {
			return id
		}
	}

	for incrementId := 0; incrementId < *userCount; incrementId++ {
		id := *fromUserId + incrementId
		if _, exists := users[id]; !exists {
			return id
		}
	}
	return 0
}

// 加载测试文件
func loadTestConfig() {
	_path := fmt.Sprintf("conf/%s/robot.toml", *env)
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		show("测试文件未找到 path:%v", _path)
		return
	}
	show("加载机器人配置完成, config:%#v", core.RobotCfg)
}

// 输出日志
func log(format string, a ...interface{}) {
	fmt.Println("["+util.GetTimestamp()+"] ", fmt.Sprintf(format, a...))
}
func debug(format string, a ...interface{}) {
	log(format, a...)
}
func show(format string, a ...interface{}) {
	log(format, a...)
}
