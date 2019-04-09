package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbcommon "cy/game/pb/common"
	pblogin "cy/game/pb/login"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func loginBySessionID(loginReq *pblogin.LoginReq) (loginRsp *pblogin.LoginRsp) {
	loginRsp = &pblogin.LoginRsp{}
	loginRsp.Head = &pbcommon.RspHead{Seq: loginReq.Head.Seq}
	loginRsp.Code = pblogin.LoginRspCode_Other
	uid := loginReq.Head.UserID
	sid := loginReq.Head.SessionID

	tlog.Info("loginBySessionID", zap.Any("req", loginReq.Head))
	defer tlog.Info("loginBySessionID", zap.Any("rsp", loginRsp))

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

	tlog.Info("loginReq", zap.Any("req", loginReq))
	defer tlog.Info("loginReq", zap.Any("rsp", loginRsp))

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
			log.Error(errors.Wrapf(err, "upsertUserInfo %+v", *loginReq))
			loginRsp.Code = pblogin.LoginRspCode_Other
			loginRsp.StrCode = err.Error()
			return
		}

		loginRsp.Code = pblogin.LoginRspCode_Succ
	case pblogin.LoginType_Phone:
		mobile := loginReq.ID

		if mobile == "" || loginReq.Password == "" {
			loginRsp.Code = pblogin.LoginRspCode_IdOrPwdFailed
			return
		}

		userInfo, err := mgo.QueryUserByMobile(mobile)
		if err != nil {
			loginRsp.Code = pblogin.LoginRspCode_MobileNoBind
			loginRsp.StrCode = err.Error()
			return
		}

		if userInfo.Password != loginReq.Password {
			loginRsp.Code = pblogin.LoginRspCode_IdOrPwdFailed
			return
		}

		loginRsp.User = userInfo
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
		s.handleLoginMobileCaptchaReq(msg.UserID, v)
	}

	return nil
}

func (s *session) handleLoginKeepAliveReq(req *pblogin.KeepAliveReq) {
	rsp := &pblogin.KeepAliveRsp{}
	s.sendPb(rsp)
}

func (s *session) handleLoginMobileCaptchaReq(userID uint64, req *pblogin.MobileCaptchaReq) {
	rsp := &pblogin.MobileCaptchaRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer s.sendPb(rsp)

	if req.Mobile == "" {
		rsp.Code = 2
		return
	}

	if !mobileCaptchaReqLimit(req.Mobile) {
		rsp.Code = 5
		return
	}

	var err error
	var userInfo *pbcommon.UserInfo
	isLoginSucced := userID != 0

	userInfo, err = mgo.QueryUserByMobile(req.Mobile)

	if !isLoginSucced {
		// 登陆时，必须是已经被绑定过的
		if err != nil {
			rsp.Code = 3
			return
		}
	} else {
		// 重置时，必须没有被其他人绑定
		if err == nil && userInfo.UserID != userID {
			rsp.Code = 4
			return
		}
	}

	captcha := generateMobileCaptcha()

	err = cache.MarkCaptcha(req.Mobile, captcha)
	if err != nil {
		rsp.Code = 6
		return
	}

	err = sendMobileCaptcha(req.Mobile, captcha)
	if err != nil {
		rsp.Code = 7
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
	client, err := sdk.NewClientWithAccessKey("default", "LTAIlctikPyE8yy1", "dFrHFwwiWZ8bMlGt63sybVCJ0Su9zg")
	if err != nil {
		return err
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["PhoneNumbers"] = mobile
	request.QueryParams["SignName"] = "三格软件"
	request.QueryParams["TemplateCode"] = "SMS_137655450"
	request.QueryParams["TemplateParam"] = "{\"code\":\"" + captcha + "\"}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		return err
	}

	var xx struct {
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
		BizID     string `json:"BizId"`
		Code      string `json:"Code"`
	}

	rspStr := response.GetHttpContentString()
	err = json.Unmarshal([]byte(rspStr), &xx)
	if err != nil {
		return err
	}
	if xx.Message == "OK" && xx.Code == "OK" {
		return nil
	}
	return fmt.Errorf("aliyun sdk err %v", xx)
}
