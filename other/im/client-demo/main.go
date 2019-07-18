package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/common/crypto/dh"
	impb "cy/other/im/pb"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"flag"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	metrics "github.com/rcrowley/go-metrics"
	_ "github.com/satori/go.uuid"
)

var (
	addr       = flag.String("a", "localhost:9876", "tcp Addr")
	addr2      = flag.String("a2", "localhost:9877", "ws Addr")
	needCrypto = flag.Bool("c", false, "need crypto")
	needMAC    = flag.Bool("m", false, "need MAC")
	cnt        = flag.Uint64("n", 1, "client count")
	uid        = flag.Uint64("uid", 1, "UID")
	toUID      = flag.Uint64("touid", 2, "to UID")
	loopCnt    = flag.Int("l", 1, "loop count")
	isWs       = flag.Bool("isws", false, "is websocket")
)

func main() {
	flag.Parse()

	printLoginReq()
	printAddFriendNotifAck()
	// printEnterExitRoom()
	printSendMsgReq()

	//printQueryInbox()
	//unmarsh()
	//return
	//onewsCli(67529)
	return
	//	var i uint64
	//	for ; i < *cnt; i++ {
	//		go oneCli(i, false)
	//		if i%9 == 0 {
	//			time.Sleep(time.Millisecond * 200)
	//		}
	//	}
	if *isWs {
		go onewsCli(*uid)
	} else {
		go oneCli(*uid, false)
	}
	select {}
}

func onewsCli(uid uint64) {
	u := url.URL{Scheme: "wss", Host: *addr2, Path: "/ws"}

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
	fmt.Printf("connecting to %s\n", u.String())

	payload := codec.NewMsgPayload()
	payload.Seq = 0
	payload.FromUID = uid
	payload.ToUID = *toUID
	payload.Flag.SetOneway(false)

	sendMsgReq := &impb.SendMsgReq{}
	sendMsgReq.To = *toUID
	sendMsgReq.From = uid
	sendMsgReq.Content = []byte("aaaaa")

	payload.PayloadName, payload.Payload, err = protobuf.Marshal(sendMsgReq)
	if err != nil {
		fmt.Println(err)
	}
	// payload.PayloadName = []byte(`123456`)
	// payload.Payload = []byte(`abc`)

	enPayload, err := payload.Encode()
	if err != nil {
		fmt.Println(err)
		return
	}

	// enPayload = []byte("abc")

	ctx := &codec.MessageCtx{}
	ctx.NeedCrypto = *needCrypto
	ctx.NeedMAC = *needMAC

	m := codec.NewMessage(ctx)
	m.Payload = enPayload
	m.Encrypto()

	wdata, err := m.Encode()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = c.WriteMessage(websocket.BinaryMessage, wdata)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		messageType, p, err := c.ReadMessage()
		fmt.Println(messageType, p, err)
		if err != nil {
			fmt.Println(err)
			break
		}

		rm := codec.NewMessage(ctx)
		err = rm.Decode(p)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = rm.Decrypto()
		if err != nil {
			fmt.Println(err)
			break
		}

		payload := codec.NewMsgPayload()
		err = payload.Decode(rm.Payload)
		if err != nil {
			return
		}

		fmt.Printf("recv %+v\n", *payload)
	}
}

func oneCli(uid uint64, debug bool) {
	conn, err := net.DialTimeout("tcp", *addr, time.Second*10)
	if err != nil {
		fmt.Println(err)
		return
	}
	if debug {
		fmt.Println(uid, "succ conn ", conn.RemoteAddr())
		go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}

	ctx := &codec.MessageCtx{}
	ctx.NeedCrypto = *needCrypto
	ctx.NeedMAC = *needMAC

	if *needCrypto || *needMAC {
		dhPri, dhPub := dh.GenDHpair()
		otherPubBuf := make([]byte, len(dhPub.Bytes()))

		rb := bufio.NewReaderSize(conn, 1024)
		_, err = io.ReadFull(rb, otherPubBuf)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = conn.Write(dhPub.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}

		otherPub := big.NewInt(0).SetBytes(otherPubBuf)
		dhAgree := dh.CalcAgreeKey(dhPri, otherPub)
		agreeKey := dhAgree.Bytes()
		//fmt.Println(agreeKey)

		ctx.AesKey = agreeKey[:32]
		ctx.AesIV = agreeKey[32 : 32+16]
		ctx.H = hmac.New(sha256.New, agreeKey[32+16:])
	}

	var mu sync.RWMutex
	seqTime := make(map[uint64]time.Time)

	go func(uid uint64) {
		var err error
		defer func() {
			//conn.Close()

			fmt.Println(err)
		}()

		var seq uint64

		for idx := 0; idx < *loopCnt; idx++ {
			//time.Sleep(time.Millisecond * 50)

			payload := codec.NewMsgPayload()
			payload.Seq = seq
			seq++
			payload.FromUID = uid
			payload.ToUID = *toUID
			payload.Flag.SetOneway(false)

			sendMsgReq := &impb.SendMsgReq{}
			sendMsgReq.To = *toUID
			sendMsgReq.From = uid
			sendMsgReq.Content = []byte("abc")

			payload.PayloadName, payload.Payload, err = protobuf.Marshal(sendMsgReq)
			if err != nil {
				fmt.Println(err)
			}
			// payload.PayloadName = []byte(`123456`)
			// payload.Payload = []byte(`abc`)

			enPayload, err := payload.Encode()
			if err != nil {
				fmt.Println(err)
				return
			}

			// enPayload = []byte("abc")

			m := codec.NewMessage(ctx)
			m.Payload = enPayload
			m.Encrypto()

			err = m.WriteTo(conn)
			if err != nil {
				fmt.Println(err)
				return
			}

			if debug {
				mu.Lock()
				seqTime[payload.Seq] = time.Now()
				mu.Unlock()
			}

		}
	}(uid)

	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
			if err != nil {
				fmt.Println(err)
			}
		}()

		rb := bufio.NewReaderSize(conn, 1024)

		for {
			m := codec.NewMessage(ctx)
			err = m.ReadFrom(rb)
			if err != nil {
				return
			}

			m.Decrypto()

			payload := codec.NewMsgPayload()
			err = payload.Decode(m.Payload)
			if err != nil {
				return
			}

			fmt.Printf("recv %+v\n", *payload)

			if debug {
				mu.Lock()
				t, ok := seqTime[payload.Seq]
				mu.Unlock()

				if ok {
					h := metrics.GetOrRegisterHistogram("delay", metrics.DefaultRegistry,
						metrics.NewExpDecaySample(1028, 0.015))

					dis := int64(time.Since(t))
					h.Update(dis)
					fmt.Printf("seq:%d send:%v recv:%v dis:%d\n", payload.Seq, t, time.Now(), dis)
				} else {
					fmt.Printf("seq:%d not find\n", payload.Seq)
				}
			}

		}
	}()

	select {}
}
