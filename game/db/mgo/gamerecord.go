package mgo

import (
	"cy/game/net"
	"strconv"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	RoomRecordTable             = "roomrecord"             //游戏房间记录
	GameRecordTable             = "gamerecord"             //游戏单局记录
	UserCurrDayStatisticsTable  = "usercurrdaystatistics"  //用户当日统计数据
	ClubCurrDayStatisticsTable  = "clubcurrdaystatistics"  //俱乐部当日统计数据
	ClubStatisticsPlayTable     = "clubstatisticsplay"     //俱乐部游戏对局次数统计数据
	ClubStatisticsIntegralTable = "clubstatisticsintegral" //俱乐部游戏积分统计数据
)

//游戏协议数据封装
type GameAction struct {
	ActName  string //消息名称
	ActValue []byte //消息序列化后的值
}

//游戏创建详情
type WriteGameConfig struct {
	GameId      string      //游戏Id
	ClubId      int64       //俱乐部Id
	DeskId      uint64      //房间号
	TotalInning uint32      //总局数
	PayType     uint32      //支付方式1 个人支付 2 AA支付
	DeskInfo    *GameAction //房间规则
	MasterUid   uint64      //房主uid
	Fee         uint32      //扣费
}

//本局数据
type WriteGameCell struct {
	RoomRecordId  string            //游戏房间创建时生成的唯一Id 用户关联该房间
	Index         uint32            //游戏当前局数
	GameStartTime int64             //开始时间 存储时间错
	GameEndTime   int64             //结束时间 存储时间错
	GamePlayers   []*RoomPlayerInfo //本局得分情况
	RePlayData    []*GameAction     //复盘数据
}

//战绩写入
type WirteRecord struct {
	CreateInfo  *WriteGameConfig //游戏房间详情
	CurGameInfo *WriteGameCell   //本局信息
}

//房间记录
type RoomRecord struct {
	RoomRecordId  string            `bson:"_id,omitempty"` //房间记录Id,用来做表主键
	GameStartTime int64             //房间开始时间
	TotalInning   uint32            //总局数
	DeskId        uint64            //房间号
	GameId        string            //游戏Id
	ClubId        int64             //俱乐部Id
	MasterUid     uint64            //房主uid
	PayType       uint32            //支付方式1 个人支付 2 AA支付
	DeskInfo      *GameAction       //房间规则
	Fee           uint32            //扣费
	GamePlayers   []*RoomPlayerInfo //玩家当前总积分详情
	GameRecords   []string          //房间每局游戏记录id
}
type RoomPlayerInfo struct {
	UserId   uint64 `bson:"userid,omitempty"`   //用户Id
	Name     string `bson:"name,omitempty"`     //姓名
	Score    int32  `bson:"score,omitempty"`    //本局得分
	PreScore int32  `bson:"prescore,omitempty"` //当前累计总得分
}

//游戏单局记录
type GameRecord struct {
	GameRecordId  string            `bson:"_id,omitempty"` //游戏记录ID 主键
	Index         uint32            //第几局
	GameStartTime int64             //开始时间 存储时间错
	GameEndTime   int64             //结束时间 存储时间错
	GamePlayers   []*RoomPlayerInfo //本局玩家得分情况
	RePlayData    []*GameAction     //游戏回放数据
}

//俱乐部当日统计数据-----------------------------------------------------------------------------------------------------
type UserCurrDayStatisticsData struct {
	UserId         uint64 `bson:"_id"` //用户Id
	StatisticsPlay int64  //当天本俱乐部次数统计
}
type ClubCurrDayStatisticsData struct {
	ClubId int64 //俱乐部Id
	UserSD map[uint64]*UserStatisticsData
}
type UserStatisticsData struct {
	UserId             uint64 //用户Id
	Name               string //姓名
	StatisticsIntegral int64  //当天输赢积分统计
	StatisticsPlay     int64  //当天本俱乐部次数统计
}

//俱乐部统计数据 现在只统计7天内的----------------------------------------------------------------------------------------
type ClubStatisticsData struct {
	ClubId         int64             //俱乐部Id
	StatisticsTime int64             //统计时间 到天 2019,20,26
	Statistics     []*StatisticsData //统计
}
type StatisticsData struct {
	UserId     uint64 //用户Id
	Name       string //姓名
	Statistics int64  //统计数据  积分统计 为积分 对局次数统计 为对局数
}

//俱乐部 当天统计缓存数据
type ClubStatisticsRedisData struct {
	ClubStatistics map[int64]*ClubStatisticsRedisItemData
}
type ClubStatisticsRedisItemData struct {
	UserId             uint64 //用户Id
	StatisticsIntegral int64  //当天输赢积分统计
	StatisticsPlay     int64  //当天本俱乐部次数统计
}

