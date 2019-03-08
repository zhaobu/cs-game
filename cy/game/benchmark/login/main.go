package main

import (
	"crypto/rand"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/pb/common"
	"cy/game/pb/login"
	"flag"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

var (
	beginWxID = flag.Int("b", 1, "beginWxID")
	endWxID   = flag.Int("e", 1000, "endWxID")
	addr      = flag.String("addr", "localhost:19876", "tcp listen address")

	wxIdx int64
)

type cliSt struct {
	c   net.Conn
	uid uint64
}

func (c *cliSt) sendPb(pb proto.Message) {
	var err error
	m := &codec.Message{}
	m.Name, m.Payload, err = protobuf.Marshal(pb)
	if err != nil {
		fmt.Println(err)
		return
	}
	pktReq := codec.NewPacket()
	pktReq.Msgs = append(pktReq.Msgs, m)

	err = pktReq.WriteTo(c.c)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (c *cliSt) detailMsg(msg *codec.Message) bool {
	fmt.Println("recv", msg.Name)

	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if msg.Name == "pblogin.LoginRsp" {
		loginRsp := pb.(*pblogin.LoginRsp)
		fmt.Printf("%v\n", loginRsp)
		return true
	}
	return false
}

func main() {
	flag.Parse()

	beginT := time.Now()

	wg := sync.WaitGroup{}
	for i := *beginWxID; i < *endWxID; i++ {
		wg.Add(1)
		go login(&wg)
		if i%9 == 0 {
			time.Sleep(time.Millisecond * 20)
		}
	}

	wg.Wait()

	fmt.Println(time.Now().Sub(beginT))
}

func login(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	cli := cliSt{}

	var err error
	cli.c, err = net.Dial("tcp4", *addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	name := make([]byte, 30)
	rand.Read(name)

	cli.sendPb(&pblogin.LoginReq{
		Head:      &pbcommon.ReqHead{Seq: 1},
		LoginType: pblogin.LoginType_WX,
		ID:        fmt.Sprintf("wx_%d", atomic.AddInt64(&wxIdx, 1)),
		Name:      string(name),
	})

	// recv loginRsp
	for {
		var err error
		pktRsp := codec.NewPacket()
		err = pktRsp.ReadFrom(cli.c)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, msg := range pktRsp.Msgs {
			if cli.detailMsg(msg) {
				return
			}
		}
	}
}
