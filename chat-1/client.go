package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type client struct {
	id       string
	isWs     bool
	c        net.Conn
	wc       *websocket.Conn
	toCli    chan []byte
	flag     int32
	closeSig chan struct{}
}

func newClient(wc *websocket.Conn) *client {
	cli := client{}
	cli.id = genRandName(30)
	cli.isWs = true
	cli.wc = wc
	cli.toCli = make(chan []byte, 1000)
	cli.closeSig = make(chan struct{}, 0)
	go cli.loopSend()
	return &cli
}

func (c *client) close() {
	if atomic.CompareAndSwapInt32(&c.flag, 0, 1) {
		close(c.closeSig)
	}
}

func (c *client) sends(itf interface{}) error {
	data, err := json.Marshal(itf)
	if err != nil {
		return err
	}
	return c.send(data)
}

func (c *client) send(msg []byte) error {
	select {
	case c.toCli <- msg:
	default:
		return fmt.Errorf("full")
	}
	return nil
}

func (c *client) loopSend() {
	for {
		select {
		case data := <-c.toCli:
			if err := c.wc.WriteMessage(websocket.TextMessage, data); err != nil {
				//fmt.Println(err)
				return
			}
		case <-c.closeSig:
			return
		}
	}
}
