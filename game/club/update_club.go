package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) UpdateClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.UpdateClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.UpdateClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubChangeInfo(req.ClubID, clubChangeTypUpdate, args.UserID)
		}
	}()

	if req.Base.Name == "" {
		rsp.Code = 2
		return
	}

	if req.Base.IsAutoCreate && req.Base.IsCustomGameArg {
		req.Base.IsCustomGameArg = false
	}

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}

	cc.Lock()
	if cc.MasterUserID != args.UserID {
		cc.Unlock()
		rsp.Code = 4
		return
	}

	cc.Name = req.Base.Name
	cc.IsAutoCreate = req.Base.IsAutoCreate
	cc.IsCustomGameArg = req.Base.IsCustomGameArg
	cc.IsMasterPay = req.Base.IsMasterPay

	cc.GameArgs = make([]*mgo.DeskSetting, 0)
	for _, v := range req.GameArgs {
		cc.GameArgs = append(cc.GameArgs, &mgo.DeskSetting{
			GameName:        v.GameName,
			GameArgMsgName:  v.GameArgMsgName,
			GameArgMsgValue: v.GameArgMsgValue,
			Enable:          v.Enable,
		})
	}

	cc.noCommit = true

	cc.Unlock()

	rsp.Code = 1


	//更新房间设置时检查是否需要重新创建房间
	if cc.IsAutoCreate && cc.f == nil {		//自动创建桌子 但是当前不存在桌子
		haveEmptyTable := false
		for _,v :=range cc.desks {
			if v.Status == "1" {	//有空桌子
				haveEmptyTable = true
				break;
			}
		}
		if !haveEmptyTable {	//不存在空桌子
			logrus.Infof("查询俱乐部时校验到没有空桌子 %x",cc.desks)
			checkAutoCreate(cc.ID)	//自动创建房间
		}
	}
	//if req.Base.IsAutoCreate {
	//	go checkAutoCreate(req.ClubID)
	//}
	return
}
