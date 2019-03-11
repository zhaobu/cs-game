package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"cy/game/pb/hall"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego/httplib"
)

func (s *session) handleHall(msg *codec.Message) error {
	pb, err := protobuf.Unmarshal(msg.Name, msg.Payload)
	if err != nil {
		return err
	}

	switch v := pb.(type) {
	case *pbhall.QueryGameListReq:
		s.handleHallQueryGameListReq(v)
	case *pbhall.QuerySessionInfoReq:
		s.handleHallQuerySessionInfoReq(v)
	case *pbhall.QueryUserInfoReq:
		s.handleHallQueryUserInfoReq(v)
	case *pbhall.QueryUserOwnDeskReq:
		s.handleHallQueryUserOwnDeskReq(v)
	case *pbhall.QueryMobileIsBindReq:
		s.handleHallQueryMobileIsBindReq(v)
	case *pbhall.UpdateBindMobileReq:
		s.handleHallUpdateBindMobileReq(msg.UserID, v)
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

func (s *session) handleHallQueryMobileIsBindReq(req *pbhall.QueryMobileIsBindReq) {
	rsp := &pbhall.QueryMobileIsBindRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	_, err := mgo.QueryUserByMobile(req.Mobile)
	if err != nil {
		rsp.Code = 2
	} else {
		rsp.Code = 1
	}

	s.sendPb(rsp)
}

func (s *session) handleHallUpdateBindMobileReq(uid uint64, req *pbhall.UpdateBindMobileReq) {
	rsp := &pbhall.UpdateBindMobileRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		s.sendPb(rsp)
	}()

	captcha, err := cache.GetCaptcha(req.Mobile)
	if err != nil || captcha != req.Mobile {
		rsp.Code = 2
		return
	}

	cache.DeleteCaptcha(req.Mobile)

	rsp.Info, err = mgo.UpdateBindMobile(uid, req.Mobile)
	if err != nil {
		rsp.Code = 3
		return
	}

	rsp.Code = 1
}

func queryGameList() (gamelist []string, err error) {
	// "http://localhost:8500/v1/kv/cy_game/game"
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
		if len(ss) == 3 {
			gamelist = append(gamelist, ss[2])
		}
	}
	return
}
