package main

import (
	"context"
	"cy/im/cache"
	"cy/im/codec"
	"fmt"
	"log"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
)

func innerServer() {
	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("Gate", new(gate), "")
	err := s.Serve("tcp", *iaddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Warn(err)
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
	logrus.WithFields(logrus.Fields{
		"Flag":        args.Flag,
		"ToUID":       args.ToUID,
		"PayloadName": args.PayloadName,
		"err":         err,
	}).Info("backend")

	sess := mgr.GetSession(args.ToUID)
	if sess == nil {
		return fmt.Errorf("can not find sess %d", args.ToUID)
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"uid": args.ToUID, "name": args.PayloadName}).Info("will send")
	return sess.send(buf)
}

func (p *gate) BroadCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	logrus.WithFields(logrus.Fields{
		"Flag":        args.Flag,
		"ToUID":       args.ToUID,
		"PayloadName": args.PayloadName,
		"err":         err,
	}).Info("broadcast")

	if !args.Flag.IsBroadCast() {
		logrus.Warn("not broad cast")
		return nil
	}

	buf, err := args.Encode()
	if err != nil {
		return err
	}

	mgr.Iter(func(uid uint64, sess *session) {
		if err := sess.send(buf); err != nil {
			logrus.Warn(err)
		} else {
			logrus.WithFields(logrus.Fields{"uid": sess.uid, "name": args.PayloadName}).Info("will send")
		}
	})

	return nil
}

func (p *gate) MultiCast(ctx context.Context, args *codec.MsgPayload, reply *interface{}) (err error) {
	logrus.WithFields(logrus.Fields{
		"Flag":        args.Flag,
		"ToUID":       args.ToUID,
		"PayloadName": args.PayloadName,
		"err":         err,
	}).Info("multi cast")

	if !args.Flag.IsMultiCast() {
		logrus.Warn("not multi cast")
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
			logrus.Warn(err)
		} else {
			logrus.WithFields(logrus.Fields{"uid": sess.uid, "name": args.PayloadName}).Info("will send")
		}
	}

	return nil
}
