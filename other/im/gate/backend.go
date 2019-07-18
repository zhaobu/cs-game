package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	. "cy/other/im/common/logger"
	"fmt"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"go.uber.org/zap"
)

func innerServer() {
	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("Gate", new(gate), "")
	err := s.Serve("tcp", *iaddr)
	if err != nil {
		Tlog.Warn("err", zap.Error(err))
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
		Log.Fatal(err)
	}
	s.Plugins.Add(r)
}

type gate struct {
}

func (p *gate) BackEnd(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	Tlog.Warn("err", zap.Error(err))
	Log.Infof("backend:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)
	sess := mgr.GetSession(args.ToUID)
	if sess == nil {
		return fmt.Errorf("can not find sess %d", args.ToUID)
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}
	Tlog.Info("will send", zap.Uint64("uid", args.ToUID), zap.String("name", args.PayloadName))
	return sess.send(buf)
}

func (p *gate) BroadCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	Log.Infof("broadcast:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)

	if !args.Flag.IsBroadCast() {
		Log.Warn("not broad cast")
		return nil
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}

	mgr.Iter(func(uid uint64, sess *session) {
		if err := sess.send(buf); err != nil {
			Log.Warn(err)
		} else {
			Tlog.Info("will send", zap.Uint64("uid", sess.uid), zap.String("name", args.PayloadName))
		}
	})

	return nil
}

func (p *gate) MultiCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	Log.Infof("multi cast:Flag:%v,ToUID:%d,PayloadName:%s,err:%v", args.Flag, args.ToUID, args.PayloadName, err)

	if !args.Flag.IsMultiCast() {
		Log.Warn("not multi cast")
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
			Log.Warn(err)
		} else {
			Tlog.Info("will send", zap.Uint64("uid", sess.uid), zap.String("name", args.PayloadName))
		}
	}

	return nil
}
