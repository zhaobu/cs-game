package main

import (
	"crypto/tls"
	"cy/game/cache"
	zaplog "cy/game/common/logger"
	"cy/game/configs"
	"cy/game/db/mgo"
	"cy/game/net"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/smallnest/rpcx/client"
	"go.uber.org/zap"
)

var (
	gateName   = flag.String("gateName", "gate", "gateName")
	addr       = flag.String("addr", "192.168.0.10:9876", "tcp listen address")
	certFile   = flag.String("cert", "", "cert file")
	keyFile    = flag.String("key", "", "key file")
	useTLS     = flag.Bool("tls", false, "use TLS")
	consulAddr = flag.String("consulAddr", "192.168.0.90:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	release    = flag.Bool("release", false, "run mode")
	redisAddr  = flag.String("redisAddr", "192.168.0.90:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.0.90:27017/game", "mongo connection URI")
	aliAppCode = flag.String("aliCode", `4fc26616d3304bacb757c9bb503e02be`, "ali APPCODE")
	netAddr    = flag.String("netaddr", `http://192.168.0.207:8096`, ",Net Addr") //后台服务器地址
	mgr        = newManager()

	cliCenter client.XClient
	cliClub   client.XClient

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/%s_%d_%s.log", *gateName, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/%s.log", *gateName)
		logLevel = "debug"
	}
	tlog = zaplog.InitLogger(logName, logLevel, !*release)
	log = tlog.Sugar()
}

// func initLogrus() {
// 	logrus.SetFormatter(&logrus.JSONFormatter{})
// 	if *release {
// 		logName := fmt.Sprintf("gate_%d_%d.log", os.Getpid(), time.Now().Unix())
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
// 		if err == nil {
// 			logrus.SetOutput(file)
// 		} else {
// 			logrus.SetOutput(os.Stdout)
// 		}
// 	} else {
// 		logName := fmt.Sprintf("./log/gate.log")
// 		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
// 		if err == nil {
// 			logrus.SetOutput(file)
// 		} else {
// 			logrus.SetOutput(os.Stdout)
// 		}
// 	}
// }

func init() {
	//如果不指定启动参数,默认读取全局配置
	*consulAddr = configs.Conf.ConsulAddr
	*release = configs.Conf.Release
	*redisAddr = configs.Conf.RedisAddr
	*redisDb = configs.Conf.RedisDb
	*mgoURI = configs.Conf.MgoURI
	*netAddr = configs.Conf.NetAddr
	*addr = configs.Conf.GateConf.Addr
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			log.Warn(string(debug.Stack()))
		}
	}()

	initLog()

	var err error
	err = cache.Init(*redisAddr, *redisDb)
	if err != nil {
		log.Error(err.Error())
		return
	}

	err = mgo.Init(*mgoURI)
	if err != nil {
		log.Error(err.Error())
		return
	}
	net.Init(*netAddr) //初始化net

	subscribeBackend(*redisAddr, *redisDb)
	{
		servicePath := "center"
		d := client.NewConsulDiscovery(*basePath, servicePath, []string{*consulAddr}, nil)
		cliCenter = client.NewXClient(servicePath, client.Failfast, client.RoundRobin, d, client.DefaultOption)
	}

	{
		servicePath := "club"
		d := client.NewConsulDiscovery(*basePath, servicePath, []string{*consulAddr}, nil)
		cliClub = client.NewXClient(servicePath, client.Failfast, client.RoundRobin, d, client.DefaultOption)
	}

	config := &serverConfig{}

	if *useTLS {
		certificate, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		config.tlsConfig = &tls.Config{Certificates: []tls.Certificate{certificate}}
	}

	tcpSrv := newTCPServer(config)
	log.Error(tcpSrv.start(*addr))

}
