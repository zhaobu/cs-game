package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"strings"

	"github.com/fwhappy/util"
	"mahjong.go/controller/game"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/router"
	gameService "mahjong.go/service/game"
	hallService "mahjong.go/service/hall"
)

// 初始化参数
var (
	// 监听ip
	listenHost = flag.String("h", "0.0.0.0", "host")
	// 监听端口
	listenPort = flag.String("p", "9090", "port")
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 配置文件夹
	confDir = flag.String("confDir", "conf", "config dir path")
	// 服务器版本
	version = flag.String("v", "", "server version")
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	defer util.RecoverPanic()

	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU() * 5)

	// 开启性能监控
	go func() {
		intPort, _ := strconv.Atoi(*listenPort)
		http.ListenAndServe(fmt.Sprintf(":%v", 10000+intPort%10), nil)
	}()

	// 初始化基础配置
	core.LoadAppConfig(core.GetConfigFile("app.toml", *env, *confDir))
	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, *confDir))
	defer core.Logger.Flush()
	// 初始化DB配置文件
	core.LoadDBConfig(core.GetConfigFile("database.toml", *env, *confDir))
	// 初始化orm配置
	core.LoadOrmConfig()
	// 初始化Redis配置
	core.LoadRedisConfig(core.GetConfigFile("redis.toml", *env, *confDir))
	// 初始化屏蔽字库
	core.InitFilter(core.GetSharedConfigFile("dict.txt", *confDir))

	// 开启监听端口 & 提供服务
	listenAndServe(*listenHost, *listenPort)
}

// listenAndServe 开始监听并提供服务
func listenAndServe(listenHost string, listenPort string) {
	remote := listenHost + ":" + listenPort
	tcpAddr, resolveErr := net.ResolveTCPAddr("tcp", ":"+listenPort)
	if resolveErr != nil {
		core.Logger.Critical("listenAndServe ResolveTCPAddr Error:%s", resolveErr.Error())
		return
	}
	tcpListener, listenErr := net.ListenTCP("tcp", tcpAddr)
	if listenErr != nil {
		core.Logger.Critical("listenAndServe ListenTCP Error:%s", listenErr.Error())
		return
	}
	core.Logger.Info("server lisen at: " + remote)

	// 记录gameRemoteAddr
	// 后续的程序，有可能用到了remote，这个逻辑必须放到最前面
	gameService.SetRemoteAddr(remote)
	// 记录服务器版本
	gameService.SetGameVersion(*version)
	// 清除大厅中的历史房间
	hallService.CleanHallRoom(remote)
	// 清除大厅中的机器人房间
	// 机器人逻辑通用了，不能再清除了
	// hallService.CleanHallRobotRoom(remote)
	// 开始记录端口的最后活动时间
	go gameService.StatLastActionTime()
	// 开启俱乐部的连接
	go gameService.CPool.Run()
	// 开启联赛服务器
	go gameService.LPool.Run()

	// 监听连接事件
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			core.Logger.Error("tcpListener.AcceptTCP: %s.", err)
			continue
		}

		if gameService.HeartBeatLogFlag {
			core.Logger.Debug("user connected, remote:%v", tcpConn.RemoteAddr().String())
		}

		// 客户端连接成功，开启新的协程，监听客户端消息
		go serve(tcpConn)
	}
}

// serve 监听客户端连接
func serve(conn *net.TCPConn) {
	// core.Logger.Debugf("New User connected: %s, server:%v", conn.RemoteAddr().String(), gameService.GetRemoteAddr())
	// 异常退出后，有一些后续的处理，还是有可能会panic，所以这里多一层revoer，防止程序crash
	// 记录当前连接的用户id
	var userId int
	// 当前用户连接成功或出错退出的消息
	c := make(chan int, 2)
	defer util.RecoverPanic()
	// 定义异常捕捉
	defer func() {
		// 捕获异常
		if err := recover(); err != nil {
			core.Logger.Critical("serve defer recover error: %s.", err)

			stack := make([]byte, 1024)
			stack = stack[:runtime.Stack(stack, true)]
			core.Logger.Critical("stack:\n%s", string(stack))
			core.Logger.Debugf("defer disconnected: %s.", conn.RemoteAddr().String())
		}

		// 断开用户连接
		if userId > 0 {
			//fixme 这里直接调用Action，有点逻辑混乱
			game.KickAction(userId)
		} else {
			conn.Close()
			c <- -1
		}
	}()

	// 检测用户连接之后，如果在规定时间内handshake成功，需要断开连接，防止无效的连接
	go gameService.ListenHandShakeSuccess(conn, c)

	// 解析消息
	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(conn)

		// 检查解析错误
		if err != nil {
			switch err {
			case io.EOF:
				// 关闭退出
				// 对方关闭的socket上执行read操作会得到EOF error，但write操作会成功
				core.Logger.Debugf("User disconnected, remote: %s.", conn.RemoteAddr().String())
			case io.ErrUnexpectedEOF:
				core.Logger.Debugf("unexpected EOF, remote: %s.", conn.RemoteAddr().String())
			default:
				// 己方已经关闭的socket上再进行read和write操作，会得到”use of closed network connection” error；
				if strings.Contains(err.Error(), "use of closed network connection") {
					core.Logger.Debug("连接关闭, remote:%v", conn.RemoteAddr().String())
				} else if strings.Contains(err.Error(), "connection reset by peer") {
					// do nothing
				} else {
					// 协议解析错误
					core.Logger.Error("%v", err.Error())
				}
			}
			break
		}
		router.Dispatch(&userId, impacket, conn, c)
	}
}
