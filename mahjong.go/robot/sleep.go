package robot

import (
	"fmt"
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
)

// 机器人准备时间间隔
func (this *Robot) prepareSleep() {
	s := core.RobotCfg.PrepareInterval[0] + util.RandIntn(core.RobotCfg.PrepareInterval[1]-core.RobotCfg.PrepareInterval[0]+1)
	time.Sleep(time.Duration(s) * time.Millisecond)
}

// 机器人胡牌时间间隔
func (this *Robot) winSleep() {
	time.Sleep(time.Duration(core.RobotCfg.WinInterval) * time.Millisecond)
}

// 机器人定缺时间间隔
func (this *Robot) lackSleep() {
	s := core.RobotCfg.LackInterval[0] + util.RandIntn(core.RobotCfg.LackInterval[1]-core.RobotCfg.LackInterval[0]+1)
	time.Sleep(time.Duration(s) * time.Millisecond)
}

// 机器人决策时间间隔
func (this *Robot) replySleep() {
	v := util.RandIntn(100) + 1
	for _, cfg := range core.RobotCfg.ReplyInterval {
		if v > cfg[0] {
			// 未命中
			v -= cfg[0]
		} else {
			// 命中
			s := cfg[1] + util.RandIntn(cfg[2]-cfg[1]+1)
			time.Sleep(time.Duration(s) * time.Millisecond)
			break
		}
	}
}

// 机器人加入时间间隔
func JoinSleep(gradeId int) {
	s := 0
	// 基本时间
	base := core.RobotCfg.JoinIntervalBase * 1000
	// 随机时间
	rand := util.RandIntn(core.RobotCfg.JoinIntervalRand*1000 + 1)
	// 根据时间段偏移时间
	hour := core.RobotCfg.JoinInterverTimes[time.Now().Hour()] * 1000
	// 排位赛根据段位便宜\
	var grade int
	if gradeId > 0 {
		grade = core.RobotCfg.JoinInterverGrades[gradeId-1] * 1000
	}

	s = base + rand + hour + grade
	fmt.Printf("[%v] [JoinSleep]base:%v,rand:%v,hour:%v,grade:%v,s:%v\n", util.GetTimestamp(), base, rand, hour, grade, s)
	time.Sleep(time.Duration(s) * time.Millisecond)
}
