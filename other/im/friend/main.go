package main

import (
	"cy/other/im/cache"
	"cy/other/im/util"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
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

	cliGate client.XClient

	tsdbCli *tablestore.TableStoreClient
)

type friend int

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logName := fmt.Sprintf("friend_%d_%d.log", os.Getpid(), time.Now().Unix())
	file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}
}

func main() {
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {

		}
		fmt.Println(string(debug.Stack()))
	}()

	if *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			fmt.Println()
			return
		}
		*addr = taddr.String()
	}

	logrus.Info("listen at:", *addr)

	if err := cache.Init(*redisAddr); err != nil {
		logrus.Error(err.Error())
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
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}

func queryPlace(userID uint64) string {
	gateID, _, _ := cache.QueryUser(userID)
	return gateID
}
