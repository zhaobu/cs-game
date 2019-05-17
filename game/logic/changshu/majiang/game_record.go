package majiang

import (
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"cy/game/util"
	"sort"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

//GameRecord文件写战绩回放

//一个人的总分
type RankCell struct {
	WinTimes   int32 //胜局
	ChairId    int32 //座位号
	TotalScore int32 //累计总分
}

type GameRecord struct {
	TotalInning uint32
	RankInfo    []*RankCell       //总分排行
	UserScore   []map[int32]int32 //战绩流水
	record      *mgo.WirteRecord
}

type GameRecordArgs struct {
	*pbgame_logic.DeskArg
	GameId string
	ClubId int64
}

func (self *GameRecord) Init(args *GameRecordArgs, players []*PlayerInfo) {
	self.RankInfo = make([]*RankCell, args.Args.PlayerCount)
	for i := int32(0); i < args.Args.PlayerCount; i++ {
		self.RankInfo[i] = &RankCell{ChairId: i}
	}
	self.UserScore = make([]map[int32]int32, 0, args.Args.RInfo.LoopCnt)
	self.record = &mgo.WirteRecord{GameId: args.GameId, ClubId: args.ClubId, DeskId: args.DeskID, TotalInning: args.Args.RInfo.LoopCnt}
	//房间号+时间戳生成RoomRecordId
	self.record.RoomRecordId = strconv.FormatUint(args.DeskID, 10) + strconv.FormatInt(time.Now().Unix(), 10)
	self.record.PayType = args.Args.PaymentType
	self.record.RoomRule, _ = proto.Marshal(args.Args)
	self.record.PlayerInfos = make([]*mgo.GamePlayerInfo, 0, len(players))
	for k, v := range players {
		info := &mgo.GamePlayerInfo{UserId: v.BaseInfo.Uid, Name: v.BaseInfo.Nickname, InitScore: 0, Score: v.BalanceInfo.Point, ChairId: int32(k)}
		self.record.PlayerInfos = append(self.record.PlayerInfos, info)
	}
}

//每局重置
func (self *GameRecord) Reset(curinning uint32) {
	for _, v := range self.record.PlayerInfos {
		v.Score = 0
	}
	self.record.Index = curinning
	self.record.GameStartTime = 0
	self.record.GameEndTime = 0
	self.record.RePlayData = nil
}

//记录游戏战绩
func (self *GameRecord) AddGameRecord(info map[int32]int32) {
	self.UserScore = append(self.UserScore, info)
	for k, v := range info {
		self.RankInfo[k].TotalScore += v
		if v > 0 {
			self.RankInfo[k].WinTimes++
		}
	}
	self.TotalInning++
	//重新排名
	sort.Slice(self.RankInfo, func(i, j int) bool {
		if self.RankInfo[i].TotalScore == self.RankInfo[j].TotalScore {
			return self.RankInfo[i].ChairId < self.RankInfo[j].ChairId
		}
		return self.RankInfo[i].TotalScore > self.RankInfo[j].TotalScore
	})
}

//查询游戏记录
func (self *GameRecord) GetGameRecord() *pbgame_logic.S2CGameRecord {
	msg := &pbgame_logic.S2CGameRecord{TotalInning: self.TotalInning}
	rankInfo := make([]*pbgame_logic.GameRecordRank, 0, len(self.RankInfo))
	for _, v := range self.RankInfo {
		rankInfo = append(rankInfo, &pbgame_logic.GameRecordRank{ChairId: v.ChairId, TotalScore: v.TotalScore, WinTimes: v.WinTimes})
	}
	msg.RankInfo = rankInfo
	gameRecord := &pbgame_logic.Json_GameRecord{}
	gameRecord.UserScore = make([]*pbgame_logic.Json_GameRecord_InningInfo, 0, len(self.UserScore))
	for _, v := range self.UserScore {
		tmp := &pbgame_logic.Json_GameRecord_InningInfo{Score: v}
		gameRecord.UserScore = append(gameRecord.UserScore, tmp)
	}
	msg.JsonRecordInfo = util.PB2JSON(gameRecord, false)
	return msg
}

//记录游戏动作
func (self *GameRecord) RecordGameAction(msg proto.Message) {
	switch msg.(type) {
	case *pbgame_logic.S2CStartGame:
		self.record.GameStartTime = time.Now().Unix()
	case *pbgame_logic.BS2CGameEnd:
		self.record.GameEndTime = time.Now().Unix()
	}
	act := &mgo.GameAction{}
	var err error
	act.ActName, act.ActValue, err = protobuf.Marshal(msg)
	if err != nil {
		log.Error("protobuf.Marshal err", zap.Error(err))
	}
	self.record.RePlayData = append(self.record.RePlayData, act)
}

//游戏结束
func (self *GameRecord) RecordGameEnd(players []*PlayerInfo) {
	for k, v := range players {
		self.record.PlayerInfos[k].Score = v.BalanceInfo.Point
	}
	mgo.AddGameRecord(self.record)
}
