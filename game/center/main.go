package main

import (
	"cy/game/cache"
	"cy/game/util"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9201", "listen address")
	release    = flag.Bool("release", false, "run mode release")
	redisAddr  = flag.String("redisaddr", "192.168.1.128:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
)

type center int

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *release {
		logName := fmt.Sprintf("center_%d_%d.log.json", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logName := fmt.Sprintf("center.log.json")
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
		r := recover()
		if r != nil {
			logrus.Warn(string(debug.Stack()))
		}

	}()

	initLog()

	err := cache.Init(*redisAddr, *redisDb)
	if err != nil {
		fmt.Println(err)
		return
	}

	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", *redisAddr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", *redisDb); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

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

	s.RegisterName("center", new(center), "")
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
