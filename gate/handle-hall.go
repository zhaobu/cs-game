package main

import (
	"encoding/json"
	"fmt"
	"game/cache"
	"game/codec"
	"game/codec/protobuf"
	"game/db/mgo"
	"game/net"
	pbcommon "game/pb/common"
	pbgame "game/pb/game"
	pbhall "game/pb/hall"
	"strings"
	"time"

	"go.uber.org/zap"

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
		s.handleHallQueryUserBuildInfoReq(msg.UserID, v)
	case *pbhall.QueryUserPointCardInfoReq:
		s.handleHallQueryUserPointCardInfoReq(msg.UserID, v)
	case *pbhall.PointCardExchangeReq:
		s.handleHallPointCardExchangeReq(msg.UserID, v)
	case *pbhall.QueryUserDeskInfosReq:
		s.handleHallQueryUserDeskInfosReq(msg.UserID, v)
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
	udata, _ := mgo.QueryUserInfo(userID)
	IsGive := udata.Mobile == "" //是否赠送
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
	if IsGive {
		go func() {
			err = net.PushUserBindPhone(userID, 1, req.Mobile)
			if err != nil {
				tlog.Error("推送用户绑定手机信息 错误", zap.Any("Uid", userID), zap.Any("Mobile", req.Mobile), zap.Any("err", err.Error()))
			}
		}()
	}
}

//绑定闲聊账号
func (s *session) handleHallBindXianLiaoAccountReq(userID uint64, req *pbhall.BindXianLiaoAccountReq) {
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
	udata, _ := mgo.QueryUserInfo(userID)
	IsGive := udata.XLID == "" //是否赠送
	_, err := mgo.BindXianLiaoID(userID, req.XianLiaoId)
	if err != nil {
		rsp.Code = 5
		return
	}
	rsp.Code = 1

	if IsGive {
		go func() {
			err = net.PushUserBindPhone(userID, 2, req.XianLiaoId)
			if err != nil {
				tlog.Error("推送用户绑定手机信息 错误", zap.Any("Uid", userID), zap.Any("Mobile", req.XianLiaoId), zap.Any("err", err.Error()))
			}
		}()
	}
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
func (s *session) handleHallQueryUserBuildInfoReq(uid uint64, req *pbhall.QueryUserBuildInfoReq) {
	rsp := &pbhall.QueryUserBuildInfoRsp{
		IsBuildPhone:    false,
		PhoneNumber:     "",
		IsBuildXianLiao: false,
		XianLiaoAccount: "",
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()
	Info, _ := mgo.QueryUserInfo(s.uid)
	if Info.Mobile != "" {
		rsp.IsBuildPhone = true
		rsp.PhoneNumber = Info.Mobile
	}
	if Info.XLID != "" {
		rsp.IsBuildXianLiao = true
		rsp.XianLiaoAccount = Info.XLID
	}
}

//查询点卡信息
func (s *session) handleHallQueryUserPointCardInfoReq(uid uint64, req *pbhall.QueryUserPointCardInfoReq) {
	rsp := &pbhall.QueryUserPointCardInfoRsp{
		PCards: []*pbhall.PointCardInfo{},
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()
	data, err := mgo.QueryUserPointcard(uid)
	if err == nil {
		for _, v := range data.Pointcards {
			rsp.PCards = append(rsp.PCards, &pbhall.PointCardInfo{
				PcId:          v.PcId,
				BuyTime:       v.Buytime,
				ExchangeNum:   v.ExchangeNum,
				ExchangeState: uint32(v.ExchangeState),
				ExchangeTime:  v.ExchangeTime,
			})
		}
	}
}

//点卡兑换
func (s *session) handleHallPointCardExchangeReq(uid uint64, req *pbhall.PointCardExchangeReq) {
	rsp := &pbhall.PointCardExchangeRsp{
		ErrorCode: 0,
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()
	code, data := mgo.ExchangePointcard(uid, req.PcId)
	rsp.ErrorCode = uint32(code)
	if code == 0 {
		userinfo, err := mgo.UpdateWealth(uid, pbgame.FeeType_FTMasonry, int64(data.ExchangeNum))
		if err == nil {
			wcmsg := &pbhall.UserWealthChange{
				UserID:        userinfo.UserID,
				Gold:          userinfo.Gold,
				GoldChange:    0,
				Masonry:       userinfo.Masonry,
				MasonryChange: int64(data.ExchangeNum),
			}
			defer s.sendPb(wcmsg)
		}
	}
}

//查询用户桌子信息
func (s *session) handleHallQueryUserDeskInfosReq(uid uint64, req *pbhall.QueryUserDeskInfosReq) {
	rsp := &pbhall.QueryUserDeskInfosRsp{
		Desks: []*pbcommon.DeskInfo{},
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	defer func() {
		s.sendPb(rsp)
	}()
	_, data := cache.GetUserDesk(uid)
	for _, v := range data {
		if deskinfo, err := cache.QueryDeskInfo(v); err == nil {
			rsp.Desks = append(rsp.Desks, deskinfo)
		}
	}
}
