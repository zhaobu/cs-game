package main

import (
	"cy/game/cache"
	"cy/game/logic/ddz/desk"
	"cy/game/util"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Abramovic/logrus_influxdb"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

const (
	gameName = "ddz" // 注册的游戏名称，不能有冲突
)

var (
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9501", "listen address")
	release    = flag.Bool("release", false, "run mode")
	redisAddr  = flag.String("redisaddr", "192.168.1.128:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.1.128:27017/game", "mongo connection URI")
	log        *logrus.Entry
	gameID     string // 默认为IP:PORT 所以IP不能为localhost和127.0.0.1
)

type ddz int

func initLog() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	if *release {
		logName := fmt.Sprintf("ddz_%d_%d.log", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			l.SetOutput(file)
		} else {
			panic(err)
		}
	} else {
		logName := fmt.Sprintf("ddz.log")
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			l.SetOutput(file)
		} else {
			panic(err)
		}
	}

	hook, err := logrus_influxdb.NewInfluxDB(&logrus_influxdb.Config{
		Host:          "192.168.1.128", // TODO
		Port:          8086,
		Database:      "cygame",
		Precision:     "ns",
		Tags:          []string{"serverid", "deskid", "uid"},
		BatchInterval: (5 * time.Second),
		Measurement:   "ddz",
		BatchCount:    0, // set to "0" to disable batching
	})

	if err == nil {
		_ = hook
		//l.Hooks.Add(hook)
	}

	log = l.WithFields(logrus.Fields{})
}

func main() {
	flag.Parse()

	initLog()

	var err error
	defer func() {
		if err != nil {
			log.Error(err.Error())
		}
	}()

	if *release && *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			return
		}
		*addr = taddr.String()
	}

	gameID = *addr

	log = log.WithFields(logrus.Fields{"serverid": gameID})
	log.Infof("listen at %s", *addr)

	err = cache.Init(*redisAddr, *redisDb)
	if err != nil {
		return
	}

	err = desk.Init(*redisAddr, *redisDb, *mgoURI, gameName, gameID, log)
	if err != nil {
		return
	}

	err = desk.LoadConfig("./config.json")
	if err != nil {
		return
	}

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("game/"+gameName, new(ddz), "") // 注册的游戏名称必须用"game/"开头
	err = s.Serve("tcp", *addr)
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
