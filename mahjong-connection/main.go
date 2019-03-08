package main

import (
	"flag"
	"fmt"
	"io"
	"mahjong-connection/core"
	"mahjong-connection/hall"
	"mahjong-connection/protocal"
	"mahjong-connection/router"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
	"time"

	"github.com/fwhappy/util"
)

var (
	host       = flag.String("h", "0.0.0.0", "listen host")
	socketPort = flag.Int("sp", 9000, "socket port")
	// webSocketPort = flag.Int("wp", 19000, "websocket port")
	httpPort = flag.Int("hp", 29000, "http port")
	env      = flag.String("env", "local", "env")
	version  = flag.String("v", "latest", "version")
	confDir  = flag.String("confDir", "conf", "config dir path")
)

var (
	// 服务器状态
	socketStatus bool
	httpStatus   bool
)

func init() {
	// 解析参数
	flag.Parse()

	// 初始化参数
	socketStatus = false
	httpStatus = false
}

// 地道贵州麻将项目连接层服务
// 提供http、websocket以及socket三种方式的服务
func main() {
	defer util.RecoverPanic()

	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU() * 5)

	// 开启性能监控
	go func() {
		fmt.Println("开启性能监控, port: ", 29001)
		http.ListenAndServe(fmt.Sprintf(":%v", 29001), nil)
	}()

	// 记录服务器的remote
	hall.RemoteAddr = fmt.Sprintf("%v:%v", *host, *socketPort)

	// 初始化配置
	// 初始化基础配置
	core.LoadAppConfig(core.GetConfigFile("app.toml", *env, *confDir))
	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, *confDir))
	defer core.Logger.Flush()
	// 初始化DB配置文件
	// 初始化orm配置
	// 初始化Redis配置

	// 开启服务
	start()

	// 防止程序异常退出，这里设置报警
	for {
		time.Sleep(time.Minute)

		// 报警，提示游戏服异常
		if !socketStatus {
			core.Logger.Critical("socket服务异常, port:%v", *socketPort)
		}
		if !httpStatus {
			core.Logger.Critical("http服务异常, port:%v", *httpPort)
		}
	}
}

func start() {
	// 开启socket服务
	go listenSocket()

	// 开启websocket服务
	// go listenWebSocket()

	// 开启http服务
	go listenHTTP()
}

// 监听socket
func listenSocket() {
	core.Logger.Info("socket服务开启, 监听: %v:%v", *host, *socketPort)
	defer util.RecoverPanic()
	defer func() {
		// 将socket状态置为异常
		socketStatus = false
		core.Logger.Critical("socket服务已退出, 监听: %v:%v", *host, *socketPort)
	}()

	tcpAddr, resolveErr := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", *host, *socketPort))
	if resolveErr != nil {
		core.Logger.Critical("[listenSocket]ResolveTCPAddr Error:%s", resolveErr.Error())
		return
	}
	tcpListener, listenErr := net.ListenTCP("tcp", tcpAddr)
	if listenErr != nil {
		core.Logger.Critical("[listenSocket]ListenTCP Error:%s", listenErr.Error())
		return
	}

	socketStatus = true
	// 监听连接事件
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			core.Logger.Error("[listenSocket]tcpListener.AcceptTCP Error: %s.", err)
			continue
		}

		// 客户端连接成功，开启新的协程，监听客户端消息
		go socketServe(conn)
	}
}

/*
// 提供websocket服务
func listenWebSocket() {
	core.Logger.Info("websocket服务开启, 监听: %v:%v", *host, *webSocketPort)
	defer func() {
		core.Logger.Critical("websocket服务已退出, 监听: %v:%v", *host, *webSocketPort)
	}()
	defer util.RecoverPanic()

	http.Handle("/ws", websocket.Handler(webSocketServe))
	if err := http.ListenAndServe(fmt.Sprintf("%v:%v", *host, *webSocketPort), nil); err != nil {
		log.Fatal("ListenAndServe web socket error:", err)
	}
}
*/
// 提供http服务
func listenHTTP() {
	core.Logger.Info("http服务开启, 监听: %v:%v", *host, *httpPort)
	defer func() {
		// 将http状态置为异常
		httpStatus = false
		core.Logger.Critical("http服务已退出, 监听: %v:%v", *host, *httpPort)
	}()

	httpStatus = true
	router.HTTPDespatch()

	err := http.ListenAndServe(fmt.Sprintf("%v:%v", *host, *httpPort), nil)
	if err != nil {
		core.Logger.Critical("ListenAndServe http error:%v", err)
		return
	}
}

// 提供socket服务
func socketServe(conn *net.TCPConn) {
	defer util.RecoverPanic()
	defer func() {
		// TODO 断开连接
		hall.KickConn(conn)
		core.Logger.Debug("disconnected: %s.", conn.RemoteAddr().String())
	}()
	core.Logger.Debug("new connected: %s.", conn.RemoteAddr().String())

	// 检测用户连接之后，如果在规定时间内handshake成功，需要断开连接，防止无效的连接
	hall.ListenHandShakeTimeout(conn)

	for {
		// 读取包内容
		impacket, err := protocal.ReadPacket(conn)
		// 检查解析错误
		if err != nil {
			switch err {
			case io.EOF:
				// 关闭退出
				// 对方关闭的socket上执行read操作会得到EOF error，但write操作会成功
				core.Logger.Trace("User disconnected, remote: %s.", conn.RemoteAddr().String())
			case io.ErrUnexpectedEOF:
				core.Logger.Trace("unexpected EOF, remote: %s.", conn.RemoteAddr().String())
			default:
				// 己方已经关闭的socket上再进行read和write操作，会得到”use of closed network connection” error；
				if strings.Contains(err.Error(), "use of closed network connection") {
					core.Logger.Trace("连接关闭, remote:%v", conn.RemoteAddr().String())
				} else if strings.Contains(err.Error(), "connection reset by peer") {
					core.Logger.Trace("连接关闭, remote:%v", conn.RemoteAddr().String())
				} else {
					// 协议解析错误
					core.Logger.Error("协议解析错误, %v", err.Error())
				}
			}
			break
		}

		// 分发客户端的数据
		router.ClientSocketDespatch(impacket, conn)
	}
}

/*
// 提供websocket服务
func webSocketServe(conn *websocket.Conn) {
	defer util.RecoverPanic()

	var err error
	for {
		var reply string
		err = websocket.Message.Receive(conn, &reply)
		if err != nil {
			switch err {
			case io.EOF:
				core.Logger.Trace("[webSocketServe]io.EOF,User disconnected, remote: %v", conn.RemoteAddr().String())
			case io.ErrUnexpectedEOF:
				core.Logger.Trace("[webSocketServe]io.ErrUnexpectedEOF,User disconnected, remote: %v", conn.RemoteAddr().String())
			default:
				// 己方已经关闭的socket上再进行read和write操作，会得到”use of closed network connection” error；
				if strings.Contains(err.Error(), "use of closed network connection") {
					core.Logger.Trace("[webSocketServe][closed network]连接已关闭, remote:%v", conn.RemoteAddr().String())
				} else if strings.Contains(err.Error(), "connection reset by peer") {
					core.Logger.Trace("[webSocketServe][reset by peer]连接已关闭, remote:%v", conn.RemoteAddr().String())
				} else {
					// 协议解析错误
					core.Logger.Error("[webSocketServe]receive error:%v", err.Error())
				}
			}
			break
		}
		core.Logger.Debug("[websocket]receive message:%v", reply)

		// 解析协议
		// 转发
		router.WebSocketDespatch()
	}
}
*/
