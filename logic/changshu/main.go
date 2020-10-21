package main

import (
	"flag"
	"fmt"
	"game/cache"
	zaplog "game/common/logger"
	"game/configs"
	"game/db/mgo"
	"game/logic/tpl"
	"game/net"
	"game/util"
	"os"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	serverplugin "github.com/smallnest/rpcx/serverplugin"
	"go.uber.org/zap"
)

const (
	gameName = "11101"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.0.90:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9601", "listen address")
	release    = flag.Bool("release", false, "run mode")
	redisAddr  = flag.String("redisAddr", "192.168.0.90:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.0.90:27017/game", "mongo connection URI")
	netAddr    = flag.String("netaddr", "http://192.168.0.207:8096", ",Net Addr") //后台服务器地址
	log        *zap.SugaredLogger                                                 //printf风格
	tlog       *zap.Logger                                                        //structured 风格
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	*consulAddr = configs.Conf.ConsulAddr
	*release = configs.Conf.Release
	*redisAddr = configs.Conf.RedisAddr
	*redisDb = configs.Conf.RedisDb
	*mgoURI = configs.Conf.MgoURI
	*netAddr = configs.Conf.NetAddr
	*addr = configs.Conf.GameNode[gameName].Addr
}

type roomHandle struct {
	*tpl.RoomServie
	gameCommond
}

// func initLogrus() {
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.JSONFormatter{})
// 	if *release {
// 		logName := fmt.Sprintf("%s_%d_%d.log", gameName, os.Getpid(), time.Now().Unix())
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
// 		if err == nil {
// 			l.SetOutput(file)
// 		} else {
// 			panic(err)
// 		}
// 	} else {
// 		l.SetLevel(logrus.TraceLevel)
// 		logName := fmt.Sprintf("./log/%s.log", gameName)
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
// 		if err == nil {
// 			l.SetOutput(file)
// 		} else {
// 			panic(err)
// 		}
// 	}

// 	// hook, err := logrus_influxdb.NewInfluxDB(&logrus_influxdb.Config{
// 	// 	Host:          "192.168.0.90", // TODO
// 	// 	Port:          8086,
// 	// 	Database:      "cygame",
// 	// 	Precision:     "ns",
// 	// 	Tags:          []string{"serverid", "deskid", "uid"},
// 	// 	BatchInterval: (5 * time.Second),
// 	// 	Measurement:   gameName,
// 	// 	BatchCount:    0, // set to "0" to disable batching
// 	// })

// 	// if err == nil {
// 	// 	l.Hooks.Add(hook)
// 	// }

// 	log = l.WithFields(logrus.Fields{})
// }

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/%s_%d_%s.log", gameName, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/%s.log", gameName)
		logLevel = "debug"
		os.RemoveAll("./log/roomlog/")
	}
	tlog = zaplog.InitLogger(logName, logLevel, !*release)
	log = tlog.Sugar()
}

func main() {
	flag.Parse()

	initLog()
	if *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			fmt.Println(err)
			return
		}
		*addr = taddr.String()
	}
	roomService := &tpl.RoomServie{}
	roomService.Init(gameName, *addr, tlog, *redisAddr, *redisDb)
	roomService.RegisterHandle(&roomHandle{RoomServie: roomService})

	if !*release {
		cache.FlushDb(*redisDb)
	}

	var err error
	err = loadArgTpl(configs.Conf.GameNode[gameName].TplName)
	if err != nil {
		tlog.Error("loadArgTpl err", zap.Error(err))
		return
	}

	err = mgo.Init(*mgoURI)
	if err != nil {
		tlog.Error("mgo.Init err", zap.Error(err))
		return
	}

	net.Init(*netAddr) //初始化net
	go func() {
		err = net.GetCondition() //获取抽奖配置表
		if err != nil {
			tlog.Error("net.GetCondition err", zap.Error(err))
		}
	}()

	s := server.NewServer()
	addRegistryPlugin(s)
	log.Infof("gameserver %s 启动成功", gameName)

	s.RegisterName("game/"+gameName, roomService.GetRpcHandle(), "")
	err = s.Serve("tcp", *addr)
	if err != nil {
		tlog.Error("Serve err", zap.Error(err))
	}
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.ConsulRegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		ConsulServers:  []string{*consulAddr},
		BasePath:       *basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
