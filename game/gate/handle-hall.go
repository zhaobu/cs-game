package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/net"
	"cy/game/pb/common"
	"cy/game/pb/hall"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"

	"github.com/astaxie/beego/httplib"
)

func (s *session) handleHall(msg *codec.Message) error {

	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		return err
	}
	tlog.Info("收到大厅消息", zap.Any("err", err))
	switch v := pb.(type) {
	case *pbhall.QueryGameListReq:
		s.handleHallQueryGameListReq(v)
	case *pbhall.QuerySessionInfoReq:
		s.handleHallQuerySessionInfoReq(v)
	case *pbhall.QueryUserInfoReq:
		s.handleHallQueryUserInfoReq(v)
	case *pbhall.QueryUserOwnDeskReq:
		s.handleHallQueryUserOwnDeskReq(v)
	case *pbhall.UpdateBindMobileReq:
		s.handleHallUpdateBindMobileReq(msg.UserID, v)
	case *pbhall.BindXianLiaoAccountReq:
		s.handleHallBindXianLiaoAccountReq(msg.UserID, v)
	case *pbhall.UpdateIdCardReq:
		s.handleHallUpdateIdCardReq(msg.UserID, v)
	case *pbhall.QueryUserBuildInfoReq:
		s.handleHallQueryUserBuildInfoReq(msg.UserID,v)
	default:
		return fmt.Errorf("bad type:%+v", v)
	}
	return nil
}

func (s *session) handleHallQueryGameListReq(req *pbhall.QueryGameListReq) {
	rsp := &pbhall.QueryGameListRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	rsp.GameNames, _ = queryGameList()
}

func (s *session) handleHallQuerySessionInfoReq(req *pbhall.QuerySessionInfoReq) {
	rsp := &pbhall.QuerySessionInfoRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	rsp.Info, _ = cache.QuerySessionInfo(s.uid)
}

func (s *session) handleHallQueryUserInfoReq(req *pbhall.QueryUserInfoReq) {
	rsp := &pbhall.QueryUserInfoRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	rsp.Info, _ = mgo.QueryUserInfo(s.uid)
}

func (s *session) handleHallQueryUserOwnDeskReq(req *pbhall.QueryUserOwnDeskReq) {
	// TODO
}

func (s *session) handleHallUpdateIdCardReq(uid uint64, req *pbhall.UpdateIdCardReq) {
	rsp := &pbhall.UpdateIdCardRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	if req.IdCard == "" || req.Cnname == "" {
		rsp.Code = 2
		return
	}

	status, msg, err := checkIDCard(req.IdCard, req.Cnname)
	if err != nil {
		rsp.Code = 3
		rsp.CodeStr = err.Error()
		return
	}

	if status != "01" {
		rsp.Code = 4
		rsp.CodeStr = msg
		return
	}

	_, err = mgo.UpdateIDCardAndName(uid, req.IdCard, req.Cnname)
	if err != nil {
		rsp.Code = 5
		return
	}

	rsp.Code = 1
}

func checkIDCard(id, name string) (status, msg string, err error) {
	r := httplib.Get(`https://idcert.market.alicloudapi.com/idcard`)
	r.SetTimeout(time.Second*2, time.Second*2)
	r.Param(`idCard`, id)
	r.Param(`name`, name)
	r.Header(`Authorization`, `APPCODE `+*aliAppCode)

	var str string
	str, err = r.String()
	if err != nil {
		return
	}

	var data struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	}
	err = json.Unmarshal([]byte(str), &data)
	return data.Status, data.Msg, err
}

func (s *session) handleHallUpdateBindMobileReq(userID uint64, req *pbhall.UpdateBindMobileReq) {
	rsp := &pbhall.UpdateBindMobileRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	if req.Mobile == "" || req.Captcha == "" || req.Password == "" {
		rsp.Code = 2
		return
	}

	captcha, err := cache.GetCaptcha(req.Mobile)
	if err != nil || captcha != req.Captcha {
		rsp.Code = 3
		return
	}

	cache.DeleteCaptcha(req.Mobile)

	isLoginSucced := (userID != 0)

	userInfo, err := mgo.QueryUserByMobile(req.Mobile)
	if !isLoginSucced {
		if err != nil {
			rsp.Code = 3
			return
		}
		userID = userInfo.UserID
	} else {
		if err == nil && userInfo.UserID != userID {
			rsp.Code = 4
			return
		}
	}
	_, err = mgo.UpdateBindMobile(userID, req.Mobile, req.Password)
	if err != nil {
		rsp.Code = 5
		return
	}
	rsp.Code = 1
	//推送net
	//tlog.Info("推送用户绑定手机信息", zap.Any("Uid", userID), zap.Any("Mobile", req.Mobile))
	go func() {
		err = net.PushUserBindPhone(userID,req.Mobile)
		if err != nil{
			tlog.Error("推送用户绑定手机信息 错误", zap.Any("Uid", userID), zap.Any("Mobile", req.Mobile),zap.Any("err",err.Error()))
		}
	}()
}

//绑定闲聊账号
func (s *session) handleHallBindXianLiaoAccountReq(userID uint64,req *pbhall.BindXianLiaoAccountReq){
	tlog.Info("收到用户绑定闲聊信息", zap.Any("userID", userID), zap.Any("req", req))
	rsp := &pbhall.BindXianLiaoAccountRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()

	if req.XianLiaoId == "" {
		rsp.Code = 2
		return
	}
	isLoginSucced := (userID != 0)

	if !isLoginSucced {
		rsp.Code = 3
		return
	} else {
		userInfo, err := mgo.QueryUserByXianLiao(req.XianLiaoId)
		if err == nil && userInfo.UserID != userID {
			rsp.Code = 4
			return
		}
	}
	_, err := mgo.BindXianLiaoID(userID, req.XianLiaoId)
	if err != nil {
		rsp.Code = 5
		return
	}
	rsp.Code = 1
}

func queryGameList() (gamelist []string, err error) {
	// "http://192.168.0.90:8500/v1/kv/cy_game/game"
	url := fmt.Sprintf("http://%s/v1/kv%s/game", *consulAddr, *basePath)

	req := httplib.Get(url)
	req.Param("recurse", "true")
	req.Param("keys", "")
	body, err := req.String()
	if err != nil {
		return nil, err
	}

	var jsonB []string
	err = json.Unmarshal([]byte(body), &jsonB)
	if err != nil {
		return nil, err
	}

	for _, v := range jsonB {
		ss := strings.Split(v, "/")
		if len(ss) == 4 {
			gamelist = append(gamelist, ss[2])
		}
	}
	return
}

//查询用户绑定信息
func (s *session) handleHallQueryUserBuildInfoReq(uid uint64, req *pbhall.QueryUserBuildInfoReq){
	rsp := &pbhall.QueryUserBuildInfoRsp{
		IsBuildPhone:false,
		PhoneNumber:"",
		IsBuildXianLiao:false,
		XianLiaoAccount:"",
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()
	Info, _ := mgo.QueryUserInfo(s.uid)
	if Info.Mobile != ""{
		rsp.IsBuildPhone = true
		rsp.PhoneNumber = Info.Mobile
	}
	if Info.XLID != ""{
		rsp.IsBuildXianLiao = true
		rsp.XianLiaoAccount = Info.XLID
	}
}
