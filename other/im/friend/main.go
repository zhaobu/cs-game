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

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	consulAddr      = flag.String("consulAddr", "localhost:8500", "consul address")
	basePath        = flag.String("base", "/cy_im", "consul prefix path")
	addr            = flag.String("addr", "", "listen address")
	endpoint        = flag.String("endpoint", `https://zztest-1009.cn-hangzhou.ots.aliyuncs.com`, "endpoint")
	instanceName    = flag.String("instanceName", `zztest-1009`, "instanceName")
	accessKeyID     = flag.String("accessKeyID", `LTAIssLCxHELxHAq`, "accessKeyId")
	accessKeySecret = flag.String("accessKeySecret", `645bzZ5iJxPru921GNrvkYNIm2Uhnf`, "accessKeySecret")
	redisAddr       = flag.String("redisaddr", "192.168.0.10:6379", "redis address")
	release         = flag.Bool("release", false, "run mode")
	nodeName        = flag.String("nodeName", "friend", "nodeName")

	cliGate client.XClient

	tsdbCli *tablestore.TableStoreClient
)

type friend int

func initlog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/%s_%d_%s.log", *nodeName, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/%s.log", *nodeName)
		logLevel = "debug"
	}
	InitLogger(logName, logLevel, !*release)
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {

		}
		fmt.Println(string(debug.Stack()))
	}()

	initlog()
	if *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			fmt.Println()
			return
		}
		*addr = taddr.String()
	}
	Log.Infof("listen at:%s", *addr)

	if err := cache.Init(*redisAddr); err != nil {
		Log.Error(err)
		return
	}

	tsdbCli = tablestore.NewClient(*endpoint, *instanceName, *accessKeyID, *accessKeySecret)

	{
		d := client.NewConsulDiscovery(*basePath, "Gate", []string{*consulAddr}, nil)
		cliGate = client.NewXClient("Gate", client.Failfast, client.SelectByUser, d, client.DefaultOption)
		cliGate.SetSelector(&selectByToID{})
	}

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("Friend", new(friend), "")
	err := s.Serve("tcp", *addr)
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
		Log.Fatal(err)
	}
	s.Plugins.Add(r)
}

func queryPlace(userID uint64) string {
	gateID, _, _ := cache.QueryUser(userID)
	return gateID
}
