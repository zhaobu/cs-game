package tpl

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbgame "cy/game/pb/game"
	pbinner "cy/game/pb/inner"
	"encoding/json"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

const (
	// 结算类型
	FeeTypeGold    = 1 // 金币
	FeeTypeMasonry = 2 // 砖石
	// 桌子类型
	DeskTypeMatch  = 1 // 匹配
	DeskTypeFriend = 2 // 好友、俱乐部
	DeskTypeLadder = 3 // 比赛
)

type (
	GameLogicPlugin interface {
		HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp)
		HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp)
		HandleGameAction(uid uint64, req *pbgame.GameAction)
		HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp)
		HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp)
		HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp)
		HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp)
		RunLongTime(deskID uint64, typ int) bool
	}
)

type RoundTpl struct {
	Tlog      *zap.Logger
	Log       *zap.SugaredLogger
	Timer     *timingwheel.TimingWheel
	plugin    GameLogicPlugin
	gameName  string
	gameID    string
	redisPool *redis.Pool
}

func (t *RoundTpl) SetName(gameName, gameID string) {
	t.gameName = gameName
	t.gameID = gameID
	t.Timer = timingwheel.NewTimingWheel(time.Second, 60) //一个节点一个定时器
	t.Timer.Start()
	t.delInvalidDesk()
	t.checkDeskLongTime()
}

func (t *RoundTpl) InitRedis(redisAddr string, redisDb int) {
	err := cache.Init(redisAddr, redisDb)
	if err != nil {
		panic(err.Error())
	}

	t.redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", redisDb); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

func (t *RoundTpl) SetPlugin(p GameLogicPlugin) {
	t.plugin = p
}

func (t *RoundTpl) ToGateNormal(pb proto.Message, uids ...uint64) error {
	t.Log.Infof("tpl send %v %s %+v", uids, proto.MessageName(pb), pb)

	msg := &codec.Message{}
	err := codec.Pb2Msg(pb, msg)
	if err != nil {
		return err
	}

	var xx struct {
		Msg  *codec.Message
		Uids []uint64
	}
	xx.Msg = msg
	xx.Uids = append(xx.Uids, uids...)

	data, err := json.Marshal(xx)
	if err != nil {
		return err
	}

	rc := t.redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	if err != nil {
		t.Log.Error(err.Error())
	}
	return err
}

func (t *RoundTpl) ToGate(pb proto.Message, uids ...uint64) error {
	var err error
	notif := &pbgame.GameNotif{}
	notif.NotifName, notif.NotifValue, err = protobuf.Marshal(pb)
	if err != nil {
		return err
	}
	return t.ToGateNormal(notif, uids...)
}

func (t *RoundTpl) SendDeskChangeNotif(cid int64, did uint64, changeTyp int32) {
	m := &codec.Message{}
	dcn := &pbinner.DeskChangeNotif{
		ClubID: cid,
		DeskID: did,
		ChangeTyp: changeTyp,
	}
	err := codec.Pb2Msg(dcn, m)
	if err == nil {
		data, err := json.Marshal(m)
		if err == nil {
			cache.Pub("inner_broadcast", data)
		}
	}
}
