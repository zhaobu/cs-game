package robot

import (
	"fmt"

	"github.com/fwhappy/util"
	"mahjong.go/library/core"
)

func (this *Robot) log(format string, a ...interface{}) {
	fmt.Println("[", util.GetTimestamp(), "]", fmt.Sprintf(format, a...))
}

func (this *Robot) trace(format string, a ...interface{}) {
	if core.RobotCfg.LogLevel < 0 {
		this.log(format, a...)
	}
}

func (this *Robot) debug(format string, a ...interface{}) {
	if core.RobotCfg.LogLevel < 1 {
		this.log(format, a...)
	}
}

func (this *Robot) show(format string, a ...interface{}) {
	if core.RobotCfg.LogLevel < 2 {
		this.log(format, a...)
	}
}
