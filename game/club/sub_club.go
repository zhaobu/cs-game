package main

import (
	"context"
	"fmt"
	"game/codec"
	pbclub "game/pb/club"
)

func (p *club) SubClubChange(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.SubClubChange)
	if !ok {
		err = fmt.Errorf("not *pbclub.SubClubChange")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	cu := mustGetUserOther(args.UserID)
	cu.Lock()

	if req.SubOrUn == 1 {
		cu.subFlag = true
		cu.Unlock()
		sendClubList(args.UserID)
	} else if req.SubOrUn == 2 {
		cu.subFlag = false
		cu.Unlock()
	}

	return
}

func sendClubList(uid uint64) {
	ids := getUserJoinedClubs(uid)
	sendPb := &pbclub.ClubList{}

	for id := range ids {
		cc := getClub(id)
		if cc == nil {
			continue
		}

		cc.RLock()
		newBriefInfo := &pbclub.BriefInfo{}
		newBriefInfo.ID = cc.ID
		newBriefInfo.Name = cc.Name
		newBriefInfo.Profile = cc.Profile
		newBriefInfo.MasterUserID = cc.MasterUserID
		newBriefInfo.Identity = cc.Members[uid].Identity
		cc.RUnlock()

		sendPb.List = append(sendPb.List, newBriefInfo)
	}

	toGateNormal(sendPb, uid)
}
