package main

import (
	"flag"
	"fmt"
	// "game/configs"

	"github.com/globalsign/mgo"
	"github.com/go-redis/redis"
)

var (
	mgoURI    = flag.String("mgo", "mongodb://127.0.0.1:27017/game", "mongodb uri")
	redisAddr = flag.String("redisAddr", "127.0.0.1:6379", "redis address")
	redisDb   = flag.Int("redisDb", 1, "redis db select")

	mgoSess  *mgo.Session
	redisCli *redis.Client
)

func init() {
	//如果不指定启动参数,默认读取全局配置
	// *redisAddr = configs.Conf.RedisAddr
	// *redisDb = configs.Conf.RedisDb
	// *mgoURI = configs.Conf.MgoURI
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
	redisCli = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",       // no password set
		DB:       *redisDb, // use default DB
	})
}