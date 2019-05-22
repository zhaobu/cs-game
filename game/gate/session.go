package main

import (
	"bufio"
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbhall "cy/game/pb/hall"
	pbinner "cy/game/pb/inner"
	pblogin "cy/game/pb/login"
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aperdana/batcher"
	"github.com/golang/protobuf/proto"
	"github.com/smallnest/rpcx/client"
	"go.uber.org/zap"
)

type session struct {
	tc net.Conn

	srvConfig *serverConfig

	uid         uint64
	isLoginSucc bool
	curSeq      uint64 //
	stopFlag    int32
	stopSig     chan struct{}
	chInput     chan *codec.Message //读消息缓冲区

	bat *batcher.Batcher //批量发送消息
}

func newSession(tc net.Conn, srvConfig *serverConfig) *session {
	s := session{}
	s.tc = tc
	s.srvConfig = srvConfig

	s.stopSig = make(chan struct{}, 0)
	s.chInput = make(chan *codec.Message, 1024)

	//设置打包方式为超时100ms,或者数据大小超过10 时执行一次s.batchOperator函数,
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

		tlog.Error("recover info", zap.Any("err", err), zap.Any("recover", r), zap.Any("stack", stack))
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
			tlog.Error("recover info", zap.Any("err", err), zap.Any("recover", r), zap.Any("stack", string(debug.Stack())))
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

		s.notifBackendOnline(false)
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

		if msg.Name != "pblogin.KeepAliveReq" {
			tlog.Info("recv client", zap.String("name", msg.Name), zap.Uint64("uid", msg.UserID))
		}

		if !s.isLoginSucc {

			if msg.Name == proto.MessageName((*pblogin.LoginReq)(nil)) {
				pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
				if err != nil {
					errorTip = &pbcommon.ErrorTip{Msg: err.Error()}
					continue
				}

				loginReq, ok := pb.(*pblogin.LoginReq)
				if !ok {
					errorTip = &pbcommon.ErrorTip{Msg: fmt.Sprintf("can not conver to %s", msg.Name)}
					continue
				}

				var loginRsp *pblogin.LoginRsp
				// 优先用userid + sessid
				if loginReq.Head != nil && loginReq.Head.UserID != 0 && loginReq.Head.SessionID != "" {
					loginRsp = loginBySessionID(loginReq)
				} else {
					loginRsp = backendLoginReq(loginReq)
				}

				if loginRsp.Code == pblogin.LoginRspCode_Succ {
					s.isLoginSucc = true
					s.uid = loginRsp.User.UserID
					mgr.SetSession(s)
					// TODO 通知其他gate进程
				}

				s.afterLoginRsp(loginRsp)

			} else if msg.Name == proto.MessageName((*pblogin.MobileCaptchaReq)(nil)) ||
				msg.Name == proto.MessageName((*pbhall.UpdateBindMobileReq)(nil)) {
				s.dispatch(msg)
			} else {
				errorTip = &pbcommon.ErrorTip{Msg: fmt.Sprintf("bad msg order")}
				continue
			}
		} else {
			s.dispatch(msg)
		}
	}

	return nil
}

func (s *session) afterLoginRsp(loginRsp *pblogin.LoginRsp) {
	if s.isLoginSucc {
		sessInfo, err := cache.QuerySessionInfo(s.uid)
		if err == nil {
			s.sendPb(sessInfo) // 客户端要求这个顺序 我也没办法 1
		}
	}

	s.sendPb(loginRsp) // 客户端要求这个顺序 我也没办法 2

	if s.isLoginSucc {
		s.notifBackendOnline(true)
	}
}

func (s *session) notifBackendOnline(online bool) {
	tlog.Debug("notifBackendOnline", zap.Uint64("uid", s.uid), zap.Bool("obline", online))
	m := &codec.Message{}
	ucn := &pbinner.UserChangeNotif{
		UserID: s.uid,
	}
	if online {
		ucn.Typ = pbinner.UserChangeType_Online
	} else {
		ucn.Typ = pbinner.UserChangeType_Offline
	}
	err := codec.Pb2Msg(ucn, m)
	if err == nil {
		data, err := json.Marshal(m)
		if err == nil {
			cache.Pub("inner_broadcast", data)
		}
	}
}

func (s *session) dispatch(msg *codec.Message) {
	msg.UserID = s.uid

	idx := strings.LastIndex(msg.Name, ".")
	if idx == -1 {
		log.Warnf("bad msg name %s", msg.Name)
		return
	}
	serviceName := msg.Name[:idx]
	serviceMethod := msg.Name[idx+1:]

	var cli client.XClient
	var err error

	ctx := context.Background()

	if serviceName == "pbcenter" || serviceName == "pbgamerecord" {
		cli = cliCenter

		rsp := &codec.Message{}
		err = cli.Call(ctx, serviceMethod, msg, rsp)
		if err != nil {
			tlog.Error("pbcenter call err", zap.String("name", msg.Name), zap.Any("err", err))
			return
		}
		s.sendMsg(rsp)
	} else if serviceName == "pbclub" {
		cli = cliClub
		err = cli.Call(ctx, serviceMethod, msg, nil)
		if err != nil {
			tlog.Error("pbclub call err", zap.String("name", msg.Name), zap.Any("err", err))
			return
		}
	} else if serviceName == "pbgame" {
		gameName, gameID := s.getGameAddr(msg)
		cli, err = getGameCli(gameName)
		if err != nil {
			tlog.Error("pbgame getGameCli err", zap.String("name", msg.Name), zap.Any("err", err))
			s.sendPb(&pbcommon.ErrorTip{Msg: err.Error()})
			return
		}

		ctx = context.WithValue(ctx, "game_id", gameID)

		err = cli.Call(ctx, serviceMethod, msg, nil) // 不用回应
		if err != nil {
			tlog.Error("pbgame call err", zap.String("name", msg.Name), zap.Any("err", err))
			return
		}
	} else { // 本地处理
		if serviceName == "pbhall" {
			s.handleHall(msg)
		} else if serviceName == "pblogin" {
			s.handleLogin(msg)
		} else {
			tlog.Error("bad serviceName", zap.String("name", serviceName))
			return
		}
	}
}

func (s *session) sendPb(pb proto.Message) {
	if pb == nil {
		log.Warn("sendpb empty")
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
	if msg.Name != "pblogin.KeepAliveRsp" {
		tlog.Info("send client", zap.Uint64("uid", s.uid), zap.String("name", msg.Name))
	}
	defer func() {
		if r := recover(); r != nil {
			tlog.Error("recover info", zap.Any("recover", r), zap.String("stack", string(debug.Stack())))
		}
	}()

	s.bat.Batch(msg)
}

//打包缓冲区满后执行的操作
func (s *session) batchOperator(reqs []interface{}) {
	if len(reqs) == 0 {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tlog.Error("recover info", zap.Any("recover", r), zap.String("stack", string(debug.Stack())))
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
		log.Warn(err.Error())
		s.stop()
		return
	}
}
