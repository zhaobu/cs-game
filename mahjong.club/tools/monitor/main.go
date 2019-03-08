package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fwhappy/util"
	"mahjong.club/core"
)

var (
	// 服务器版本
	version = flag.String("v", "2.0.0", "server version")
	// 监控端口
	port = flag.String("p", "38438", "port")
	// 服务器环境
	env = flag.String("env", "local", "env")
	// nohup目录
	nohupDir = "/data/mahjong.nohup"
	// 项目名称
	project = "mahjong.club"
	// 检测间隔
	loopInterval = time.Duration(1) * time.Second
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	// 初始化基础配置
	core.LoadAppConfig(core.GetConfigFile("app.toml", *env, "etc"))

	cCommand := getCheckCommand()
	sCommand := getStartCommand()
	fmt.Println("开启监听进程, checkCommand:", cCommand, ", startCommand:", sCommand)
	for {
		pids := getPids(cCommand)
		if len(pids) == 0 {
			fmt.Println("未检测到运行中的进程，启动一条")
			startCommand(sCommand)
			body := ""
			body += "俱乐部报警：<br />"
			body += "类型：恢复<br />"
			body += "时间:" + util.GetTimestamp() + "<br />"
			sendAlarm(body)
		}
		time.Sleep(loopInterval)
	}
}

func getCheckCommand() string {
	return fmt.Sprintf("ps aux | grep 'bin/%v.%v' | grep -v 'grep'", project, *version)
}

func getStartCommand() string {
	command := fmt.Sprintf("../../bin/%v.%v -env=%v -port=%v >> %v/%v.nohup.out.%v 2>&1 &",
		project, *version, *env, *port, nohupDir, project, *version)
	return command
}

// 获取进程的pids
func getPids(command string) []int {
	pids := make([]int, 0)
	cmd := exec.Command("/bin/sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return pids
	}
	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")
		ft := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				ft = append(ft, t)
			}
		}
		pid, err := strconv.Atoi(ft[1])
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}

// 执行一个shell命令，不关心返回
func startCommand(command string) {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Start()
}

func sendAlarm(body string) {
	content := make(map[string]interface{})
	content["target"] = "default"
	content["subject"] = core.AppConfig["alarm_subject"].(string)
	content["body"] = body
	data, _ := util.InterfaceToJsonString(content)
	resp, err := http.Post(core.AppConfig["alarm_host"].(string),
		"application/x-www-form-urlencoded", strings.NewReader("*="+data))
	if err != nil {
		fmt.Println("[Send Mail error]error:", err)
		return
	}
	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[Send Mail error]error:", err)
		return
	}
	fmt.Println("[sendAlarm]response:", string(response))
}
