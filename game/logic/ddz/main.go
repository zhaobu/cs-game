package main

import (
	"cy/game/cache"
	"cy/game/logic/ddz/desk"
	"cy/game/util"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

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

	gameID string // 默认为IP:PORT 所以IP不能为localhost和127.0.0.1
)

type ddz int

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *release {
		logName := fmt.Sprintf("ddz_%d_%d.log", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logName := fmt.Sprintf("ddz.log")
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	}
	// logrus.AddHook()
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			logrus.Warn(string(debug.Stack()))
		}
	}()

	initLog()

	err := cache.Init(*redisAddr, *redisDb)
	if err != nil {
		fmt.Println(err)
		return
	}

	if *release && *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			logrus.Warn(err.Error())
			return
		}
		*addr = taddr.String()
	}
	logrus.Infof("listen at %s", *addr)

	gameID = *addr

	if err := desk.LoadConfig("./config.json"); err != nil {
		logrus.WithFields(logrus.Fields{"err": err.Error()}).Error("load config")
		return
	}

	if err := desk.Init(*redisAddr, *redisDb, *mgoURI, gameName, gameID); err != nil {
		return
	}

	// fmt.Println(desk.UpdateWealthPreSure(1002, 2, 50))
	// return

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("game/"+gameName, new(ddz), "") // 注册的游戏名称必须用"game/"开头
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
