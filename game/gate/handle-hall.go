package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/pb/common"
	"cy/game/pb/hall"

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

	rsp.Info, _ = queryUserInfo(s.uid)
}

func (s *session) handleHallQueryUserOwnDeskReq(req *pbhall.QueryUserOwnDeskReq) {
	// TODO
}

func queryGameList() (gamelist []string, err error) {
	// "http://192.168.1.128:8500/v1/kv/cy_game/game"
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
