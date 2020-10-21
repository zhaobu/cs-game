// +build consul

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"game/cache"
	"game/codec"
	"game/codec/protobuf"
	zaplog "game/common/logger"
	"game/configs"
	"game/db/mgo"
	pbinner "game/pb/inner"
	"game/util"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"go.uber.org/zap"
)

type club int

var (
	consulAddr = flag.String("consulAddr", "192.168.0.90:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	addr       = flag.String("addr", "localhost:9301", "listen address")
	release    = flag.Bool("release", false, "run mode release")
	redisAddr  = flag.String("redisaddr", "192.168.0.90:6379", "redis address")
	redisDb    = flag.Int("redisDb", 1, "redis db select")
	mgoURI     = flag.String("mgo", "mongodb://192.168.0.90:27017/game", "mongo connection URI")

	redisCli *redis.Client

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	*consulAddr = configs.Conf.ConsulAddr
	*release = configs.Conf.Release
	*redisAddr = configs.Conf.RedisAddr
	*redisDb = configs.Conf.RedisDb
	*mgoURI = configs.Conf.MgoURI
	*addr = configs.Conf.ClubConf.Addr
}

func initLog() {
	var logName, logLevel string
	if *release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/club_%d_%d.log", os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/club.log")
		logLevel = "debug"
	}
	tlog = zaplog.InitLogger(logName, logLevel, !*release)
	log = tlog.Sugar()
}

func main() {
	flag.Parse()

	defer func() {
		r := recover()
		if r != nil {
			log.Errorf(string(debug.Stack()))
		}
	}()

	initLog()

	err := cache.Init(*redisAddr, *redisDb)
	if err != nil {
		fmt.Println(err)
		return
	}

	go util.RedisXread(*redisAddr, *redisDb, "inner_broadcast", onMessage)

	redisCli = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",       // no password set
		DB:       *redisDb, // use default DB
	})

	err = mgo.Init(*mgoURI)
	if err != nil {
		tlog.Error(err.Error())
		return
	}

	loadDB()
	syncDB()

	if *release && *addr == "" {
		taddr, err := util.AllocListenAddr()
		if err != nil {
			tlog.Warn(err.Error())
			return
		}
		*addr = taddr.String()
	}
	log.Infof("listen at %s", *addr)

	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("club", new(club), "")
	err = s.Serve("tcp", *addr)
	if err != nil {
		log.Errorf("Serve err :%s", err)
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

	tlog.Debug("send toGateNormal", zap.Any("pb", pb), zap.Any("uids", uids), zap.String("msgname", proto.MessageName(pb)))

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

	_, err = util.RedisXadd(redisCli, "backend_to_gate", data)
	if err != nil {
		tlog.Error(err.Error())
	}
	return err
}

func onMessage(channel string, msgData []byte) error {
	var m codec.Message
	err := json.Unmarshal(msgData, &m)
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
