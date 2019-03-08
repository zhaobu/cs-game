package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"cy/game/pb/login"
	"fmt"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
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

	uinfo, err := mgo.QueryUserInfo(uid)
	if err != nil {
		loginRsp.StrCode = err.Error()
		return
	}

	if uinfo.SessionID != sid {
		loginRsp.StrCode = fmt.Sprintf("bad sessionid %s", sid)
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

	logrus.WithFields(logrus.Fields{"req": loginReq}).Info("loginReq")
	defer logrus.WithFields(logrus.Fields{"rsp": loginRsp}).Info("loginReq")

	switch loginReq.LoginType {
	case pblogin.LoginType_WX:
		u := &pbcommon.UserInfo{}
		u.WxID = loginReq.ID // 登陆标示
		// 更新的信息
		u.Longitude = loginReq.Longitude
		u.Latitude = loginReq.Latitude
		u.Name = loginReq.Name
		u.Sex = loginReq.Sex
		u.Profile = loginReq.Profile

		var err error
		loginRsp.User, err = mgo.UpsertUserInfo(u)
		if err != nil {
			logrus.Error(errors.Wrapf(err, "upsertUserInfo %+v", *loginReq))
			loginRsp.Code = pblogin.LoginRspCode_Other
			loginRsp.StrCode = err.Error()
			return
		}

		loginRsp.Code = pblogin.LoginRspCode_Succ
	case pblogin.LoginType_Phone:
		mobile := loginReq.ID
		reqCaptcha := string(loginReq.Password)

		if mobile == "" || reqCaptcha == "" {
			loginRsp.Code = pblogin.LoginRspCode_IdOrPwdFailed
			return
		}

		captcha, err := cache.GetCaptcha(mobile)
		if err != nil || captcha != reqCaptcha {
			loginRsp.Code = pblogin.LoginRspCode_IdOrPwdFailed
			return
		}

		cache.DeleteCaptcha(mobile)

		loginRsp.User, err = mgo.QueryUserByMobile(mobile)
		if err != nil {
			loginRsp.Code = pblogin.LoginRspCode_Other
			loginRsp.StrCode = err.Error()
			return
		}

		loginRsp.Code = pblogin.LoginRspCode_Succ

	default:
		loginRsp.Code = pblogin.LoginRspCode_Other
	}

	if loginRsp.Code == pblogin.LoginRspCode_Succ {
		if rid, err := uuid.NewV4(); err == nil {
			loginRsp.User, _ = mgo.UpdateSessionID(loginRsp.User.UserID, rid.String())
		}
	}

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
	case *pblogin.MobileCaptchaReq:
		s.handleLoginMobileCaptchaReq(v)
	}

	return nil
}

func (s *session) handleLoginKeepAliveReq(req *pblogin.KeepAliveReq) {
	rsp := &pblogin.KeepAliveRsp{}
	s.sendPb(rsp)
}

func (s *session) handleLoginMobileCaptchaReq(req *pblogin.MobileCaptchaReq) {
	rsp := &pblogin.MobileCaptchaRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer s.sendPb(rsp)

	if req.Mobile == "" {
		rsp.Code = 2
		return
	}

	var err error
	_, err = mgo.QueryUserByMobile(req.Mobile) // 提前是已经被绑定了
	if err != nil {
		rsp.Code = 3
		return
	}

	if !mobileCaptchaReqLimit(req.Mobile) {
		rsp.Code = 4
		return
	}

	captcha := generateMobileCaptcha()

	err = cache.MarkCaptcha(req.Mobile, captcha)
	if err != nil {
		rsp.Code = 5
		return
	}

	err = sendMobileCaptcha(req.Mobile, captcha)
	if err != nil {
		rsp.Code = 6
		return
	}

	rsp.Code = 1
	rsp.TestCaptcha = captcha // TODO
}

func mobileCaptchaReqLimit(mobile string) bool {
	// 暂时只做时间上的限制
	_, err := cache.GetCaptcha(mobile)
	if err == nil {
		return false
	}
	return true
}

var (
	digit = []byte("1234567890")
)

func generateMobileCaptcha() string {
	str := ""
	l := len(digit)
	for i := 0; i < 6; i++ {
		str += string(digit[rand.Intn(l)])
	}
	return str
}

func sendMobileCaptcha(mobile string, captcha string) error {
	return nil
}
