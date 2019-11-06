package main

import (
	"context"
	"fmt"
	"game/codec"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) SetClubIsProofeReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.SetClubIsProofeReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubNoticeReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.SetClubIsProofeRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubChangeInfo(req.ClubID, clubChangeTypUpdateNoTips, args.UserID)
		}
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 2
		return
	}

	cc.Lock()

	m, find := cc.Members[args.UserID]
	if !find || (m.Identity != identityMaster && m.Identity != identityAdmin) {
		cc.Unlock()
		rsp.Code = 3
		return
	}
	cc.IsProofe = req.IsProofe
	cc.noCommit = true
	desk := cc.desks
	IsAutoCreate := cc.IsAutoCreate
	f := cc.f
	cc.Unlock()

	if req.IsProofe { //打烊
		destorydesks := []*pbcommon.DeskInfo{}
		for _, v := range desk {
			if v.Status == "1" {
				destorydesks = append(destorydesks, v)
			}
		}
		if len(destorydesks) > 0 {
			defer destoryDesk(0, destorydesks[0:]...)
		}
	} else {
		////在查询时 做一下俱乐部桌子校验 防止游戏服务器重启 自动开放俱乐部的桌子不存在的情况
		if IsAutoCreate && f == nil { //自动创建桌子 但是当前不存在桌子
			setting, cid, masterUserID := checkAutoCreate(cc.ID)
			if len(setting) > 0 {
				defer createDesk(setting, cid, masterUserID)
			}
		}
	}
	rsp.Code = 1
	return
}
