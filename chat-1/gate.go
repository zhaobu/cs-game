package main

import (
	. "cy/chat/def"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"runtime/debug"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Gate struct {
	l           net.Listener
	acceptedCnt int64
}

func NewGate() *Gate {
	g := Gate{}
	return &g
}

func (g *Gate) StartListen(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		panic("err listen")
	}
	g.l = l

	for {
		nc, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}

		go g.handle(nc)
	}
}

func (g *Gate) handle(nc net.Conn) {

}

func (g *Gate) HandleWs(wc *websocket.Conn) {
	atomic.AddInt64(&g.acceptedCnt, 1)

	cli := newClient(wc)

	sendLoginRsp(cli)

	// recv loop
	var err error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Print(r)
			}
			if err != nil {
				debug.PrintStack()
			}

			fmt.Printf("client %s exit\n", cli.id)

			cliDrop(cli.id)
			atomic.AddInt64(&g.acceptedCnt, -1)
		}()

		for {
			_, p, err := wc.ReadMessage()
			if err != nil {
				return
			}

			g.do(p, cli)
		}
	}()
}

func (g *Gate) do(jsonData []byte, cli *client) {
	m := Msg{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		fmt.Println(err)
		return
	}

	t, ok := jsonType[m.Kind]
	if !ok {
		fmt.Println("not supper kind ", m.Kind)
		return
	}

	nv := reflect.New(t.T).Interface()
	if err := json.Unmarshal(jsonData, &nv); err != nil {
		fmt.Println(err)
		return
	}

	t.F(nv, cli)
}
