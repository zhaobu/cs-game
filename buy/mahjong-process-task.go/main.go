package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/fwhappy/util"
	"mahjong-process-task.go/config"
	"mahjong-process-task.go/core"
	"mahjong-process-task.go/ierror"
)

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 内网地址
	host = flag.String("h", "127.0.0.1", "local host")
	// 监听端口
	port = flag.Int("p", 12121, "port")
)

func init() {
	flag.Parse()
}

func main() {
	defer util.RecoverPanic()

	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, "conf"))
	defer core.Logger.Flush()

	http.HandleFunc("/test", hello)
	http.HandleFunc("/start", start)
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/install", install)
	http.HandleFunc("/checkInstall", checkInstall)
	core.Logger.Info("start listen:%v:%v", *host, *port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", *host, *port), nil)
	if err != nil {
		core.Logger.Errorf("ListenAndServe:%v", err)
		return
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

// 启动一个服务
func start(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 要启动的端口
	port, _ := strconv.Atoi(r.Form.Get("port"))
	// 检查端口范围
	if !config.VerifyServerPort(*host, port) {
		w.Write(buildError(ierror.NewError(-4, *host, port)))
		return
	}
	// 要启动的版本
	version := r.Form.Get("version")
	// 检查版本是否已经存在
	if !verfiryVersionIsInstalled(version) {
		w.Write(buildError(ierror.NewError(-5, *host, version)))
		return
	}

	// 生成并执行启动命令，返回执行结果
	command := fmt.Sprintf("%v start %v %v >> /data/mahjong-process-task/command.log", config.GetDeployFile(), version, getServer(port))
	core.Logger.Debug("[start]build command:%v", command)
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		core.Logger.Info("[start]fail, server:%v, port:%v, version:%v", *host, port, version)
		w.Write(buildError(ierror.NewError(-6, *host, command)))
		return
	}
	core.Logger.Info("[start]success, server:%v, port:%v, version:%v", *host, port, version)
	w.Write(buildSuccess(""))
}

// 关闭一个端口
func stop(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 要关闭的端口
	port, _ := strconv.Atoi(r.Form.Get("port"))
	// 检查端口范围
	if !config.VerifyServerPort(*host, port) {
		w.Write(buildError(ierror.NewError(-4, *host, port)))
		return
	}
	// 生成并执行关闭命令，返回执行结果
	command := fmt.Sprintf("%v stop %v >> /data/mahjong-process-task/command.log", config.GetDeployFile(), getServer(port))
	core.Logger.Debug("[stop]command:%v", command)
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		core.Logger.Info("[stop]fail, server:%v, port:%v", *host, port)
		w.Write(buildError(ierror.NewError(-6, *host, command)))
		return
	}
	core.Logger.Info("[stop]success, server:%v, port:%v", *host, port)
	w.Write(buildSuccess(""))
}

// 安装一个版本
func install(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 要关闭的端口
	tag := r.Form.Get("tag") // tag号或者是分支
	if tag == "" {
		w.Write(buildError(ierror.NewError(-2, "install", "tag")))
		return
	}
	version := r.Form.Get("version") // 编译目标版本号
	if version == "" {
		w.Write(buildError(ierror.NewError(-2, "install", "version")))
		return
	}

	// 生成安装命令并执行，因为持续时间较长，不等待执行结果
	command := fmt.Sprintf("nohup %v inpull %v %v >> /data/mahjong-process-task/command.log &", config.GetDeployFile(), tag, version)
	core.Logger.Debug("[install]command:%v", command)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Run()
	// err := cmd.Run()
	// if err != nil {
	// 	core.Logger.Info("[install]fail, server:%v, tag:%v, version:%v", *host, tag, version)
	// 	w.Write(buildError(ierror.NewError(-6, *host, command)))
	// 	return
	// }
	core.Logger.Info("[install]success, server:%v, tag:%v, version:%v", *host, tag, version)
	w.Write(buildSuccess("安装申请发起成功，预计需要30s左右可完成"))
}

// 检查版本是否安装成功
func checkInstall(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	// 检查版本号
	version := r.Form.Get("version")
	if version == "" {
		w.Write(buildError(ierror.NewError(-2, "checkInstall", "version")))
		return
	}
	// 目标文件
	file := config.GetMainFile(version)
	info, err := os.Stat(file)
	if err == nil || os.IsExist(err) {
		w.Write(buildSuccess(fmt.Sprintf("版本已安装, server:%v, version:%v, 安装时间:%v", *host, version, info.ModTime().Format("2006-01-02 15:04:05"))))
		return
	}
	w.Write(buildError(ierror.NewError(-7, *host, version)))
}

// 拼接服务器名称
func getServer(port int) string {
	core.Logger.Debug("host:%v", config.GetServerNameByIP(*host))
	return fmt.Sprintf("%v_%v", config.GetServerNameByIP(*host), port)
}

// 检查版本对应的编译文件是否存在
func verfiryVersionIsInstalled(version string) bool {
	file := config.GetMainFile(version)
	fmt.Println("file:", file)
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}
