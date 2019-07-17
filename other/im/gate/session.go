package main

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"cy/other/im/crypto/dh"
	"cy/other/im/cache"
	"cy/other/im/codec"
	_ "cy/other/im/friend/pb"
	impb "cy/other/im/pb"
	"fmt"
	"io"
	"math/big"
	"net"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/smallnest/rpcx/client"
	"go.uber.org/zap"
)

type session struct {
	c         net.Conn
	srvConfig *serverConfig
	stopFlag  int32
	stopSig   chan struct{}
	outputCh  chan *codec.Message
	ctx       *codec.MessageCtx
	uid       uint64
	curSeq    uint64

	lastRecvTime time.Time
}

func newSessionTCP(nc net.Conn, srvConfig *serverConfig) *session {
	s := session{}
	s.c = nc
	s.srvConfig = srvConfig
	s.stopSig = make(chan struct{}, 0)
	s.outputCh = make(chan *codec.Message, 1000)
	s.ctx = &codec.MessageCtx{}
	s.ctx.NeedCrypto = srvConfig.needCrypto
	s.ctx.NeedMAC = srvConfig.needMAC

	go s.run()
	go s.doSend()

	return &s
}

func (s *session) stop() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, 0, 1) {
		s.c.Close()
		close(s.stopSig)
	}
}

func (s *session) run() {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("recover info:%s,stack:%s", r, string(debug.Stack()))
		}

		if err != nil {
			log.Errorf("session close:%v", err)
		}
	}()

	if s.srvConfig.needCrypto || s.srvConfig.needMAC {
		//
		tNow := time.Now()
		s.c.SetReadDeadline(tNow.Add(time.Second * 5))
		s.c.SetWriteDeadline(tNow.Add(time.Second * 5))

		err = s.handshake()
		if err != nil {
			return
		}
	}

	err = s.handleRecv()
}

func (s *session) handshake() error {

	dhPri, dhPub := dh.GenDHpair()
	dhPubKeyLen := len(dhPub.Bytes())

	chwr := make(chan error, 1)

	go func() {
		_, err := s.c.Write(dhPub.Bytes())
		chwr <- err
	}()

	rb := bufio.NewReaderSize(s.c, 1024)
	otherPubBuf := make([]byte, dhPubKeyLen)
	_, err := io.ReadFull(rb, otherPubBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if werr := <-chwr; werr != nil {
		fmt.Println(err)
		return werr
	}

	// calc agree key
	otherPub := big.NewInt(0).SetBytes(otherPubBuf)
	dhAgree := dh.CalcAgreeKey(dhPri, otherPub)
	agreeKey := dhAgree.Bytes()

	s.genKey(agreeKey)
	return nil
}

func (s *session) genKey(agreeKey []byte) {
	s.ctx.AesKey = agreeKey[:32]
	s.ctx.AesIV = agreeKey[32 : 32+16]
	s.ctx.H = hmac.New(sha256.New, agreeKey[32+16:])
	return
}

func (s *session) handleRecv() (err error) {
	defer func() {
		tlog.Info("session close", zap.Uint64("uid", s.uid), zap.Error(err))

		if s.uid != 0 {
			placeChange(s.uid, false, s.srvConfig.id)
			// TODO danger
			// mgr.DelSessionByUID(s.uid)
		}
		s.c.Close()

		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			log.Errorf("recover info:stack:%s", string(debug.Stack()))
		}
	}()

	rb := bufio.NewReaderSize(s.c, 1024)

	ifFirstPkg := true
	for {
		if s.srvConfig.readTimeout != 0 {
			s.c.SetReadDeadline(time.Now().Add(s.srvConfig.readTimeout))
		}

		m := codec.NewMessage(s.ctx)
		err = m.ReadFrom(rb)
		if err != nil {
			return err
		}

		err = m.Decrypto()
		if err != nil {
			return err
		}

		args := codec.NewMsgPayload()
		err = args.Decode(m.Payload)
		if err != nil {
			return err
		}

		// if args.Seq != s.curSeq {
		// 	return fmt.Errorf("bad seq get(%d) want(%d)", args.Seq, s.curSeq)
		// }
		// s.curSeq++

		if ifFirstPkg {
			// 禁止特殊ID的玩家
			const broadCastID = 100
			if args.FromUID == 0 || args.FromUID == broadCastID {
				err = fmt.Errorf("bad fromuid %d", args.FromUID)
				return
			}

			if args.PayloadName != proto.MessageName(&impb.LoginReq{}) {
				continue
			}

			s.call(args)

			ifFirstPkg = false

			s.uid = args.FromUID

			// TODO 通知其他gate 等待断开
			mgr.SetSession(s)

			placeChange(s.uid, true, s.srvConfig.id)

			continue
		}

		//
		if args.FromUID != s.uid {
			return fmt.Errorf("bad fromuid want %d get %d", s.uid, args.FromUID)
		}

		// TODO
		// if args.IsHeartbeat() {
		// 	continue
		// }

		if s.lastRecvTime.IsZero() {
			s.lastRecvTime = time.Now()
		} else {
			n := time.Now()
			if n.Sub(s.lastRecvTime) < time.Millisecond*500 {
				continue
			}
			s.lastRecvTime = time.Now()
		}

		s.call(args)
	}
}

func (s *session) call(args *codec.MsgPayload) {
	idx := strings.LastIndex(args.PayloadName, ".")
	serviceName := args.PayloadName[:idx]
	serviceMethod := args.PayloadName[idx+1:]

	var cli client.XClient
	if serviceName == "impb" {
		cli = cliLogic
	} else if serviceName == "friendpb" {
		cli = cliFriend
	} else {
		return
	}

	if args.IsOneway() {
		cli.Call(context.Background(), serviceMethod, args, nil)
	} else {
		reply := codec.NewMsgPayload()
		callErr := cli.Call(context.Background(), serviceMethod, args, reply)
		if callErr != nil {
			log.Warn(callErr.Error())
		} else {
			if reply.PayloadName != "" {
				replyBuf, err := reply.Encode()
				if err == nil {
					tlog.Info("will send", zap.Uint64("uid", s.uid), zap.String("name", reply.PayloadName))
					err = s.send(replyBuf)
				}
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}
}

func placeChange(uid uint64, online bool, gateID string) {
	var err error
	if online {
		err = cache.UserOnline(uid, gateID)
	} else {
		err = cache.UserOffline(uid)
	}
	if err != nil {
		log.Error(err.Error())
	}
}

func (s *session) send(payload []byte) (err error) {
	m := codec.NewMessage(s.ctx)
	m.Payload = payload
	err = m.Encrypto()
	if err != nil {
		return
	}
	if s.srvConfig.writeTimeout != 0 {
		s.c.SetWriteDeadline(time.Now().Add(s.srvConfig.writeTimeout))
	}

	select {
	case s.outputCh <- m:
	default:
		err = fmt.Errorf("uid:%d output buf full", s.uid)
	}
	return
}

func (s *session) doSend() {
	for {
		select {
		case <-s.stopSig:
			return
		case m, ok := <-s.outputCh:
			if ok {
				if err := m.WriteTo(s.c); err != nil {
					log.Infof("send to %d failed %s", s.uid, err.Error())
				}
			}
		}
	}
}
