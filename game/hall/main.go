package main

import (
	"cy/game/db/mgo"
	_ "cy/game/docs"
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

var (
	consulAddr = flag.String("consulAddr", "192.168.1.128:8500", "consul address")
	basePath   = flag.String("base", "/cy_game", "consul prefix path")
	mgoURI     = flag.String("mgo", "mongodb://192.168.1.128:27017/game", "mongo connection URI")
)

func main() {
	flag.Parse()

	var err error
	err = mgo.Init(*mgoURI)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	r := gin.Default()
	r.GET("/userinfo/:id", userinfo)
	r.POST("/updatewealth", updateWealth)
	r.POST("/bindagent", bindAgent)
	r.GET("/gamelist", gameList)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run()
}
