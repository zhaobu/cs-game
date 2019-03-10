package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rcrowley/go-metrics"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	addr = flag.String("l", "localhost:9876", "listen addr")

	gate *Gate
)

func ws(c *gin.Context) {
	wc, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	gate.HandleWs(wc)
}

func debugStatus(c *gin.Context) {
	c.JSON(200, atomic.LoadInt64(&gate.acceptedCnt))
}

func main() {
	flag.Parse()

	go metrics.Log(metrics.DefaultRegistry, time.Second*10, log.New(os.Stderr, "m", log.LstdFlags))

	gin.SetMode(gin.ReleaseMode)

	gate = NewGate()
	go gate.StartListen(*addr)

	r := gin.Default()
	r.GET("/ws", ws)
	r.GET("/debug", debugStatus)
	r.Run() // listen and serve on 0.0.0.0:8080
}
