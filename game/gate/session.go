package main

import (
	"bufio"
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbinner "cy/game/pb/inner"
	pblogin "cy/game/pb/login"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aperdana/batcher"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
)

type session struct {
	tc net.Conn

	srvConfig *serverConfig

	uid         uint64
	isLoginSucc bool
	sessionID   string
	curSeq      uint64 //
	stopFlag    int32
	stopSig     chan struct{}
	chInput     chan *codec.Message

	bat *batcher.Batcher // TODO
}

func newSession(tc net.Conn, srvConfig *serverConfig) *session {
	s := session{}
	s.tc = tc
	s.srvConfig = srvConfig

	s.stopSig = make(chan struct{}, 0)
	s.chInput = make(chan *codec.Message, 1024)

	s.bat = batcher.New(s.batchOperator, time.Millisecond*100, batcher.SetMaxBatchSize(10))
	s.bat.Listen()

	go s.recv()
	go s.handleInput()

	return &s
}

func (s *session) stop() {
	if atomic.CompareAndSwapInt32(&s.stopFlag, 0, 1) {
		close(s.stopSig)
		s.tc.Close()
		close(s.chInput)
	}
}

func (s *session) repeatLogin(newSess *session) {
	brokenTip := &pbcommon.BrokenTip{Code: 1}
	brokenTip.Msg = fmt.Sprintf("oldaddr:%s newaddr:%s", s.tc.RemoteAddr().String(), newSess.tc.RemoteAddr().String())

	m := &codec.Message{}
	m.Name, m.Payload, _ = protobuf.Marshal(brokenTip)

	p := codec.NewPacket()
	p.Msgs = append(p.Msgs, m)

	p.WriteTo(s.tc)
}

func (s *session) recv() (err error) {
	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		logrus.WithFields(logrus.Fields{
			"err":   err,
			"r":     r,
			"stack": stack,
		}).Error()

		s.stop()
	}()

	rb := bufio.NewReaderSize(s.tc, 2048)
	for {
		if s.srvConfig.readTimeout != 0 {
			s.tc.SetReadDeadline(time.Now().Add(s.srvConfig.readTimeout))
		}

		pkt := codec.NewPacket()
		err = pkt.ReadFrom(rb)
		if err != nil {
			return err
		}

		for _, msg := range pkt.Msgs {
			s.chInput <- msg
		}
	}

	return nil
}

func (s *session) handleInput() (err error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"err":   err,
				"r":     r,
				"stack": string(debug.Stack()),
			}).Error()
		}

		if err != nil {
			brokenTip := &pbcommon.BrokenTip{Msg: err.Error()}

			m := &codec.Message{}
			m.Name, m.Payload, _ = protobuf.Marshal(brokenTip)

			p := codec.NewPacket()
			p.Msgs = append(p.Msgs, m)

			p.WriteTo(s.tc)
		}

		s.stop()
	}()

	var errorTip *pbcommon.ErrorTip

	for {
		select {
		case <-s.stopSig:
			return
		default:
		}

		if errorTip != nil {
			s.sendPb(errorTip)
			errorTip = nil
		}

		msg, ok := <-s.chInput
		if !ok {
			return
		}

		logrus.WithFields(logrus.Fields{"name": msg.Name, "uid": msg.UserID}).Info("recv")

		if !s.isLoginSucc {
			if msg.Name != proto.MessageName((*pblogin.LoginReq)(nil)) {
				errorTip = &pbcommon.ErrorTip{Msg: fmt.Sprintf("msg not pblogin.LoginReq")}
				continue
			}

			pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
			if err != nil {
				errorTip = &pbcommon.ErrorTip{Msg: err.Error()}
				continue
			}

			loginReq, ok := pb.(*pblogin.LoginReq)
			if !ok {
				errorTip = &pbcommon.ErrorTip{Msg: fmt.Sprintf("can not conver to *pblogin.LoginReq")}
				continue
			}

			var loginRsp *pblogin.LoginRsp
			if loginReq.Head != nil && loginReq.Head.UserID != 0 && loginReq.Head.SessionID != "" {
				loginRsp = loginBySessionID(loginReq)
			} else {
				loginRsp = backendLoginReq(loginReq)
			}

			if loginRsp.Code == pblogin.LoginRspCode_Succ {
				s.isLoginSucc = true
				s.uid = loginRsp.User.UserID
				s.sessionID = loginRsp.User.SessionID
				mgr.SetSession(s)
				// TODO 通知其他gate进程
			}

			s.beforeLoginRsp()
			s.sendPb(loginRsp) // 客户端要求这个放在末尾。。。
			s.afterLoginRsp()
		} else {
			s.dispatch(msg)
		}
	}

	return nil
}

