package zaplog

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(logpath string, loglevel string, debugmode bool) *zap.Logger {

	var allCore []zapcore.Core

	hook := lumberjack.Logger{
		Filename:   logpath, //日志文件路径
		MaxSize:    2048,    // megabytes
		MaxBackups: 30,      //最多保留备份个数
		MaxAge:     7,       //days
		Compress:   true,    //是否压缩 disabled by default
	}
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	fileWriter := zapcore.AddSync(&hook)

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)

	// for human operators.
	var encoderConfig zapcore.EncoderConfig
	timeFormat := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
	}
	//如果是debug模式,同时输出到到终端
	if debugmode {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = timeFormat
		allCore = append(allCore, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), consoleDebugging, level))
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = timeFormat
	}
	allCore = append(allCore, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, level))

	core := zapcore.NewTee(allCore...)

	// From a zapcore.Core, it's easy to construct a Logger.
	return zap.New(core).WithOptions(zap.AddCaller())
}
