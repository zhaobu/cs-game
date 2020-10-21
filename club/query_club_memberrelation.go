package main

import (
	"context"
	"fmt"
	"game/codec"
	"game/db/mgo"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

//查询俱乐部成员关系列表
func (p *club) QueryClubMemberRelationReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbclub.QueryClubMemberRelationReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByIDReq")
		tlog.Error(err.Error())
		return
	}

	rsp := &pbclub.QueryClubMemberRelationRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}
	}()

	cc := getClub(req.ClubID)
	cc.RLock()
	defer cc.RUnlock()
	if cc == nil { //俱乐部不存在
		rsp.Code = 2
		return
	}
	// 权限检查
	permisOK := false
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}
	if !permisOK { //操作用户权限够
		rsp.Code = 3
		return
	}
	if mber, ok := cc.Members[req.UserID]; !ok {
		rsp.Code = 4
		return
	} else {
		for _, v := range mber.Relation {
			cu := mustGetUserOther(v)
			cu.RLock()
			rsp.Members = append(rsp.Members, &pbclub.MemberInfo{
				UserID:   cc.Members[v].UserID,
				Identity: cc.Members[v].Identity,
				Agree:    cc.Members[v].Agree,
				UserName: cu.UserName,
				Profile:  cu.Profile,
			})
			cu.RUnlock()
		}
	}
	rsp.Code = 1
	return
}

//添加俱乐部成员关系
func (p *club) AddClubMemberRelationReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.AddClubMemberRelationReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByIDReq")
		tlog.Error(err.Error())
		return
	}

	rsp := &pbclub.AddClubMemberRelationRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}
	}()

	cc := getClub(req.ClubID)
	cc.Lock()
	defer cc.Unlock()
	defer mgo.SaveClub(cc.Club) //保存俱乐部数据
	if cc == nil {              //俱乐部不存在
		rsp.Code = 2
		return
	}
	// 权限检查
	permisOK := false
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}
	if !permisOK { //操作用户权限够
		rsp.Code = 3
		return
	}
	if req.UserID != req.RelationUserID {
		if mber, ok := cc.Members[req.UserID]; !ok {
			rsp.Code = 4
			return
		} else {
			if bmber, ok := cc.Members[req.RelationUserID]; !ok {
				rsp.Code = 4
				return
			} else {
				for _, v := range mber.Relation {
					if v == bmber.UserID {
						rsp.Code = 5
						return
					}
				}
				mber.Relation = append(mber.Relation, bmber.UserID)
				bmber.Relation = append(bmber.Relation, mber.UserID)
			}
		}
	} else {
		rsp.Code = 6
		return
	}
	rsp.Code = 1
	return
}

//移除俱乐部成员关系
func (p *club) RemoveClubMemberRelationReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.RemoveClubMemberRelationReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByIDReq")
		tlog.Error(err.Error())
		return
	}

	rsp := &pbclub.RemoveClubMemberRelationRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}
	}()

	cc := getClub(req.ClubID)
	cc.Lock()
	defer cc.Unlock()
	defer mgo.SaveClub(cc.Club) //保存俱乐部数据
	if cc == nil {              //俱乐部不存在
		rsp.Code = 2
		return
	}
	// 权限检查
	permisOK := false
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}
	if !permisOK { //操作用户权限够
		rsp.Code = 3
		return
	}
	if mber, ok := cc.Members[req.UserID]; !ok {
		rsp.Code = 4
		return
	} else {
		if bmber, ok := cc.Members[req.RelationUserID]; !ok {
			rsp.Code = 4
			return
		} else {
			iskp1, ispk2 := false, false
			for i, v := range mber.Relation {
				if v == bmber.UserID {
					mber.Relation = append(mber.Relation[:i], mber.Relation[i+1:]...)
					iskp1 = true
					break
				}
			}
			for i, v := range bmber.Relation {
				if v == mber.UserID {
					bmber.Relation = append(bmber.Relation[:i], bmber.Relation[i+1:]...)
					ispk2 = true
				}
			}
			if !iskp1 || !ispk2 {
				rsp.Code = 5
				return
			}
		}
	}
	rsp.Code = 1
	return
}

//校验是否能加入俱乐部房间(判断是否有亲属关系)
func (p *club) CheckCanJoinClubDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.CheckCanJoinClubDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByIDReq")
		tlog.Error(err.Error())
		return
	}

	rsp := &pbclub.CheckCanJoinClubDeskRsp{}
	cc := getClub(req.ClubID)
	if cc == nil { //俱乐部不存在
		rsp.Code = 2
		return
	}
	cc.RLock()
	if mber, ok := cc.Members[req.JoinUserId]; !ok {
		rsp.Code = 3
		return
	} else {
		for _, v := range mber.Relation {
			for _, v1 := range req.DeskUserIds {
				if v == v1 {
					rsp.Code = 4
					return
				}
			}
		}
	}
	cc.RUnlock()
	rsp.Code = 1
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败" + err.Error())
	}
	return
}
