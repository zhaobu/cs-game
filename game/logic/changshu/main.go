package main

import (
	"cy/game/cache"
	zaplog "cy/game/common/logger"
	"cy/game/configs"
	"cy/game/db/mgo"
	"cy/game/logic/tpl"
	"cy/game/util"
	"flag"
	"fmt"
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

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	globalcnf := configs.GetConfig("../../configs/globalconf.json")
	*consulAddr = globalcnf.ConsulAddr
	*release = globalcnf.Release
	*redisAddr = globalcnf.RedisAddr
	*redisDb = globalcnf.RedisDb
	*mgoURI = globalcnf.MgoURI
}

type roomHandle struct {
	*tpl.RoomServie
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
		_, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Printf("err os.OpenFile()")
		}
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
	roomService.RegisterHandle(&roomHandle{roomService})

	if !*release {
		cache.FlushDb(*redisDb)
	}

	var err error
	err = loadArgTpl("../changshou.json")
	if err != nil {
		fmt.Println(err)
	}

	err = mgo.Init(*mgoURI)
	if err != nil {
		return
	}

	s := server.NewServer()
	addRegistryPlugin(s)
	log.Infof("gameserver %s 启动成功", gameName)

	s.RegisterName("game/"+gameName, roomService.GetRpcHandle(), "")
	err = s.Serve("tcp", *addr)
	if err != nil {
		fmt.Println(err)
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
