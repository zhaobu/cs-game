package mgo

import (
	"github.com/globalsign/mgo/bson"
	"strconv"
)

const (
	RoomRecordTable = "roomrecord"				//游戏记录数据
	GameRecordTable = "gamerecord"				//游戏记录数据
	UserGameRecordTable = "usergamerecord"		//用户游戏记录
	GameRePlayDataTable = "gamereplay"			//游戏复盘数据
	ClubStatisticsPlayTable = "clubstatisticsplay"		//俱乐部游戏对局次数统计数据
	ClubStatisticsIntegralTable = "clubstatisticsintegral"		//俱乐部游戏积分统计数据
)

//WriteGameRecordReq 重写 避免直接应用导致包体穿插报错的问题
type WirteRecord struct {
	 RoomRecordId string                   //游戏房间创建时生成的唯一Id 用户关联该房间
	 GameId string	 			            //游戏Id
	 RoomType int32				        //房间类型 1 俱乐部房间  2 好友房间
	 ClubId int64			            //俱乐部Id 当 AreaType 1 时需要填写
	 RoomId uint32                        //房间号
	 Index int32                           //游戏当前局数
	 GameStartTime int64		            //开始时间 存储时间错
	 GameEndTime int64			            //结束时间 存储时间错
	 PlayerInfos []*GamePlayerInfo    		//游戏内玩家数据
	 RePlayData []byte             		//复盘数据                暂时可不写
}


//用户游戏记录
type UserGameRecord struct {
	UserId 		uint64			//用户Id
	GameId		string			//游戏Id
	RoomType	int32			//房间类型 1俱乐部房间 2好友房
	RoomId		uint32			//房间号
	GameStartTime int64		//开始时间 存储时间错
	WinIntegral	int32			//该房间内累计输赢积分
	RoomRecord  string			//房间记录关联Id
	GameRecord 	[]uint64		//游戏记录关联Id
}

//房间记录
type RoomRecord struct {
	RoomRecordId string				//房间记录Id 关联房间的数据库唯一id
	RoomId		uint32				//房间号
	GameId		string				//游戏Id
	RoomType	int32				//房间类型 1俱乐部房间 2好友房
	ClubId		int64				//俱乐部Id 当 AreaType 1 时需要填写
	GamePlayers map[uint64]*RoomPlayerInfo	//游戏参与玩家信息 为了减少查询量 将必要信息进行存储
	GameRecords []string			//房间内游戏数据 数组
}
type RoomPlayerInfo struct {
	UserId uint64			//用户Id
	Name string				//姓名
	WinIntegral	int32		//输赢积分 房间内的统计数据
}

//游戏记录
type GameRecord struct {
	GameRecordId string				//游戏记录ID 主键
	GameId		string				//游戏Id
	RoomType	int32				//房间类型 1俱乐部房间 2好友房
	ClubId		int64				//俱乐部Id 当 AreaType 1 时需要填写
	RoomId		uint32				//房间号
	Index		int32				//第几局
	GameStartTime int64			//开始时间 存储时间错
	GameEndTime int64			//结束时间 存储时间错
	GamePlayers []*GamePlayerInfo	//游戏参与玩家信息 为了减少查询量 将必要信息进行存储
}
type GamePlayerInfo struct {
	UserId uint64			//用户Id
	Name string				//姓名
	BringinIntegral int32	//带入积分
	WinIntegral	int32		//输赢积分
}

//游戏复盘记录					由于复盘数据过大 采取拆分单独数据表
type GameRePlayData struct {
	GameRecordId	string			//游戏记录ID
	RePlayData	[]byte			//复盘协议数据
}

//俱乐部统计数据 现在只统计7天内的
type ClubStatisticsData struct {
	ClubId		int64				//俱乐部Id 当 AreaType 1 时需要填写
	StatisticsTime int64			//统计时间 到天 2019,20,26
	Statistics []*StatisticsData	//统计
}
type StatisticsData struct {
	UserId uint64			//用户Id
	Name string				//姓名
	Statistics int64		//统计数据  积分统计 为积分 对局次数统计 为对局数
}

//俱乐部 当天统计缓存数据
type ClubStatisticsRedisData struct {
	ClubStatistics map[int64]*ClubStatisticsRedisItemData
}
type ClubStatisticsRedisItemData struct {
	UserId uint64				//用户Id
	StatisticsIntegral int64	//当天输赢积分统计
	StatisticsPlay int64		//当天本俱乐部次数统计
}

