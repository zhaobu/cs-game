package main

import (
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/pb/common"
	"cy/game/pb/login"
	"cy/game/util"
	"fmt"

	"github.com/dropbox/godropbox/errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func loginBySessionID(loginReq *pblogin.LoginReq) (loginRsp *pblogin.LoginRsp) {
	loginRsp = &pblogin.LoginRsp{}
	loginRsp.Head = &pbcommon.RspHead{Seq: loginReq.Head.Seq}
	loginRsp.Code = pblogin.LoginRspCode_Other
	uid := loginReq.Head.UserID
	sid := loginReq.Head.SessionID

	logrus.WithFields(logrus.Fields{"req": loginReq.Head}).Info("loginBySessionID")
	defer logrus.WithFields(logrus.Fields{"rsp": loginRsp}).Info("loginBySessionID")

	sess, find := mgr.GetSession(uid)
	if !find {
		loginRsp.StrCode = fmt.Sprintf("can not find uid %d", uid)
		return
	}

	if sess.sessionID != sid {
		loginRsp.StrCode = fmt.Sprintf("bad sessionid %s", sid)
		return
	}

	uinfo, err := queryUserInfo(uid)
	if err != nil {
		loginRsp.StrCode = err.Error()
		return
	}

	loginRsp.User = uinfo
	loginRsp.Code = pblogin.LoginRspCode_Succ
	return
}

func backendLoginReq(loginReq *pblogin.LoginReq) (loginRsp *pblogin.LoginRsp) {
	loginRsp = &pblogin.LoginRsp{}
	if loginReq.Head != nil {
		loginRsp.Head = &pbcommon.RspHead{Seq: loginReq.Head.Seq}
	}

	logrus.WithFields(logrus.Fields{"req": loginReq}).Info("backendLoginReq")
	defer logrus.WithFields(logrus.Fields{"rsp": loginRsp}).Info("backendLoginReq")

	switch loginReq.LoginType {
	case pblogin.LoginType_WX:
		u := &pbcommon.UserInfo{}
		u.WxID = loginReq.ID
		u.Longitude = loginReq.Longitude
		u.Latitude = loginReq.Latitude
		u.Name = loginReq.Name
		u.Sex = loginReq.Sex
		u.Profile = loginReq.Profile

		if rid, err := uuid.NewRandom(); err != nil {
			logrus.Errorf(err.Error())
		} else {
			u.SessionID = rid.String()
		}

		var err error
		loginRsp.User, err = upsertUserInfo(u)
		if err != nil {
			logrus.Error(errors.Wrapf(err, "upsertUserInfo %+v", *loginReq))
			loginRsp.Code = pblogin.LoginRspCode_Other
			loginRsp.StrCode = err.Error()
			return
		}

		loginRsp.Code = pblogin.LoginRspCode_Succ
	default:
		loginRsp.Code = pblogin.LoginRspCode_Other
	}

	return
}

func upsertUserInfo(u *pbcommon.UserInfo) (*pbcommon.UserInfo, error) {
	coll := mgoSess.DB("").C("userinfo")

	var find = make(bson.M)

	err := coll.Find(bson.M{"wxid": u.WxID}).One(find)
	if err != nil {
		if err == mgo.ErrNotFound {
			var err2 error
			u.UserID, err2 = incUserID()
			if err2 != nil {
				return nil, err2
			}

			// 新玩家初始财富，必须要赋值，不能用客户端传过来的
			u.Gold = 5000
			u.Masonry = 8
			u.GoldPre = 0
			u.MasonryPre = 0
			bs, _ := util.Struct2bson(u)
			return u, coll.Insert(bs)
		}
		return nil, err
	}

	old := &pbcommon.UserInfo{}
	err = util.Bson2struct(find, old)
	if err != nil {
		return nil, err
	}

	// 这几个信息不能更新
	u.UserID = old.UserID
	u.WxID = old.WxID
	u.Gold = old.Gold
	u.Masonry = old.Masonry
	u.GoldPre = old.GoldPre
	u.MasonryPre = old.MasonryPre

	bs, _ := util.Struct2bson(u)
	return u, coll.Update(bson.M{"wxid": u.WxID}, bs)
}

func incUserID() (uint64, error) {
	result := bson.M{}
	_, err := mgoSess.DB("").C("userid").Find(nil).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{"max": int64(1)}},
	}, result)
	if err != nil {
		return 0, err
	}

	r, _ := result["max"].(int64)
	return uint64(r), nil
}

func queryUserInfo(uid uint64) (info *pbcommon.UserInfo, err error) {
	info = &pbcommon.UserInfo{}
	result := bson.M{}
	err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).One(result)
	if err != nil {
		return nil, err
	}
	err = util.Bson2struct(result, info)
	return
}

func (s *session) handleLogin(msg *codec.Message) error {
	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		return err
	}

	switch v := pb.(type) {
	case *pblogin.KeepAliveReq:
		s.handleLoginKeepAliveReq(v)
	}

	return nil
}

func (s *session) handleLoginKeepAliveReq(req *pblogin.KeepAliveReq) {
	rsp := &pblogin.KeepAliveRsp{}
	s.sendPb(rsp)
}
