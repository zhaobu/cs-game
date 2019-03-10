package main

import (
	"flag"
	"fmt"
	"mahjong-sensitive/core"
	"net/http"
	"strconv"
	"time"

	"github.com/fwhappy/util"
)

var (
	// 监听端口
	host = flag.String("host", "0.0.0.0", "host")
	// 监听端口
	port = flag.Int("port", 53782, "port")
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 配置文件夹
	confDir = flag.String("confDir", "conf", "config dir path")
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	defer util.RecoverPanic()

	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, *confDir))
	defer core.Logger.Flush()
	// 初始化屏蔽字库
	core.InitFilter(core.GetConfigFile("dict.txt", "", *confDir))

	http.HandleFunc("/test", hello)
	// 选服服务
	http.HandleFunc("/replace", sensitiveReplace)
	http.HandleFunc("/check", sensitiveCheck)
	core.Logger.Info("start listen:%v:%v", *host, *port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", *host, *port), nil)
	if err != nil {
		core.Logger.Errorf("ListenAndServe:%v", err)
		return
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sleepTime, _ := strconv.Atoi(r.Form.Get("sleep"))
	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
	w.Write([]byte("Hello world"))
}

// 屏蔽字过滤
func sensitiveReplace(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var content string
	// 优先解析post参数
	content = r.PostFormValue("s")
	if content == "" {
		content = r.Form.Get("s")
	}
	if content == "" {
		core.Logger.Warn("[sensitiveReplace]参数[s]为空, from:%v", r.RemoteAddr)
		w.Write([]byte{})
		return
	}
	core.Logger.Debug("[sensitiveReplace]原字符串,:%v, from:%v", content, r.RemoteAddr)
	replacedContent, err := core.FilterManager.Filter().Replace(content, '*')
	if err != nil {
		core.Logger.Warn("[sensitiveReplace]敏感词过滤失败, content:%v, err:%v", content, err.Error())
	} else {
		core.Logger.Debug("[sensitiveReplace]content:%v, replacedContent:%v", content, replacedContent)
		content = replacedContent
	}

	w.Write([]byte(content))
}

// 检测是否包含屏蔽字
func sensitiveCheck(w http.ResponseWriter, r *http.Request) {
	var content string
	// 优先解析post参数
	content = r.PostFormValue("s")
	if content == "" {
		content = r.Form.Get("s")
	}
	if content == "" {
		core.Logger.Warn("[sensitiveCheck]参数[s]为空, from:%v", r.RemoteAddr)
		w.Write([]byte("0"))
		return
	}
	core.Logger.Debug("[sensitiveCheck]原字符串,:%v, from:%v", content, r.RemoteAddr)
	result, err := core.FilterManager.Filter().Filter(content)
	if err != nil {
		core.Logger.Warn("[sensitiveCheck]敏感词检测失败, content:%v, err:%v", content, err.Error())
	} else {
		core.Logger.Debug("[sensitiveCheck]content:%v, result:%v", content, result)
		if len(result) > 0 {
			w.Write([]byte("1"))
		} else {
			w.Write([]byte("0"))
		}
	}
}
