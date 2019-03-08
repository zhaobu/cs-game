package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego/logs"
	"github.com/fwhappy/util"
)

// LogConfig 配置
type LogConfig struct {
	Log_adapter_console        bool
	Log_console_level          int
	Log_file                   string
	Log_file_level             int
	Log_enable_func_call_depth bool
	Log_async                  bool
	Log_chan_length            int
	Log_maxlines               int
	Log_maxsize                int
	Log_daily                  bool
	Log_maxdays                int
	Log_rotate                 bool
	Log_multifile              bool
	Log_separate               []string
}

type MLog struct {
	*logs.BeeLogger
}

var Logger *MLog

// LoadLoggerConfig 加载日志配置
func LoadLoggerConfig(cfgFile string) {
	var logConfig LogConfig

	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		panic(err)
	}
	if _, err := toml.Decode(string(content), &logConfig); err != nil {
		panic(err)
	}

	// 加载log
	log := logs.NewLogger()
	// 设置异步输出
	if logConfig.Log_async {
		log.Async(int64(logConfig.Log_chan_length))
	}
	// 设置输出文件名、文件行数
	if logConfig.Log_enable_func_call_depth {
		log.EnableFuncCallDepth(true)
	}
	// 设置控制台输出
	if logConfig.Log_adapter_console {
		consoleConfig := make(map[string]int)
		consoleConfig["level"] = logConfig.Log_console_level
		byt, _ := json.Marshal(consoleConfig)
		log.SetLogger(logs.AdapterConsole, string(byt))
	}

	fileConfig := make(map[string]interface{})
	fileConfig["filename"] = logConfig.Log_file
	fileConfig["maxlines"] = logConfig.Log_maxlines
	fileConfig["maxsize"] = logConfig.Log_maxsize
	fileConfig["daily"] = logConfig.Log_daily
	fileConfig["maxdays"] = logConfig.Log_maxdays
	fileConfig["rotate"] = logConfig.Log_rotate
	// fileConfig["level"] = logConfig.Log_file_level
	if logConfig.Log_multifile {
		fileConfig["separate"] = logConfig.Log_separate
		byt, _ := json.Marshal(fileConfig)
		log.SetLogger(logs.AdapterMultiFile, string(byt))
	} else {
		byt, _ := json.Marshal(fileConfig)
		log.SetLogger(logs.AdapterFile, string(byt))
	}

	// 据说不这样做，会有一些性能问题
	log.SetLevel(logConfig.Log_file_level)

	Logger = &MLog{log}
}

// Trace 直接输出到stdout
func (ml *MLog) Trace(format string, v ...interface{}) {
	fmt.Printf(util.GetTimestamp()+" [T] "+format+"\n", v...)
}
