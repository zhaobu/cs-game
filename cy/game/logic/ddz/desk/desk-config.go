package desk

import (
	"cy/game/codec/protobuf"

	"cy/game/pb/common"
	"cy/game/pb/game"
	"cy/game/pb/game/ddz"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

var (
	deskCfg struct {
		M *pbgame_ddz.MatchConfig      `json:"match"`
		F *pbgame_ddz.FriendsConfigTpl `json:"friends"`
	}
	muDeskCfg sync.RWMutex
)

func LoadConfig(fn string) error {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &deskCfg)
	if err != nil {
		return err
	}

	for _, v := range deskCfg.M.RoomList {
		v.SeatCnt = seatNumber
	}
	deskCfg.F.Def.SeatCnt = seatNumber

	log.Infof("load cfg: %+v\n", deskCfg)
	return nil
}

func QueryConfig(uid uint64, queryGameConfigReq *pbgame.QueryGameConfigReq) {
	queryGameConfigRsp := &pbgame.QueryGameConfigRsp{}
	if queryGameConfigReq.Head != nil {
		queryGameConfigRsp.Head = &pbcommon.RspHead{Seq: queryGameConfigReq.Head.Seq}
	}

	switch queryGameConfigReq.Type {
	case 1:
		queryGameConfigRsp.Name, queryGameConfigRsp.Value, _ = protobuf.Marshal(queryMatchConfig())
	case 2:
		queryGameConfigRsp.Name, queryGameConfigRsp.Value, _ = protobuf.Marshal(queryFriendsConfigTpl())
	default:
	}

	toGateNormal(log, queryGameConfigRsp, uid)

}

func QueryMatchRoomArg(roomID uint32) *pbgame_ddz.RoomArg {
	muDeskCfg.RLock()
	defer muDeskCfg.RUnlock()

	for _, v := range deskCfg.M.RoomList {
		if roomID == v.RoomId {
			return v
		}
	}
	return nil
}

func queryMatchConfig() *pbgame_ddz.MatchConfig {
	muDeskCfg.RLock()
	defer muDeskCfg.RUnlock()

	return deskCfg.M
}

func queryFriendsConfigTpl() *pbgame_ddz.FriendsConfigTpl {
	muDeskCfg.RLock()
	defer muDeskCfg.RUnlock()

	return deskCfg.F
}

//  检查好友场建桌参数
func checkMakeDeskArg(req *pbgame.MakeDeskReq) (*pbgame_ddz.RoomArg, error) {
	pb, err := protobuf.Unmarshal(req.GameArgMsgName, req.GameArgMsgValue)
	if err != nil {
		return nil, err
	}

	cfg, ok := pb.(*pbgame_ddz.RoomArg)
	if !ok {
		return nil, fmt.Errorf("not *pbgame_ddz.RoomArg")
	}

	tplCfg := queryFriendsConfigTpl()

	if cfg.BaseScore < tplCfg.BaseScoreLow || cfg.BaseScore > tplCfg.BaseScoreHigh {
		return nil, fmt.Errorf("bad BaseScore:%d limit:[%d ~ %d]", cfg.BaseScore, tplCfg.BaseScoreLow, tplCfg.BaseScoreHigh)
	}

	feeTypeOk := false
	for _, v := range tplCfg.FeeType {
		if v.T == cfg.FeeType {
			feeTypeOk = true
			break
		}
	}
	if !feeTypeOk {
		return nil, fmt.Errorf("bad FeeType %d", cfg.FeeType)
	}

	paymentTypeOk := false
	for _, v := range tplCfg.PaymentType {
		if v.T == cfg.PaymentType {
			paymentTypeOk = true
			break
		}
	}
	if !paymentTypeOk {
		return nil, fmt.Errorf("bad PaymentType %d", cfg.PaymentType)
	}

	feeOk := false
	for _, v := range tplCfg.RInfo {
		if v.LoopCnt == cfg.LoopCnt && v.Fee == cfg.Fee {
			feeOk = true
			break
		}
	}
	if !feeOk {
		return nil, fmt.Errorf("bad LoopCnt:%d Fee:%d", cfg.LoopCnt, cfg.Fee)
	}

	return cfg, nil
}
