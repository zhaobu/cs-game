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

	self.record = &mgo.WirteRecord{CreateInfo: &mgo.WriteGameConfig{GameId: args.GameId, ClubId: args.ClubId, DeskId: args.DeskID, TotalInning: args.Args.RInfo.LoopCnt}}
	self.record.CurGameInfo = &mgo.WriteGameCell{}
	//房间号+时间戳生成RoomRecordId
	self.record.CurGameInfo.RoomRecordId = strconv.FormatUint(args.DeskID, 10) + strconv.FormatInt(time.Now().Unix(), 10)
	self.record.CreateInfo.PayType = args.Args.PaymentType
	tmp := &mgo.GameAction{}
	tmp.ActName, tmp.ActValue, _ = protobuf.Marshal(args.DeskArg)
	self.record.CreateInfo.RoomRule = tmp
	self.record.CurGameInfo.GamePlayers = make([]*mgo.RoomPlayerInfo, 0, len(players))
	for k, v := range players {
		info := &mgo.RoomPlayerInfo{UserId: v.BaseInfo.Uid, Name: v.BaseInfo.Nickname, Score: 0, TotalScore: 0, ChairId: int32(k), Profile: v.BaseInfo.Profile}
		self.record.CurGameInfo.GamePlayers = append(self.record.CurGameInfo.GamePlayers, info)
	}
}

//每局重置
func (self *GameRecord) Reset(curinning uint32) {
	for _, v := range self.record.CurGameInfo.GamePlayers {
		v.Score = 0
	}
	self.record.CurGameInfo.Index = curinning
	self.record.CurGameInfo.GameStartTime = 0
	self.record.CurGameInfo.GameEndTime = 0
	self.record.CurGameInfo.RePlayData = nil
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
		self.record.CurGameInfo.GameStartTime = time.Now().Unix()
	case *pbgame_logic.BS2CGameEnd:
		self.record.CurGameInfo.GameEndTime = time.Now().Unix()
	}
	act := &mgo.GameAction{}
	var err error
	act.ActName, act.ActValue, err = protobuf.Marshal(msg)
	if err != nil {
		log.Error("protobuf.Marshal err", zap.Error(err))
	}
	self.record.CurGameInfo.RePlayData = append(self.record.CurGameInfo.RePlayData, act)
}

//游戏结束
func (self *GameRecord) RecordGameEnd(players []*PlayerInfo) {
	for k, v := range players {
		self.record.CurGameInfo.GamePlayers[k].Score = v.BalanceInfo.Point
		self.record.CurGameInfo.GamePlayers[k].TotalScore = v.BalanceResult.Point
	}
	mgo.AddGameRecord(self.record)
}
