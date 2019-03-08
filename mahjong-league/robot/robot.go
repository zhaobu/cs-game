package robot

import (
	"mahjong-league/core"
	"sync"

	"github.com/fwhappy/util"
)

const MIN_ROBOT_ID = 90000
const ROBOT_COUNT = 10000

const FILL_ROBOT_FROM = 2000
const FILL_ROBOT_COUNT = 8000

// PRobots 占用中的机器人
type PRobots struct {
	mux    *sync.RWMutex
	robots map[int]int
}

// 占用中的机器人
var playingRobots *PRobots

func init() {
	playingRobots = &PRobots{
		mux:    &sync.RWMutex{},
		robots: make(map[int]int, 0),
	}
}

// IsRobot 判断某个id是否机器人
func IsRobot(robotId int) bool {
	return robotId >= MIN_ROBOT_ID && robotId < (MIN_ROBOT_ID+ROBOT_COUNT)
}

// 是否是系统机器人用户
func IsRobotUser(robotId int) bool {
	if robotId >= 2000 && robotId <= 10000 {
		return true
	}
	if robotId >= 50000 && robotId < 100000 {
		return true
	}
	return false
}

// FetchFill 取一个补位的机器人id
func FetchFill() int {
	return FILL_ROBOT_FROM + util.RandIntn(FILL_ROBOT_COUNT+1)
}

// Fetch 取n个可用的机器人
func Fetch(n int) []int {
	playingRobots.mux.Lock()
	defer playingRobots.mux.Unlock()

	robots := make([]int, 0)
	for i := 0; i < n; i++ {
		robotId := getValid()
		robots = append(robots, robotId)
		playingRobots.robots[robotId] = robotId
	}
	core.Logger.Info("[robot.Remove]robots:%v", robots)
	return robots
}

// Occupied 占用
func Occupied(robotId int) {
	playingRobots.mux.Lock()
	defer playingRobots.mux.Unlock()
	playingRobots.robots[robotId] = robotId

	core.Logger.Info("[robot.Occupied]robotId:%v", robotId)
}

// Remove 取消机器人的占位
func Remove(robotId int) {
	playingRobots.mux.Lock()
	defer playingRobots.mux.Unlock()
	delete(playingRobots.robots, robotId)

	core.Logger.Info("[robot.Remove]robotId:%v", robotId)
}

// 取一个可用的机器人
func getValid() int {
	var id int
	for i := 0; i < 20; i++ {
		id = MIN_ROBOT_ID + util.RandIntn(ROBOT_COUNT)
		if _, exists := playingRobots.robots[id]; !exists {
			return id
		}
	}
	return MIN_ROBOT_ID
}
