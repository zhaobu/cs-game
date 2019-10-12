package main

import (
	"context"
	"game/codec"
	"game/pb/club"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

//校验解散俱乐部房间权限
func (p *club) CheckCanDestoryDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error("解析消息 1 失败"+ err.Error())
		return err
	}
	req, ok := pb.(*pbclub.CheckCanDestoryDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbclub.CheckCanDestoryDeskRsp{
		Error: 0,
	}
	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Error = 3
	}else{
		if  cc.MasterUserID  == req.UserID {
			lasttimer := time.Unix(cc.LastDestoryDeskNumTime,0)
			if 	lasttimer.Day() != time.Now().Day() {
				cc.CurrDayDestoryDeskNum = 0
			}
			if cc.CurrDayDestoryDeskNum < 3 {
				rsp.Error = 0
				cc.CurrDayDestoryDeskNum ++
				cc.LastDestoryDeskNumTime = time.Now().Unix()
			}else{
				rsp.Error = 1
			}
		}else{
			rsp.Error = 2
		}
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	return
}