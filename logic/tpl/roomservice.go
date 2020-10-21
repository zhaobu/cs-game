package tpl

import (
	"encoding/json"
	"game/cache"
	"game/codec"
	"game/codec/protobuf"
	pbcommon "game/pb/common"
	pbgame "game/pb/game"
	pbhall "game/pb/hall"
	pbinner "game/pb/inner"
	"game/util"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
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
	HandleGameUserVoiceStatusReq(uid uint64, req *pbgame.GameUserVoiceStatusReq)
	HandleGameCommandReq(uid uint64, req *pbgame.GameCommandReq)
	HandleVoteDestroyDeskReq(uid uint64, req *pbgame.VoteDestroyDeskReq)
	HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp)
	HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp)
	HandleGameAction(uid uint64, req *pbgame.GameAction)
	HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp)
	HandleSitDownReq(uid uint64, req *pbgame.SitDownReq, rsp *pbgame.SitDownRsp)
	HandleMakeDeskReq(uid, clubMasterUid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) bool
	HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp)
	HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp)
	OnOffLine(uid uint64, online bool)
}

type RoomServie struct {
	roomHandle IRoomHandle              //房间请求处理
	rpcHandle  RpcHandle                //rpc请求
	tlog       *zap.Logger              //structured 风格
	log        *zap.SugaredLogger       //printf风格
	Timer      *timingwheel.TimingWheel //定时器
	GameName   string                   //游戏编号
	GameID     string                   //游戏ip+port
	redisCli   *redis.Client
}

func (self *RoomServie) Init(gameName, gameID string, _tlog *zap.Logger, redisAddr string, redisDb int) {
	self.initRedis(redisAddr, redisDb)
	self.GameName = gameName
	self.GameID = gameID
	self.tlog = _tlog
	self.log = _tlog.Sugar()
	self.Timer = timingwheel.NewTimingWheel(time.Second, 60) //一个节点一个定时器
	self.Timer.Start()
	// self.delInvalidDesk()
	go util.RedisXread(redisAddr, redisDb, "inner_broadcast", self.onMessage)
}

func (self *RoomServie) initRedis(redisAddr string, redisDb int) {
	err := cache.Init(redisAddr, redisDb)
	if err != nil {
		panic(err.Error())
	}

	self.redisCli = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",      // no password set
		DB:       redisDb, // use default DB
	})
}

func (self *RoomServie) RegisterHandle(roomhandle IRoomHandle) {
	self.roomHandle = roomhandle
	self.rpcHandle.service = self
}

func (self *RoomServie) GetRpcHandle() *RpcHandle {
	return &self.rpcHandle
}

//ToGateNormal发送消息
func (self *RoomServie) ToGateNormal(pb proto.Message, printLog bool, uids ...uint64) error {
	if len(uids) == 0 {
		return nil
	}

	if printLog { //pbgame.GameNotif消息打印在ToGate里
		// self.tlog.Info("ToGateNormal", zap.Any("uids", uids), zap.String("msgName", proto.MessageName(pb)), zap.String("msgValue", util.PB2JSON(pb, false)))
		self.log.Infof("ToGateNormal uid: %v,msgName: %s,msgValue: %s", uids, proto.MessageName(pb), util.PB2JSON(pb, true))
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

	_, err = util.RedisXadd(self.redisCli, "backend_to_gate", data)

	if err != nil {
		self.log.Error(err.Error())
	}
	return err
}

//ToGate发送游戏消息(只发送类似pbgame_logic包了一层的消息)
func (self *RoomServie) ToGate(pb proto.Message, printLog bool, uids ...uint64) error {
	if len(uids) == 0 {
		return nil
	}
	// self.tlog.Info("ToGate", zap.Any("uids", uids), zap.String("msgName", proto.MessageName(pb)), zap.String("msgValue", util.PB2JSON(pb, false)))
	//pbgame_logic消息打印在房间日志里
	if printLog {
		self.log.Infof("ToGate uid: %v,msgName: %s,msgValue: %s", uids, proto.MessageName(pb), util.PB2JSON(pb, true))
	}
	var err error
	notif := &pbgame.GameNotif{}
	notif.NotifName, notif.NotifValue, err = protobuf.Marshal(pb)
	if err != nil {
		self.tlog.Error("protobuf.Marshal err", zap.Error(err))
		return err
	}
	return self.ToGateNormal(notif, false, uids...)
}

func (self *RoomServie) SendDeskChangeNotif(cid int64, did uint64, changeTyp int32, deskInfo *pbcommon.DeskInfo) {
	//通知房主更新房间列表
	msg := pbhall.PushMasterDeskChangeInfo{DeskID: did, ChangeTyp: changeTyp}
	if changeTyp != 3 { //删除时不需要详细信息
		msg.Desks = deskInfo
	}
	self.ToGateNormal(&msg, true, deskInfo.CreateUserID)

	self.tlog.Info("SendDeskChangeNotif", zap.Int64("cid", cid), zap.Uint64("did", did), zap.Int32("changeTyp", changeTyp))
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
			cache.RedisXadd("inner_broadcast", data)
		}
	}
}

// func (self *RoomServie) delInvalidDesk() {
// 	var delDesks []uint64
// 	for _, key := range cache.SCAN("deskinfo:*", 50) {
// 		var deskID uint64
// 		fmt.Sscanf(key, "deskinfo:%d", &deskID)
// 		deskInfo, err := cache.QueryDeskInfo(deskID)
// 		if err != nil {
// 			continue
// 		}

// 		if deskInfo.GameName != self.GameName || deskInfo.GameID != self.GameID {
// 			continue
// 		}

// 		delDesks = append(delDesks, deskID)
// 	}

// 	for _, key := range cache.SCAN("sessioninfo:*", 50) {
// 		var userID uint64
// 		fmt.Sscanf(key, "sessioninfo:%d", &userID)
// 		sessInfo, err := cache.QuerySessionInfo(userID)
// 		if err != nil {
// 			continue
// 		}

// 		if sessInfo.GameName != self.GameName ||
// 			sessInfo.GameID != self.GameID ||
// 			sessInfo.Status != pbcommon.UserStatus_InGameing {
// 			continue
// 		}

// 		for _, deskID := range delDesks {
// 			if sessInfo.AtDeskID == deskID {
// 				cache.ExitGame(userID, sessInfo.GameName, sessInfo.GameID, deskID)
// 				break
// 			}
// 		}
// 	}

// 	for _, deskID := range delDesks {
// 		cache.DeleteClubDeskRelation(deskID)
// 		cache.DelDeskInfo(deskID)
// 		cache.FreeDeskID(deskID)
// 	}
// }

func (self *RoomServie) onMessage(channel string, msgData []byte) error {
	var m codec.Message
	err := json.Unmarshal(msgData, m)
	if err != nil {
		return err
	}

	pb, err := protobuf.Unmarshal(m.Name, m.Payload)
	if err != nil {
		return err
	}

	switch v := pb.(type) {
	case *pbinner.UserChangeNotif:
		self.roomHandle.OnOffLine(v.UserID, v.Typ == pbinner.UserChangeType_Online)
	}
	return nil
}
