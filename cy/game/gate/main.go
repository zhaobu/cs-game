package main

import (
	"crypto/tls"
	"cy/game/cache"
	"cy/game/db/mgo"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
)

var (
	addr       = flag.String("addr", "localhost:9876", "tcp listen address")
	certFile   = flag.String("cert", "", "cert file")
	keyFile    = flag.String("key", "", "key file")
	useTLS     = flag.Bool("tls", false, "use TLS")
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	release    = flag.Bool("release", false, "run mode")
	redisAddr  = flag.String("redisaddr", "192.168.1.128:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.1.128:27017/game", "mongo connection URI")

	mgr = newManager()

	cliCenter client.XClient
	cliClub   client.XClient
)

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *release {
		logName := fmt.Sprintf("gate_%d_%d.log", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logName := fmt.Sprintf("gate.log")
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	}
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			logrus.Warn(string(debug.Stack()))
		}
	}()

	initLog()

	var err error

	err = cache.Init(*redisAddr, *redisDb)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	err = mgo.Init(*mgoURI)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

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
			logrus.Warn(err.Error())
			return
		}

		config.tlsConfig = &tls.Config{Certificates: []tls.Certificate{certificate}}
	}

	tcpSrv := newTCPServer(config)
	logrus.Error(tcpSrv.start(*addr))

}
