package majiang

import (
	zaplog "game/common/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RoomLog struct {
	Tlog    *zap.Logger
	Log     *zap.SugaredLogger //printf风格
	LogName string
}

func (self *RoomLog) InitLog(gameName string, deskId uint64, release *bool) {
	var logLevel string
	self.LogName = fmt.Sprintf("./log/roomlog/%s_%d_%s.log", gameName, deskId, time.Now().Format("01021504"))
	if *release {
		logLevel = "info"
	} else {
		logLevel = "debug"
	}
	self.Tlog = zaplog.InitLogger(self.LogName, logLevel, !*release)
	self.Log = self.Tlog.Sugar()
}