//添加游戏记录
func AddGameRecord(gr *WirteRecord) (err error) {
	rgd := &GameRecord{ //游戏单局记录
		GameRecordId:  gr.CurGameInfo.RoomRecordId + strconv.Itoa(int(gr.CurGameInfo.Index)),
		Index:         gr.CurGameInfo.Index,
		GameStartTime: gr.CurGameInfo.GameStartTime,
		GameEndTime:   gr.CurGameInfo.GameEndTime,
		RePlayData:    gr.CurGameInfo.RePlayData,
		GamePlayers:   make([]*RoomPlayerInfo, 0, len(gr.CurGameInfo.GamePlayers)),
	}
	rrd := &RoomRecord{}           //房间记录
	if gr.CurGameInfo.Index == 1 { //第一局插入新的记录
		rrd = &RoomRecord{
			RoomRecordId:  gr.CurGameInfo.RoomRecordId,
			GameStartTime: gr.CurGameInfo.GameStartTime,
			TotalInning:   gr.CreateInfo.TotalInning,
			DeskId:        gr.CreateInfo.DeskId,
			GameId:        gr.CreateInfo.GameId,
			ClubId:        gr.CreateInfo.ClubId,
			MasterUid:     gr.CreateInfo.MasterUid,
			Fee:           gr.CreateInfo.Fee,
			PayType:       gr.CreateInfo.PayType,
			DeskInfo:      gr.CreateInfo.DeskInfo,
			GamePlayers:   make([]*RoomPlayerInfo, 0, len(gr.CurGameInfo.GamePlayers)),
		}
		// 插入一条房间记录
		if err = mgoSess.DB("").C(RoomRecordTable).Insert(rrd); err != nil {
			return
		}
	}
	//构建玩家信息
	for _, v := range gr.CurGameInfo.GamePlayers {
		rgd.GamePlayers = append(rgd.GamePlayers, &RoomPlayerInfo{ //单局记录
			UserId:   v.UserId,
			Name:     v.Name,
			Score:    v.Score,
			PreScore: v.PreScore,
		})

		rrd.GamePlayers = append(rrd.GamePlayers, &RoomPlayerInfo{ //玩家总分
			UserId:   v.UserId,
			Name:     v.Name,
			PreScore: v.PreScore,
		})
	}
	//更新每局记录id和总分详情
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"gameplayers": rrd.GamePlayers}, "$push": bson.M{"gamerecords": rgd.GameRecordId}},
		ReturnNew: false,
		Remove:    false,
		Upsert:    true,
	}
	mgoSess.DB("").C(RoomRecordTable).Find(bson.M{"_id": gr.CurGameInfo.RoomRecordId}).Apply(change, nil)

	//记录当局详情
	err = mgoSess.DB("").C(GameRecordTable).Insert(rgd)
	if err == nil {
		if gr.CreateInfo.ClubId != 0 {
			AddClubCurrDayStatistics(gr) //写入统计数据
		}
		AddUserCurrDayStatistics(gr)
	}
	return
}

//添加俱乐部当日统计数据
func AddUserCurrDayStatistics(gr *WirteRecord) (err error) {
	for _, v := range gr.CurGameInfo.GamePlayers {
		ud := &UserCurrDayStatisticsData{}
		_err := mgoSess.DB("").C(UserCurrDayStatisticsTable).Find(bson.M{"_id": v.UserId}).One(ud)
		if _err != nil {
			ud = &UserCurrDayStatisticsData{
				UserId:         v.UserId,
				StatisticsPlay: 0,
			}
		}
		ud.StatisticsPlay++
		if int(ud.StatisticsPlay) == net.Bureau { //达到抽奖次数
			net.GetletteNum(ud.UserId, int(ud.StatisticsPlay))
		}
		_, err = mgoSess.DB("").C(UserCurrDayStatisticsTable).Upsert(bson.M{"_id": v.UserId}, ud)
		if err != nil {
			return err
		}
		SetUserGamePlay(v.UserId, v.Score > 0)
	}
	return err
}

func CleraUserCurrDayStatistics() (err error) {
	err = mgoSess.DB("").C(UserCurrDayStatisticsTable).Remove(nil)
	return
}

//添加俱乐部当日统计数据
func AddClubCurrDayStatistics(gr *WirteRecord) (err error) {
	csd := &ClubCurrDayStatisticsData{}
	_err := mgoSess.DB("").C(ClubCurrDayStatisticsTable).Find(bson.M{"clubid": gr.CreateInfo.ClubId}).One(csd)
	if _err != nil {
		csd = &ClubCurrDayStatisticsData{
			ClubId: gr.CreateInfo.ClubId,
			UserSD: make(map[uint64]*UserStatisticsData),
		}
	}
	for _, v := range gr.CurGameInfo.GamePlayers {
		if u, ok := csd.UserSD[v.UserId]; ok {
			u.StatisticsPlay++
			u.StatisticsIntegral += int64(v.Score)
		} else {
			csd.UserSD[v.UserId] = &UserStatisticsData{
				UserId:             v.UserId,
				StatisticsPlay:     1,
				StatisticsIntegral: int64(v.Score),
			}
		}
	}
	_, err = mgoSess.DB("").C(ClubCurrDayStatisticsTable).Upsert(bson.M{"clubid": gr.CreateInfo.ClubId}, csd)
	if err != nil {
		return err
	}
	return err
}

