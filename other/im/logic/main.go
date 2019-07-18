package main

import (
	"cy/other/im/cache"
	"cy/other/im/util"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	. "cy/other/im/common/logger"

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
	release  = flag.Bool("release", false, "run mode")
	nodeName = flag.String("nodeName", "logic", "nodeName")
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
		consulAddr = flag.String("consulAddr", "192.168.0.10:8500", "consul address")
		basePath   = flag.String("base", "/cy_im", "consul prefix path")
		addr       = flag.String("addr", "", "listen address")
		redisAddr  = flag.String("redisaddr", "192.168.0.10:6379", "redis address")
	)
	flag.Parse()

	initLog()

	defer func() {
		if r := recover(); r != nil {
			Log.Error(string(debug.Stack()))
		}
	}()

	InitTS()

	if err := cache.Init(*redisAddr); err != nil {
		Log.Error(err.Error())
		return
	}

	//tw.Start()

	var err error

	if *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			Log.Error(err.Error())
			return
		}
		*addr = taddr.String()
	}

	Log.Info("listen at:", *addr)

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
		Log.Error(err.Error())
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
		Log.Error(err.Error())
	}
	s.Plugins.Add(r)
}

func queryPlace(userID uint64) string {
	gateID, _, _ := cache.QueryUser(userID)
	return gateID
}
