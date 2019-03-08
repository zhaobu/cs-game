package robot

import (
	"fmt"
	"mahjong-league/core"
	"time"

	"github.com/fwhappy/util"
)

// JoinSleep 机器人加入时间间隔
func JoinSleep(interval int) {
	s := 0
	// 基本时间
	base := core.RobotCfg.JoinIntervalBase * 1000
	// 随机时间
	rand := util.RandIntn(core.RobotCfg.JoinIntervalRand*1000 + 1)
	// 根据时间段偏移时间
	hour := core.RobotCfg.JoinInterverTimes[time.Now().Hour()] * 1000
	// 固定偏移

	s = base + rand + hour + interval*1000
	fmt.Printf("[%v] [JoinSleep]base:%v,rand:%v,hour:%v,interval:%v,s:%v\n", util.GetTimestamp(), base, rand, hour, interval, s)
	time.Sleep(time.Duration(s) * time.Millisecond)
}
