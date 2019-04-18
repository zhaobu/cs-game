package main

import (
	"cy/game/configs"
	"flag"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/gomodule/redigo/redis"
)

var (
	mgoURI    = flag.String("mgo", "mongodb://127.0.0.1:27017/game", "mongodb uri")
	redisAddr = flag.String("redisAddr", "127.0.0.1:6379", "redis address")
	redisDb   = flag.Int("redisDb", 1, "redis db select")

	mgoSess   *mgo.Session
	redisPool *redis.Pool
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	globalcnf := configs.GetConfig("./run_env/globalconf.json")
	*redisAddr = globalcnf.RedisAddr
	*redisDb = globalcnf.RedisDb
	*mgoURI = globalcnf.MgoURI
}

func main() {
	flag.Parse()

	initDb()
	initRedis()

	genDeskID(123456, 123456+100)
	genClubID()

}

func initDb() {
	var err error
	mgoSess, err = mgo.Dial(*mgoURI)
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
}

func initRedis() {
	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", *redisAddr, redis.DialDatabase(*redisDb))
		},
	}
}
