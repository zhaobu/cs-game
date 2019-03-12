package main

import (
	"cy/game/cache"
	"cy/game/db/mgo"
	"cy/game/logic/tpl"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

const (
	gameName = "mj-cs"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9401", "listen address")
	release    = flag.Bool("release", false, "run mode")
	redisAddr  = flag.String("redisaddr", "192.168.1.128:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.1.128:27017/game", "mongo connection URI")
	log        *logrus.Entry
)

type mjcs struct {
	tpl.RoundTpl
}

func initLog() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	if *release {
		logName := fmt.Sprintf("%s_%d_%d.log", gameName, os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			l.SetOutput(file)
		} else {
			panic(err)
		}
	} else {
		logName := fmt.Sprintf("%s.log", gameName)
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err == nil {
			l.SetOutput(file)
		} else {
			panic(err)
		}
	}

	// hook, err := logrus_influxdb.NewInfluxDB(&logrus_influxdb.Config{
	// 	Host:          "192.168.1.128", // TODO
	// 	Port:          8086,
	// 	Database:      "cygame",
	// 	Precision:     "ns",
	// 	Tags:          []string{"serverid", "deskid", "uid"},
	// 	BatchInterval: (5 * time.Second),
	// 	Measurement:   gameName,
	// 	BatchCount:    0, // set to "0" to disable batching
	// })

	// if err == nil {
	// 	l.Hooks.Add(hook)
	// }

	log = l.WithFields(logrus.Fields{})
}

func main() {
	flag.Parse()

	initLog()

	var cs mjcs
	cs.Log = log
	cs.RoundTpl.SetName(gameName, *addr)
	cs.SetPlugin(&cs)

	err := cache.Init(*redisAddr, *redisDb)
	if err != nil {
		return
	}

	if err := mgo.Init(*mgoURI); err != nil {
		return
	}
	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("game/"+gameName, &cs, "")
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
