package main

import (
	"cy/other/im/cache"
	. "cy/other/im/common/logger"
	"cy/other/im/util"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/smallnest/rpcx/client"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.0.10:8500", "consul address")
	basePath   = flag.String("base", "/cy_im", "consul prefix path")
	addr       = flag.String("addr", "192.168.0.10:9876", "tcp listen address")
	wsAddr     = flag.String("wsaddr", "192.168.0.10:9877", "ws listen address")
	iaddr      = flag.String("iaddr", "", "inner listen address")
	redisAddr  = flag.String("redisaddr", "192.168.0.10:6379", "redis address")
	release    = flag.Bool("release", false, "run mode")
	nodeName   = flag.String("nodeName", "gate", "nodeName")

	mgr = newManager()

	cliLogic  client.XClient
	cliFriend client.XClient
)

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./Log/%s_%d_%s.Log", *nodeName, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./Log/%s.Log", *nodeName)
		logLevel = "debug"
	}
	InitLogger(logName, logLevel, !*release)
}

func main() {
	var (
		needCrypto = flag.Bool("nc", false, "need crypto")
		needMAC    = flag.Bool("nm", false, "need MAC")
	)

	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			Log.Errorf("recover info:stack:%s", string(debug.Stack()))
		}
	}()

	initLog()

	if err := cache.Init(*redisAddr); err != nil {
		Log.Error(err.Error())
		return
	}

	go innerServer()

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

	Log.Info("listen at:", *iaddr)


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

	Log.Info(stcp.start("tcp", *addr))
}
