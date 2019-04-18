// +build consul

package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/configs"
	"cy/game/db/mgo"
	pbinner "cy/game/pb/inner"
	"cy/game/util"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.0.90:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9301", "listen address")
	release    = flag.Bool("release", false, "run mode release")
	redisAddr  = flag.String("redisaddr", "192.168.0.90:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.0.90:27017/game", "mongo connection URI")

	redisPool *redis.Pool
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	globalcnf := configs.GetConfig("./run_env/globalconf.json")
	*consulAddr = globalcnf.ConsulAddr
	*release = globalcnf.Release
	*redisAddr = globalcnf.RedisAddr
	*redisDb = globalcnf.RedisDb
	*mgoURI = globalcnf.MgoURI
	*addr = globalcnf.ClubConf.Addr
}

type club int

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if *release {
		logName := fmt.Sprintf("club_%d_%d.log", os.Getpid(), time.Now().Unix())
		file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logrus.SetOutput(file)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logName := fmt.Sprintf("./log/club.log")
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

	go util.Subscribe(*redisAddr, *redisDb, "inner_broadcast", onMessage)

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

	err = mgo.Init(*mgoURI)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	loadDB()
	syncDB()

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

	s.RegisterName("club", new(club), "")
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

func toGateNormal(pb proto.Message, uids ...uint64) error {
	if len(uids) == 0 {
		return nil
	}

	logrus.WithFields(logrus.Fields{"pb": pb, "to": uids, "name": proto.MessageName(pb)}).Info("send")

	msg := &codec.Message{}
	err := codec.Pb2Msg(pb, msg)
	if err != nil {
		return err
	}

	var xx struct {
		Msg  *codec.Message
		Uids []uint64
	}
	xx.Msg = msg
	xx.Uids = append(xx.Uids, uids...)

	data, err := json.Marshal(xx)
	if err != nil {
		return err
	}

	rc := redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	if err != nil {
		logrus.Error(err.Error())
	}
	return err
}

func onMessage(channel string, data []byte) error {
	m := &codec.Message{}
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}

	pb, err := protobuf.Unmarshal(m.Name, m.Payload)
	if err != nil {
		return err
	}

	switch v := pb.(type) {
	case *pbinner.UserChangeNotif:
		cu := mustGetUserOther(v.UserID)
		cu.Lock()
		if v.Typ == pbinner.UserChangeType_Online {
			cu.Online = 1
		} else if v.Typ == pbinner.UserChangeType_Offline {
			cu.Online = 0
		}
		cu.Unlock()
	case *pbinner.DeskChangeNotif:
		flashDesk(v)
	}

	return nil
}