//添加游戏记录
func AddGameRecord(gr *WirteRecord)(err error)  {
	rgd := &GameRecord{
		GameRecordId : gr.RoomRecordId + strconv.Itoa(int(gr.Index)),
		GameId:gr.GameId,
		RoomType:gr.RoomType,
		ClubId : gr.ClubId,
		RoomId : gr.RoomId,
		Index : gr.Index,
		GameStartTime : gr.GameStartTime,
		GameEndTime : gr.GameEndTime,
		GamePlayers :[]*GamePlayerInfo{},
	}
	rrd := &RoomRecord{}
	grp := &GameRePlayData{
		GameRecordId:rgd.GameRecordId,
		RePlayData:gr.RePlayData,
	}
	_err := mgoSess.DB("").C(RoomRecordTable).Find(bson.M{"roomrecordid": gr.RoomRecordId}).One(rrd)
	if _err != nil {
		rrd = &RoomRecord{
			RoomRecordId:gr.RoomRecordId,
			RoomId:gr.RoomId,
			GameId:gr.GameId,
			RoomType:gr.RoomType,
			ClubId:gr.ClubId,
			GamePlayers:make(map[uint64]*RoomPlayerInfo),
			GameRecords:[]string{},
		}
	}
	rrd.GameRecords = append(rrd.GameRecords,rgd.GameRecordId)

	for _,v:= range gr.PlayerInfos{
		ugr := &UserGameRecord{}
		_err1 := mgoSess.DB("").C(UserGameRecordTable).Find(bson.M{"userid": v.UserId,"roomrecord":gr.RoomRecordId}).One(ugr)
		if _err1 != nil {
			ugr = &UserGameRecord{
				UserId:v.UserId,
				GameId:gr.GameId,
				RoomType:gr.RoomType,
				RoomId:gr.RoomId,
				GameStartTime:gr.GameStartTime,
				WinIntegral:v.WinIntegral,
				RoomRecord : gr.RoomRecordId,
			}
		}
		_, err = mgoSess.DB("").C(UserGameRecordTable).Upsert(bson.M{"userid": ugr.UserId,"roomrecord":ugr.RoomRecord}, ugr)
		if err != nil {
			return err
		}
		if p,ok := rrd.GamePlayers[v.UserId];ok{
			p.WinIntegral += v.WinIntegral
		}else{
			rrd.GamePlayers[v.UserId] = &RoomPlayerInfo{
				UserId :v.UserId,
				Name :v.Name,
				WinIntegral	:v.WinIntegral,
			}
		}
		rgd.GamePlayers = append(rgd.GamePlayers,&GamePlayerInfo{
			UserId :v.UserId,
			Name :v.Name,
			BringinIntegral:v.BringinIntegral,
			WinIntegral	:v.WinIntegral,
		})

	}
	err = mgoSess.DB("").C(GameRecordTable).Insert(rgd)
	if err != nil {
		return err
	}
	_, err = mgoSess.DB("").C(RoomRecordTable).Upsert(bson.M{"roomrecordid": rrd.RoomRecordId}, rrd)
	if err != nil {
		return err
	}
	_, err = mgoSess.DB("").C(GameRePlayDataTable).Upsert(bson.M{"gamerecordid": grp.GameRecordId}, grp)
	return err
}

//添加俱乐部对局统计
func AddClubPlayStatistics(clubs *ClubStatisticsData)(err error)  {
	err = mgoSess.DB("").C(ClubStatisticsPlayTable).Insert(clubs)
	return  err
}

//添加俱乐部积分统计
func AddClubIntegralStatistics(clubs *ClubStatisticsData)(err error)  {
	err = mgoSess.DB("").C(ClubStatisticsIntegralTable).Insert(clubs)
	return  err
}
//查询用户游戏记录
func QueryUserRoomRecord(uid uint64,start int64,end int64)(rsp []*RoomRecord,err error)  {
	rsp = make([]*RoomRecord, 0)
	find := make([]*UserGameRecord, 0)
	err = mgoSess.DB("").C(UserGameRecordTable).Find(bson.M{"userid": uid,"gamestarttime": bson.M{"$gte": start, "$lt": end}}).All(&find)
	if err != nil {
		return
	}
	for _, f := range find {
		ce := &RoomRecord{}
		if err := mgoSess.DB("").C(RoomRecordTable).Find(bson.M{"roomrecordid": f.RoomRecord}).One(ce); err == nil {
			rsp = append(rsp, ce)
		}
	}
	return rsp,nil
}

//查询俱乐部的战绩数据
func QueryClubRoomRecord(clubid uint64,start int64,end int64)(rsp []*RoomRecord,err error)  {
	rsp = make([]*RoomRecord, 0)
	err = mgoSess.DB("").C(RoomRecordTable).Find(bson.M{"clubid": clubid,"gamestarttime": bson.M{"$gte": start, "$lt": end}}).All(&rsp)
	if err != nil{
		return nil,err
	}
	return rsp,nil
}

//查询游戏具体详情数据
func QueryGameRecord(gamerecordId []string) (rsp []*GameRecord,err error) {
	rsp = []*GameRecord{}
	for _,v:= range gamerecordId {
		ce := &GameRecord{}
		if err := mgoSess.DB("").C(GameRecordTable).Find(bson.M{"gamerecordid": v}).One(ce); err == nil {
			rsp = append(rsp, ce)
		}
	}
	return rsp,nil
}

//查询游戏复盘数据
func QueryGameRePlayRecord(gamerecordId string) (rsp *GameRePlayData,err error) {
	ce := &GameRePlayData{}
	if err := mgoSess.DB("").C(GameRePlayDataTable).Find(bson.M{"gamerecordid": gamerecordId}).One(ce); err == nil {
		return ce,nil
	}else{
		return  nil,err
	}
}

//查询俱乐部对局统计数据
func QueryClubPlayStatistics(clubid uint64,start int64,end int64) (rsp []*ClubStatisticsData,err error) {
	ce := []*ClubStatisticsData{}
	if err := mgoSess.DB("").C(ClubStatisticsPlayTable).Find(bson.M{"clubid": clubid,"statisticstime": bson.M{"$gte": start, "$lt": end}}).All(ce); err == nil {
		return ce,nil
	}else{
		return  nil,err
	}
}

//查询俱乐部积分统计数据
func QueryClubIntegralStatistics(clubid uint64,start int64,end int64) (rsp []*ClubStatisticsData,err error) {
	ce := []*ClubStatisticsData{}
	if err := mgoSess.DB("").C(ClubStatisticsIntegralTable).Find(bson.M{"clubid": clubid,"statisticstime": bson.M{"$gte": start, "$lt": end}}).All(ce); err == nil {
		return ce,nil
	}else{
		return  nil,err
	}
}