//查询并清除全部俱乐部
func QueryAndClearAllClubCurrDayStatistics() (rsp []*ClubCurrDayStatisticsData, err error) {
	ce := make([]*ClubCurrDayStatisticsData, 0)
	err = mgoSess.DB("").C(ClubCurrDayStatisticsTable).Find(nil).All(&ce)
	mgoSess.DB("").C(ClubCurrDayStatisticsTable).Remove(nil)
	return ce, err
}

//添加俱乐部对局统计
func AddClubPlayStatistics(clubs *ClubStatisticsData) (err error) {
	err = mgoSess.DB("").C(ClubStatisticsPlayTable).Insert(clubs)
	return err
}

//添加俱乐部积分统计
func AddClubIntegralStatistics(clubs *ClubStatisticsData) (err error) {
	err = mgoSess.DB("").C(ClubStatisticsIntegralTable).Insert(clubs)
	return err
}

//查询用户游戏记录
func QueryUserRoomRecord(uid uint64, start, end int64, _curPage, _limit int32) (rsp []*RoomRecord, err error) {
	var curPage, limit = int(_curPage), int(_limit)
	if limit == 0 {
		limit = 30
	}
	rsp = make([]*RoomRecord, 0)
	query := bson.M{"gameplayers.userid": uid, "gamestarttime": bson.M{"$gte": start, "$lt": end}}
	err = mgoSess.DB("").C(RoomRecordTable).Find(query).Sort("-gamestarttime").Skip(curPage * limit).Limit(limit).All(&rsp)
	return
}

//查询俱乐部的战绩数据
func QueryClubRoomRecord(clubid, start, end int64, _curPage, _limit int32) (rsp []*RoomRecord, err error) {
	var curPage, limit = int(_curPage), int(_limit)
	if limit == 0 {
		limit = 30
	}
	rsp = make([]*RoomRecord, 0)
	query := bson.M{"clubid": clubid, "gamestarttime": bson.M{"$gte": start, "$lt": end}}
	err = mgoSess.DB("").C(RoomRecordTable).Find(query).Sort("-gamestarttime").Skip(curPage * limit).Limit(limit).All(&rsp)
	if err != nil {
		return
	}
	return
}

//查询俱乐部的战绩数据
func QueryClubRoomRecordByRoom(clubid int64, deskid uint64, _curPage, _limit int32) (rsp []*RoomRecord, err error) {
	var curPage, limit = int(_curPage), int(_limit)
	if limit == 0 {
		limit = 30
	}
	rsp = []*RoomRecord{}
	query := bson.M{"clubid": clubid, "deskid": deskid}
	err = mgoSess.DB("").C(RoomRecordTable).Find(query).Sort("-gamestarttime").Skip(curPage * limit).Limit(limit).All(&rsp)
	if err != nil {
		return
	}
	return
}

//查询俱乐部的战绩数据
func QueryRoomRecordByRoom(deskid uint64) (rsp *RoomRecord, err error) {
	rsp = &RoomRecord{}
	query := bson.M{"deskid": deskid}
	err = mgoSess.DB("").C(RoomRecordTable).Find(query).Sort("-gamestarttime").One(rsp)
	if err != nil {
		return
	}
	return
}

//查询游戏具体详情数据
func QueryGameRecord(roomRecordId string) (rsp []*GameRecord, err error) {
	rsp = []*GameRecord{}
	//找到对应的roomRecord
	find := &RoomRecord{}
	err = mgoSess.DB("").C(RoomRecordTable).Find(bson.M{"_id": roomRecordId}).One(find)
	if err != nil {
		return
	}
	for _, v := range find.GameRecords {
		ce := &GameRecord{}
		if err := mgoSess.DB("").C(GameRecordTable).Find(bson.M{"_id": v}).Select(bson.M{"replaydata": 0}).One(ce); err == nil { //不返回游戏回放数据
			rsp = append(rsp, ce)
		}
	}
	return
}

//查询游戏复盘数据
func QueryGameRePlayRecord(gamerecordId string) (rsp *GameRecord, err error) {
	rsp = &GameRecord{}
	err = mgoSess.DB("").C(GameRecordTable).Find(bson.M{"_id": gamerecordId}).One(&rsp)
	return
}

//查询俱乐部对局统计数据
func QueryClubPlayStatistics(clubid uint64, start int64, end int64) (rsp []*ClubStatisticsData, err error) {
	ce := make([]*ClubStatisticsData, 0)
	if err := mgoSess.DB("").C(ClubStatisticsPlayTable).Find(bson.M{"clubid": clubid, "statisticstime": bson.M{"$gte": start, "$lt": end}}).All(&ce); err == nil {
		return ce, nil
	} else {
		return ce, err
	}
}

//查询俱乐部积分统计数据
func QueryClubIntegralStatistics(clubid uint64, start int64, end int64) (rsp []*ClubStatisticsData, err error) {
	ce := make([]*ClubStatisticsData, 0)
	if err := mgoSess.DB("").C(ClubStatisticsIntegralTable).Find(bson.M{"clubid": clubid, "statisticstime": bson.M{"$gte": start, "$lt": end}}).All(&ce); err == nil {
		return ce, nil
	} else {
		return nil, err
	}
}