func (s *session) beforeLoginRsp() {
	if !s.isLoginSucc {
		return
	}

	sessInfo, err := cache.QuerySessionInfo(s.uid)
	if err == nil {
		s.sendPb(sessInfo)
	}
}

func (s *session) afterLoginRsp() {
	if !s.isLoginSucc {
		return
	}

	sessInfo, err := cache.QuerySessionInfo(s.uid)
	if err != nil {
		return
	}

	if sessInfo.Status == pbcommon.UserStatus_InGameing {
		cli, err := getGameCli(sessInfo.GameName)
		if err != nil {
			return
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, "game_id", sessInfo.GameID)

		msg := &codec.Message{}
		msg.UserID = s.uid
		msg.Name, msg.Payload, err = protobuf.Marshal(&pbinner.UserLogin{UserID: msg.UserID})
		if err != nil {
			return
		}
		cli.Call(ctx, "UserLogin", msg, nil)
	}
}

func (s *session) dispatch(msg *codec.Message) {
	msg.UserID = s.uid

	idx := strings.LastIndex(msg.Name, ".")
	if idx == -1 {
		logrus.Warnf("bad msg name %s", msg.Name)
		return
	}
	serviceName := msg.Name[:idx]
	serviceMethod := msg.Name[idx+1:]

	var cli client.XClient
	var err error

	ctx := context.Background()

	if serviceName == "pbcenter" {
		cli = cliCenter

		rsp := &codec.Message{}
		err = cli.Call(ctx, serviceMethod, msg, rsp)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err, "name": msg.Name}).Warn()
			return
		}
		s.sendMsg(rsp)
	} else if serviceName == "pbclub" {
		cli = cliClub
		err = cli.Call(ctx, serviceMethod, msg, nil)
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err, "name": msg.Name}).Warn()
			return
		}
	} else if serviceName == "pbgame" {
		gameName, gameID := s.getGameAddr(msg)
		cli, err = getGameCli(gameName)
		if err != nil {
			s.sendPb(&pbcommon.ErrorTip{Msg: err.Error()})
			return
		}

		ctx = context.WithValue(ctx, "game_id", gameID)

		err = cli.Call(ctx, serviceMethod, msg, nil) // 不用回应
		if err != nil {
			logrus.WithFields(logrus.Fields{"err": err, "name": msg.Name}).Warn()
			return
		}
	} else { // 本地处理
		if serviceName == "pbhall" {
			s.handleHall(msg)
		} else if serviceName == "pblogin" {
			s.handleLogin(msg)
		} else {
			logrus.WithFields(logrus.Fields{"name": serviceName}).Warn("bad serviceName:")
			return
		}
	}
}

func (s *session) sendPb(pb proto.Message) {
	if pb == nil {
		logrus.Warn("sendpb empty")
		return
	}

	m := &codec.Message{}
	var err error
	m.Name, m.Payload, err = protobuf.Marshal(pb)
	if err != nil {
		return
	}

	s.sendMsg(m)
}

func (s *session) sendMsg(msg *codec.Message) {
	if msg.Name == "" {
		return
	}

	logrus.WithFields(logrus.Fields{"uid": s.uid, "name": msg.Name}).Info("send")

	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"r":     r,
				"stack": string(debug.Stack()),
			}).Error()
		}
	}()

	s.bat.Batch(msg)
}

func (s *session) batchOperator(reqs []interface{}) {
	if len(reqs) == 0 {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"r":     r,
				"stack": string(debug.Stack()),
			}).Error()
		}
	}()

	pkt := codec.NewPacket()
	for _, req := range reqs {
		msg, ok := req.(*codec.Message)
		if !ok {
			continue
		}

		pkt.Msgs = append(pkt.Msgs, msg)
	}

	if len(pkt.Msgs) == 0 {
		return
	}

	err := pkt.WriteTo(s.tc)
	if err != nil {
		logrus.Warn(err.Error())
		s.stop()
		return
	}
}
