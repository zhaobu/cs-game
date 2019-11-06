package main

import (
	"flag"
	"fmt"
	"game/cache"
	"game/configs"
	"game/db/mgo"
	"game/util"
	"os"
	"runtime/debug"
	"time"

	zaplog "game/common/logger"

	"github.com/go-redis/redis"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"go.uber.org/zap"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.0.90:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9201", "listen address")
	release    = flag.Bool("release", false, "run mode release")
	redisAddr  = flag.String("redisaddr", "192.168.0.90:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.0.90:27017/game", "mongo connection URI")

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	*consulAddr = configs.Conf.ConsulAddr
	*release = configs.Conf.Release
	*redisAddr = configs.Conf.RedisAddr
	*redisDb = configs.Conf.RedisDb
	*mgoURI = configs.Conf.MgoURI
	*addr = configs.Conf.CenterConf.Addr
}

type center int

// func initLog() {
// 	logrus.SetFormatter(&logrus.JSONFormatter{})
// 	if *release {
// 		logName := fmt.Sprintf("center_%d_%d.log", os.Getpid(), time.Now().Unix())
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
// 		if err == nil {
// 			logrus.SetOutput(file)
// 		} else {
// 			logrus.SetOutput(os.Stdout)
// 		}
// 	} else {
// 		logName := fmt.Sprintf("./log/center.log")
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
// 		if err == nil {
// 			logrus.SetOutput(file)
// 		} else {
// 			logrus.SetOutput(os.Stdout)
// 		}
// 	}
// }

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/center_%d_%d.log", os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/center.log")
		logLevel = "debug"
	}
	tlog = zaplog.InitLogger(logName, logLevel, !*release)
	log = tlog.Sugar()
}

func main() {
	flag.Parse()

	defer func() {
		r := recover()
		if r != nil {
			tlog.Warn(string(debug.Stack()))
		}

	}()

	initLog()

	GameRecord_Init() //战绩初始化

	err := cache.Init(*redisAddr, *redisDb)
	if err != nil {
		fmt.Println(err)
		return
	}

	redisCli = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",       // no password set
		DB:       *redisDb, // use default DB
	})

	err = mgo.Init(*mgoURI)
	if err != nil {
		tlog.Error(err.Error())
		return
	}

	if *release && *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			tlog.Warn(err.Error())
			return
		}
		*addr = taddr.String()
	}
	log.Infof("listen at %s", *addr)

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("center", new(center), "")
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
