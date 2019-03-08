package main

import (
	"flag"
	"fmt"
	"io"
	"mahjong-league/cli"
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/model"
	"mahjong-league/protocal"
	"mahjong-league/router"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"

	"github.com/fwhappy/util"
)

var (
	// 监听ip
	host = flag.String("h", "0.0.0.0", "host")
	// 监听端口
	port = flag.String("port", "31212", "listen port")
	// webposrt
	wport = flag.String("wport", "31213", "web port")
	// 配置文件夹
	confDir = flag.String("confDir", "etc", "config dir path")
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 服务器版本
	version = flag.String("v", "latest", "server version")
)

func init() {
	// 解析url参数
	flag.Parse()
}

func main() {
	// 开启多核模式
	runtime.GOMAXPROCS(runtime.NumCPU() * 5)

	// 开启性能监控
	go func() {
		http.ListenAndServe(":31214", nil)
	}()

	// 记录服务器相关信息
	hall.ServerRemote = *host + ":" + *port
	hall.ServerVersion = *version

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
	// 初始化机器人配置
	core.LoadRobotConfig(core.GetConfigFile("robot.toml", *env, *confDir))

	core.Logger.Info("配置文件初始化完成")

	// 支持热重启，需要在启动时候恢复数据
	// 恢复大厅比赛列表
	model.LeagueList.Restore()
	// 回复比赛奖励
	model.LeagueRewardsList.Restore()
	// 恢复比赛列表
	model.RaceList.Restore()
	// 恢复比赛用户数据
	model.RestoreLeagueRaceUser()
	// 恢复比赛房间
	model.RestoreRaceRooms()

	// 恢复排赛channel
	model.RestoreLeagueRacePlanChannel()
	core.Logger.Info("从数据库到内存的数据恢复完成")

	// 开启监听比赛列表
	go cli.ListenLeagueListRefresh()
	// 开启监听排赛队列
	go cli.ListenFullRace()
	// 开启定时赛的监听
	go cli.ListenFixRace()

	// 记录服务器版本
	model.SetRemoteVersion(hall.ServerRemote, hall.ServerVersion)
	// 开启系统活跃心跳
	go hall.StatLastActionTime()

	// 开启端口监听
	tcpListener, err := startListen()
	if err != nil {
		core.Logger.Error("[startListen]开启端口监听失败,error:%v", err.Error())
		return
	}

	// 提供服务
	go acceptListen(tcpListener)

	acceptWeb()
}

func startListen() (*net.TCPListener, error) {
	listenRemote := ":" + *port
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenRemote)
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	core.Logger.Info("server lisen at: %v", listenRemote)
	return tcpListener, nil
}

// listen 开始监听并提供服务
func acceptListen(tcpListener *net.TCPListener) {
	core.Logger.Info("开始接受客户端请求")
	// 监听连接事件
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			core.Logger.Error("[AcceptTCP] %s.", err)
			continue
		}

		// 客户端连接成功，开启新的协程，监听客户端消息
		go serve(tcpConn)
	}
}

func serve(conn *net.TCPConn) {
	defer util.RecoverPanic()
	defer func() {
		hall.KickConn(conn)
		core.Logger.Debug("disconnected: %s.", conn.RemoteAddr().String())
	}()
	core.Logger.Debug("new connected: %s.", conn.RemoteAddr().String())

	// 检测用户连接之后，如果在规定时间内handshake成功，需要断开连接，防止无效的连接
	hall.ListenHandShakeTimeout(conn)

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
				core.Logger.Debug("User disconnected, remote: %s.", conn.RemoteAddr().String())
			case io.ErrUnexpectedEOF:
				core.Logger.Debug("unexpected EOF, remote: %s.", conn.RemoteAddr().String())
			default:
				if strings.Contains(err.Error(), "use of closed network connection") {
					core.Logger.Debug("连接关闭, use of closed network connection, remote:%v", conn.RemoteAddr().String())
				} else if strings.Contains(err.Error(), "connection reset by peer") {
					core.Logger.Debug("连接关闭, connection reset by peer, remote:%v", conn.RemoteAddr().String())
				} else {
					// 协议解析错误
					core.Logger.Error("%v", err.Error())
				}
			}
			break
		}
		// 分发请求
		router.Despatch(conn, impacket)
	}
}

// 监听web服务
func acceptWeb() {
	// 开启路由
	router.WDespatch()

	err := http.ListenAndServe(fmt.Sprintf(":%v", *wport), nil)
	if err != nil {
		core.Logger.Error("[acceptWeb]ListenAndServe:%v", err)
		return
	}
}
