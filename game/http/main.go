package main

import (
	"game/configs"
	"game/db/mgo"
	"game/http/api"
	"flag"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"runtime/debug"
)

var (
	httpName  = flag.String("mame", "http", "mame")
	release   = flag.Bool("release", false, "run mode")
	addr	= flag.String("addr", ":9080", "run addr")
	redisAddr = flag.String("redisaddr", "127.0.0.1:6379", "redis address")
	redisDb   = flag.Int("redisDb", 1, "redis db select")
	mgoURI    = flag.String("mgo", "mongodb://192.168.0.205:27017/game", "mongo connection URI")
)

func init() {
	*release = configs.Conf.Release
	*addr = configs.Conf.HttpConf.Addr
	*redisAddr = configs.Conf.RedisAddr
	*redisDb = configs.Conf.RedisDb
	*mgoURI = configs.Conf.MgoURI
}

func main() {
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			api.Log.Warn(string(debug.Stack()))
		}
	}()
	api.InitLog(*release,*httpName)
	err := mgo.Init(*mgoURI)
	if err != nil {
		api.Tlog.Error(err.Error())
		return
	}
	api.InitRedis(*redisAddr, *redisDb)

	r := gin.Default()
	r.POST("/UpdateUserWealth", api.UpdateUserWealthReq)							//更新用户财富
	r.POST("/SetUserRedName", api.SetUserRedNameReq)								//设置用户红名
	r.POST("/QueryUserParGameInfo", api.QueryUserParGameInfoReq)					//查询用户参与游戏信息
	r.POST("/QueryUserGameRecord", api.QueryUserGameRecordReq)						//查询用户战绩信息
	r.POST("/BuildAgent", api.BuildAgentReq)										//绑定代理信息
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(*addr)
}

