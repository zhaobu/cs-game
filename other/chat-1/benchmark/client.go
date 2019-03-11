package main

import (
	"bytes"
	. "cy/chat/def"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rcrowley/go-metrics"
)

var (
	addr      = flag.String("a", "127.0.0.1:8080", "server addr")
	number    = flag.Int64("n", 100, "client number")
	debugMode = flag.Bool("d", false, "debug print")
	//msgLen    = flag.Int("l", )
)

func oneWsClient() {
	var cliID string

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	//fmt.Printf("connecting to %s\n", u.String())

	dial := websocket.Dialer{}
	dial.Proxy = http.ProxyFromEnvironment
	dial.HandshakeTimeout = 45 * time.Second
	dial.ReadBufferSize = 2048
	dial.WriteBufferSize = 2048
	c, _, err := dial.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	groupList := []string{"group_1"}

	sendGroupMsgReq := SendGroupMsgReq{}
	sendGroupMsgReq.Kind = OpSendGroupMsgReq
	sendGroupMsgReq.ToGroup = groupList[rand.Intn(len(groupList))]

	jr := JoinGroupReq{}
	jr.Kind = OpJoinGroupReq
	jr.GroupName = sendGroupMsgReq.ToGroup
	c.WriteJSON(jr)

	var muTR sync.RWMutex
	timeRecord := make(map[uint64]time.Time)

	go func() {
		defer func() {

		}()

		for {
			time.Sleep(time.Second * time.Duration(rand.Intn(10)+2))

			seq := rand.Uint64()
			sendGroupMsgReq.Seq = seq
			sendGroupMsgReq.Content = []byte(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("a"), 10)))

			if *debugMode {
				muTR.Lock()
				timeRecord[seq] = time.Now()
				muTR.Unlock()
			}

			if err := c.WriteJSON(sendGroupMsgReq); err != nil {
				return
			}
			if *debugMode {
				fmt.Println("send seq:", sendGroupMsgReq.Seq)
			}
		}
	}()

	go func() {
		for {
			_, jsonData, err := c.ReadMessage()
			if err != nil {
				return
			}

			recvT := time.Now()

			m := Msg{}
			if err := json.Unmarshal(jsonData, &m); err != nil {
				fmt.Println(err)
				return
			}

			if m.Kind == OpMsgNotify {
				rsp := MsgNotify{}
				if err := json.Unmarshal(jsonData, &rsp); err != nil {
					fmt.Println(err)
					return
				}
				//fmt.Println(rsp)

				if rsp.From == cliID && *debugMode {
					muTR.RLock()
					sendT, ok := timeRecord[rsp.Seq]
					muTR.RUnlock()
					if ok {
						fmt.Println("recv seq:", rsp.Seq)
						//						fmt.Printf("SendTime:%v RecvTime:%v diff: %d\n",
						//							sendT, recvT, recvT.Sub(sendT).Nanoseconds())

						h := metrics.GetOrRegisterHistogram("MsgNotifyDelay", metrics.DefaultRegistry,
							metrics.NewExpDecaySample(1028, 0.015))
						h.Update(recvT.Sub(sendT).Nanoseconds())
					} else {
						fmt.Print("ERROR: seq bad", rsp.Seq)
					}
				}
			} else if m.Kind == OpLoginRsp {
				rsp := LoginRsp{}
				if err := json.Unmarshal(jsonData, &rsp); err != nil {
					fmt.Println(err)
					return
				}
				//fmt.Printf("recv LoginRsp %+v\n", rsp)
				cliID = rsp.Name
			}
		}
	}()
}

func main() {
	flag.Parse()

	if *debugMode {
		go metrics.Log(metrics.DefaultRegistry, time.Second*10, log.New(os.Stderr, "m", log.LstdFlags))
	}

	rand.Seed(time.Now().Unix())

	var i int64
	for ; i < *number; i++ {
		go oneWsClient()
		if i%8 == 0 {
			time.Sleep(time.Millisecond * 200)
		}
	}

	select {}
}
