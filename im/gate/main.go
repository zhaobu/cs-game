package main

import (
	"cy/im/cache"
	"cy/util"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
)

var (
	consulAddr = flag.String("consulAddr", "localhost:8500", "consul address")
	basePath   = flag.String("base", "/cy_im", "consul prefix path")
	addr       = flag.String("addr", "localhost:9876", "tcp listen address")
	wsAddr     = flag.String("wsaddr", "localhost:9877", "ws listen address")
	iaddr      = flag.String("iaddr", "", "inner listen address")
	redisAddr  = flag.String("redisaddr", "192.168.0.213:6379", "redis address")

	mgr = newManager()

	cliLogic  client.XClient
	cliFriend client.XClient
)

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logName := fmt.Sprintf("gate_%d_%d.log", os.Getpid(), time.Now().Unix())
	file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}
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
		logrus.WithFields(logrus.Fields{}).Warn(string(debug.Stack()))
	}()

	initLog()

	if err := cache.Init(*redisAddr); err != nil {
		logrus.Error(err.Error())
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

	logrus.Info("listen at:", *iaddr)

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
