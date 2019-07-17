package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"fmt"
	"time"

	"go.uber.org/zap"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

func innerServer() {
	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("Gate", new(gate), "")
	err := s.Serve("tcp", *iaddr)
	if err != nil {
		tlog.Warn("err", zap.Error(err))
	}
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.ConsulRegisterPlugin{
		ServiceAddress: "tcp@" + *iaddr,
		ConsulServers:  []string{*consulAddr},
		BasePath:       *basePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}

type gate struct {
}

func (p *gate) BackEnd(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	tlog.Warn("err", zap.Error(err))
	log.Infof("backend:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)

	sess := mgr.GetSession(args.ToUID)
	if sess == nil {
		return fmt.Errorf("can not find sess %d", args.ToUID)
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}
	tlog.Info("will send", zap.Uint64("uid", args.ToUID), zap.String("name", args.PayloadName))
	return sess.send(buf)
}

func (p *gate) BroadCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	log.Infof("broadcast:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)

	if !args.Flag.IsBroadCast() {
		log.Warn("not broad cast")
		return nil
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}

	mgr.Iter(func(uid uint64, sess *session) {
		if err := sess.send(buf); err != nil {
			log.Warn(err)
		} else {
			tlog.Info("will send", zap.Uint64("uid", sess.uid), zap.String("name", args.PayloadName))
		}
	})

	return nil
}

func (p *gate) MultiCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	log.Infof("multi cast:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)

	if !args.Flag.IsMultiCast() {
		log.Warn("not multi cast")
		return nil
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}

	// TODO 优化 一定时间内，取本地成员列表即可
	uids := cache.RoomUsers(args.ToUID)
	for _, uid := range uids {
		sess := mgr.GetSession(uid)
		if sess == nil {
			continue
		}

		if err := sess.send(buf); err != nil {
			log.Warn(err)
		} else {
			tlog.Info("will send", zap.Uint64("uid", sess.uid), zap.String("name", args.PayloadName))
		}
	}

	return nil
}
