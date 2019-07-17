package main

import (
	"cy/other/im/cache"
	zaplog "cy/other/im/common/logger"
	"cy/other/im/util"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/smallnest/rpcx/client"
	"go.uber.org/zap"
)

var (
	consulAddr = flag.String("consulAddr", "127.0.0.1:8500", "consul address")
	basePath   = flag.String("base", "/cy_im", "consul prefix path")
	addr       = flag.String("addr", "127.0.0.1:9876", "tcp listen address")
	wsAddr     = flag.String("wsaddr", "127.0.0.1:9877", "ws listen address")
	iaddr      = flag.String("iaddr", "", "inner listen address")
	redisAddr  = flag.String("redisaddr", "192.168.0.10:6379", "redis address")
	release    = flag.Bool("release", false, "run mode")
	nodeName   = flag.String("nodeName", "gate", "nodeName")

	mgr = newManager()

	cliLogic  client.XClient
	cliFriend client.XClient

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/%s_%d_%s.log", *nodeName, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/%s.log", *nodeName)
		logLevel = "debug"
	}
	tlog = zaplog.InitLogger(logName, logLevel, !*release)
	log = tlog.Sugar()
}

func main() {
	var (
		needCrypto = flag.Bool("nc", false, "need crypto")
		needMAC    = flag.Bool("nm", false, "need MAC")
	)

	flag.Parse()

	defer func() {
		if r := recover(); r != nil {

		}
		log.Errorf("recover info:stack:%s", string(debug.Stack()))
	}()

	initLog()

	if err := cache.Init(*redisAddr); err != nil {
		log.Error(err.Error())
		return
	}

	{
		d := client.NewConsulDiscovery(*basePath, "Logic", []string{*consulAddr}, nil)
		cliLogic = client.NewXClient("Logic", client.Failtry, client.SelectByUser, d, client.DefaultOption)
		cliLogic.SetSelector(&selectByID{})
	}

	{
		d := client.NewConsulDiscovery(*basePath, "Friend", []string{*consulAddr}, nil)
		cliFriend = client.NewXClient("Friend", client.Failover, client.RoundRobin, d, client.DefaultOption)
	}

	if *iaddr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			return
		}
		*iaddr = taddr.String()
	}

	log.Info("listen at:", *iaddr)

	go innerServer()

	cfg := &serverConfig{
		id:         *iaddr,
		needCrypto: *needCrypto,
		needMAC:    *needMAC,
	}

	go func() {
		sWs := newWsServer()
		sWs.start()
	}()

	stcp := newTcpServer(cfg)

	fmt.Println(stcp.start("tcp", *addr))
}
