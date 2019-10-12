package api

import (
	"crypto/md5"
	"game/codec"
	zaplog "game/common/logger"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

const (
	Key string = "eo05Efekb*1sMuM6"
)

// RspCode RspCode
type RspCode int

const (
	Succ RspCode = iota
	Failed
	ArgInvalid
	NotFound
)

type apiRsp struct {
	Code RspCode     `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

var (
	redisPool *redis.Pool
	Log       *zap.SugaredLogger //printf风格
	Tlog      *zap.Logger        //structured 风格
)

func InitLog(release bool, logname string) {
	var logName, logLevel string
	if release {
		logLevel = "info"
		logName = fmt.Sprintf("./log/%s_%d_%s.log", logname, os.Getpid(), time.Now().Format("2006_01_02"))
	} else {
		logName = fmt.Sprintf("./log/%s.log", logname)
		logLevel = "debug"
	}
	Tlog = zaplog.InitLogger(logName, logLevel, !release)
	Log = Tlog.Sugar()
}

func InitRedis(redisAddr string, redisDb int) {
	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", redisDb); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

func MakeMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

//向路由发送消息
func ToGateNormal(pb proto.Message, uids ...uint64) error {
	if len(uids) == 0 {
		return nil
	}
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
		Log.Error(err.Error())
	}
	return err
}
