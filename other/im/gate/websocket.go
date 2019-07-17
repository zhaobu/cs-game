package main

import (
	"bufio"
	"cy/other/im/codec"
	_ "cy/other/im/friend/pb"
	_ "cy/other/im/pb"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	istls    = flag.Bool("tls", false, "use tls")
	certFile = flag.String("certFile", "215082383820147.pem", "certFile")
	keyFile  = flag.String("keyFile", "215082383820147.key", "keyFile")
)

type wsServer struct {
}

func newWsServer() *wsServer {
	t := &wsServer{}
	return t
}

func (w *wsServer) start() {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ws", w.ws)
	r.GET("/wsjson", w.wsJSON)
	var err error
	if *istls {
		err = r.RunTLS(*wsAddr, *certFile, *keyFile)
	} else {
		err = r.Run(*wsAddr)
	}
	if err != nil {
		fmt.Println("websocket exit with ", err)
	}
}

func (w *wsServer) wsJSON(c *gin.Context) {
	frontend, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		frontend.Close()
	}()

	backend, err := net.DialTimeout("tcp", *addr, time.Second*10)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		backend.Close()
	}()

	ctx := codec.MessageCtx{}
	go func() {
		for {
			var err error
			defer func() {
				frontend.Close()
				backend.Close()
				if err != nil {
					fmt.Println("ws frontend", err)
				}
			}()

			pay := codec.NewMsgPayload()
			err = frontend.ReadJSON(pay)
			if err != nil {
				return
			}

			msg := codec.NewMessage(&ctx)
			msg.Payload, err = pay.Encode()
			if err != nil {
				return
			}

			err = msg.Encrypto()
			if err != nil {
				return
			}

			err = msg.WriteTo(backend)
			if err != nil {
				return
			}
		}
	}()

	rb := bufio.NewReaderSize(backend, 1024)
	for {
		var err error

		msg := codec.NewMessage(&ctx)
		err = msg.ReadFrom(rb)
		if err != nil {
			fmt.Println("ws backend recv:", err)
			return
		}

		err = msg.Decrypto()
		if err != nil {
			fmt.Println("ws backend recv:", err)
			return
		}

		pay := codec.NewMsgPayload()
		err = pay.Decode(msg.Payload)
		if err != nil {
			fmt.Println("ws backend recv:", err)
			return
		}

		err = frontend.WriteJSON(pay)
		if err != nil {
			fmt.Println("ws backend recv:", err)
			return
		}
	}
}

func (w *wsServer) ws(c *gin.Context) {
	frontend, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		frontend.Close()
	}()

	backend, err := net.DialTimeout("tcp", *addr, time.Second*10)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		backend.Close()
	}()

	go func() {
		for {
			var err error
			defer func() {
				frontend.Close()
				backend.Close()
				if err != nil {
					fmt.Println("ws frontend", err)
				}
			}()

			_, p, err := frontend.ReadMessage()
			if err != nil {
				return
			}

			_, err = backend.Write(p)
			if err != nil {
				return
			}
		}
	}()

	rb := bufio.NewReaderSize(backend, 1024)
	rbuf := make([]byte, 1024)
	for {
		var err error
		rn, err := rb.Read(rbuf)
		if err != nil {
			fmt.Println("ws backend recv:", err)
			return
		}

		err = frontend.WriteMessage(websocket.BinaryMessage, rbuf[:rn])
		if err != nil {
			fmt.Println("ws backend write:", err)
			return
		}
	}
}
