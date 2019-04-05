package tpl

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbinner "cy/game/pb/inner"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

//rpc消息处理接口
// type IRpcHandle interface {
// 	DestroyDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	ExitDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	JoinDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// 	QueryGameConfigReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error)
// }

type IRoomHandle interface {
	HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp)
	HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp)
	HandleGameAction(uid uint64, req *pbgame.GameAction)
	HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp)
	HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp)
	HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp)
	HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp)
	RunLongTime(deskID uint64, typ int) bool
}

type RoomServie struct {
	roomHandle IRoomHandle              //房间请求处理
	rpcHandle  RpcHandle                //rpc请求
	tlog       *zap.Logger              //structured 风格
	log        *zap.SugaredLogger       //printf风格
	Timer      *timingwheel.TimingWheel //定时器
	gameName   string                   //游戏编号
	gameID     string                   //游戏ip+port
	redisPool  *redis.Pool
}

func (self *RoomServie) Init(gameName, gameID string, _tlog *zap.Logger, redisAddr string, redisDb int) {
	self.initRedis(redisAddr, redisDb)
	self.gameName = gameName
	self.gameID = gameID
	self.tlog = _tlog
	self.log = _tlog.Sugar()
	self.Timer = timingwheel.NewTimingWheel(time.Second, 60) //一个节点一个定时器
	self.Timer.Start()
	self.delInvalidDesk()
	self.checkDeskLongTime()
}

func (self *RoomServie) initRedis(redisAddr string, redisDb int) {
	err := cache.Init(redisAddr, redisDb)
	if err != nil {
		panic(err.Error())
	}

	self.redisPool = &redis.Pool{
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

func (self *RoomServie) RegisterHandle(roomhandle IRoomHandle) {
	self.roomHandle = roomhandle
	self.rpcHandle.service = self
}

func (self *RoomServie) GetRpcHandle() *RpcHandle {
	return &self.rpcHandle
}

//ToGateNormal发送消息
func (self *RoomServie) ToGateNormal(pb proto.Message, uids ...uint64) error {
	if _, ok := pb.(*pbgame.GameNotif); !ok { //游戏消息在ToGate里打印
		self.tlog.Info("ToGateNormal", zap.Any("uids", uids), zap.String("msgName", proto.MessageName(pb)), zap.Any("msgValue", pb))
	}
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

	rc := self.redisPool.Get()
	defer rc.Close()

	_, err = rc.Do("PUBLISH", "backend_to_gate", data)
	if err != nil {
		self.log.Error(err.Error())
	}
	return err
}

//ToGate发送游戏消息
func (self *RoomServie) ToGate(pb proto.Message, uids ...uint64) error {
	self.tlog.Info("ToGate", zap.Any("uids", uids), zap.String("msgName", proto.MessageName(pb)), zap.Any("msgValue", pb))
	var err error
	notif := &pbgame.GameNotif{}
	notif.NotifName, notif.NotifValue, err = protobuf.Marshal(pb)
	if err != nil {
		return err
	}
	return self.ToGateNormal(notif, uids...)
}

func (self *RoomServie) SendDeskChangeNotif(cid int64, did uint64, changeTyp int32) {
	m := &codec.Message{}
	dcn := &pbinner.DeskChangeNotif{
		ClubID:    cid,
		DeskID:    did,
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

func (self *RoomServie) delInvalidDesk() {
	var delDesks []uint64
	for _, key := range cache.SCAN("deskinfo:*", 50) {
		var deskID uint64
		fmt.Sscanf(key, "deskinfo:%d", &deskID)
		deskInfo, err := cache.QueryDeskInfo(deskID)
		if err != nil {
			continue
		}

		if deskInfo.GameName != self.gameName || deskInfo.GameID != self.gameID {
			continue
		}

		delDesks = append(delDesks, deskID)
	}

	for _, key := range cache.SCAN("sessioninfo:*", 50) {
		var userID uint64
		fmt.Sscanf(key, "sessioninfo:%d", &userID)
		sessInfo, err := cache.QuerySessionInfo(userID)
		if err != nil {
			continue
		}

		if sessInfo.GameName != self.gameName ||
			sessInfo.GameID != self.gameID ||
			sessInfo.Status != pbcommon.UserStatus_InGameing {
			continue
		}

		for _, deskID := range delDesks {
			if sessInfo.AtDeskID == deskID {
				cache.ExitGame(userID, sessInfo.GameName, sessInfo.GameID, deskID)
				break
			}
		}
	}

	for _, deskID := range delDesks {
		cache.DeleteClubDeskRelation(deskID)
		cache.DelDeskInfo(deskID)
		cache.FreeDeskID(deskID)
	}
}

func (self *RoomServie) checkDeskLongTime() {
	go func() {
		time.Sleep(time.Minute * 5)
		now := time.Now().UTC()

		for _, key := range cache.SCAN("deskinfo:*", 50) {
			var deskID uint64
			fmt.Sscanf(key, "deskinfo:%d", &deskID)
			deskInfo, err := cache.QueryDeskInfo(deskID)
			if err != nil {
				continue
			}

			if deskInfo.GameName != self.gameName || deskInfo.GameID != self.gameID {
				continue
			}

			doneOk := false
			du := now.Sub(time.Unix(deskInfo.CreateTime, 0)).Minutes()

			if du > (time.Minute * 60).Minutes() {
				doneOk = self.roomHandle.RunLongTime(deskID, 2)
			} else if du > (time.Minute*30).Minutes() && deskInfo.Status == "" {
				doneOk = self.roomHandle.RunLongTime(deskID, 1)
			}

			if doneOk {
				cache.DeleteClubDeskRelation(deskID)
				cache.DelDeskInfo(deskID)
				cache.FreeDeskID(deskID)
			}
		}
	}()
}
