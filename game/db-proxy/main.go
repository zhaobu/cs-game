package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cy/game/db-proxy/mysql"
	"cy/util"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9301", "listen address")
	release    = flag.Bool("release", false, "run mode release")
	dsn        = flag.String("dsn", `root:ZhengZhong~123@tcp(192.168.0.213:3306)/game`, `data source name`)

	mdb *mysql.MysqlDb
)

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *release {
		logName := fmt.Sprintf("dbproxy_%d_%d.log", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logName := fmt.Sprintf("./log/dbproxy.log")
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

	var err error
	mdb, err = mysql.NewMyDb(*dsn)
	if err != nil {
		logrus.Warn(err.Error())
		return
	}

	initLog()

	if *release && *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			logrus.Warn(err.Error())
			return
		}
		*addr = taddr.String()
	}
	logrus.Infof("listen at %s", *addr)

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("dbproxy", new(dbproxy), "")
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
