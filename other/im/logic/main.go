package main

import (
	"cy/im/cache"
	"cy/im/logic/db"
	"cy/util"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/RussellLuo/timingwheel"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

type logic int

var (
	cliGate client.XClient

	//tw = timingwheel.NewTimingWheel(time.Millisecond, 20)
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logName := fmt.Sprintf("logic_%d_%d.log", os.Getpid(), time.Now().Unix())
	file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}
}

func main() {
	var (
		consulAddr = flag.String("consulAddr", "localhost:8500", "consul address")
		basePath   = flag.String("base", "/cy_im", "consul prefix path")
		addr       = flag.String("addr", "", "listen address")
		redisAddr  = flag.String("redisaddr", "192.168.0.213:6379", "redis address")
	)
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			logrus.Error(string(debug.Stack()))
		}
	}()

	db.InitTS()

	if err := cache.Init(*redisAddr); err != nil {
		logrus.Error(err.Error())
		return
	}

	//tw.Start()

	var err error

	if *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			logrus.Error(err.Error())
			return
		}
		*addr = taddr.String()
	}

	logrus.Info("listen at:", *addr)

	{
		d := client.NewConsulDiscovery(*basePath, "Gate", []string{*consulAddr}, nil)
		cliGate = client.NewXClient("Gate", client.Failfast, client.SelectByUser, d, client.DefaultOption)
		cliGate.SetSelector(&selectByToID{})
	}

	s := server.NewServer()
	addRegistryPlugin(s, *addr, *consulAddr, *basePath)

	s.RegisterName("Logic", new(logic), "")
	err = s.Serve("tcp", *addr)
	if err != nil {
		logrus.Error(err.Error())
	}
}

func addRegistryPlugin(s *server.Server, addr, consulAddr, basePath string) {
	r := &serverplugin.ConsulRegisterPlugin{
		ServiceAddress: "tcp@" + addr,
		ConsulServers:  []string{consulAddr},
		BasePath:       basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Error(err.Error())
	}
	s.Plugins.Add(r)
}

func queryPlace(userID uint64) string {
	gateID, _, _ := cache.QueryUser(userID)
	return gateID
}
