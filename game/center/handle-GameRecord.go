package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/gamerecord"
	sort "cy/game/util/tools/Sort"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	ClubStatistics *ClubGameStatistics
	clublock  sync.RWMutex
)

type ClubGameStatistics struct {
	ClubSD map[int64]*ClubStatisticsData
}

type ClubStatisticsData struct {
	UserSD map[uint64]*UserStatisticsData
}

type UserStatisticsData struct {
	UserId uint64				//用户Id
	Name string					//姓名
	StatisticsIntegral int64	//当天输赢积分统计
	StatisticsPlay int64		//当天本俱乐部次数统计
}

func GameRecord_Init()  {
	StartTimer_ResetClubStatistics()
}

//游戏服务器请求写入战绩
func (p *center)WriteGameRecordReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.WriteGameRecordReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return err
	}
	gr := &mgo.WirteRecord{
		RoomRecordId :  req.RoomRecordId,
		GameId		:	req.GameId,
		RoomType 	: 	req.RoomType,
		ClubId		:   req.ClubId,
		RoomId		:   req.RoomId,
		Index		:   req.Index,
		GameStartTime : req.GameStartTime,
		GameEndTime :   req.GameEndTime,
		PayType		:	req.PayType,
		PlayerInfos :   []*mgo.GamePlayerInfo{},
		RePlayData	:	req.RePlayData,
	}
	for _,v := range req.PlayerInfos{
		gr.PlayerInfos = append(gr.PlayerInfos,&mgo.GamePlayerInfo{
			UserId :v.UserId,					//用户Id
			Name :v.Name,						//姓名
			BringinIntegral :v.BringinIntegral,	//带入积分
			WinIntegral	:v.WinIntegral,			//输赢积分
		})
	}
	err = mgo.AddGameRecord(gr)
	if err != nil {
		panic("写入战绩错误 err = " + err.Error())
	}
	WriteClubGameStatistics(gr)					//写入统计数据
	return
}

//请求查询房间记录
func (p *center)QueryRoomRecordReq(ctx context.Context, args *codec.Message, reply *codec.Message)(err error)   {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error("解析消息 1 失败"+ err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryRoomRecordReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgamerecord.QueryRoomRecordRsp{
		Error: 0,
		Datas: []*pbgamerecord.RoomRecord{},
	}
	querydata := []*mgo.RoomRecord{}
	if req.QueryType == 1 || req.QueryType == 2 {
		querydata,err = mgo.QueryUserRoomRecord(req.QueryParam,req.QueryStartTime,req.QueryEndTime)
		if err != nil {
			logrus.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}else if req.QueryType == 3 {
		querydata,err = mgo.QueryClubRoomRecord(req.QueryParam,req.QueryStartTime,req.QueryEndTime)
		if err != nil {
			logrus.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}else if req.QueryType == 4 {
		querydata,err = mgo.QueryClubRoomRecordByRoom(req.QueryParam,req.QueryParam2)
		if err != nil {
			logrus.Warn("查询俱乐部房间数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}
	for _,v := range querydata{
		_data := &pbgamerecord.RoomRecord{
			RoomRecordId:v.RoomRecordId,
			GameStartTime:v.GameStartTime,
			RoomId:v.RoomId,
			GameId:v.GameId,
			RoomType:v.RoomType,
			ClubId:v.ClubId,
			TotalJuNun:v.TotalJuNun,
			GamePlayers : []*pbgamerecord.RoomPlayerInfo{},
			GameRecordIds:v.GameRecords,
		}
		for _,v1 := range v.GamePlayers {
			_data.GamePlayers = append(_data.GamePlayers,&pbgamerecord.RoomPlayerInfo{
				UserId:v1.UserId,
				Name:v1.Name,
				WinIntegral:v1.WinIntegral,
			})
		}
		rsp.Datas = append(rsp.Datas,_data)
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	return
}

//请求查询游戏详情记录
func (p *center)QueryGameRecordReq(ctx context.Context, args *codec.Message, reply *codec.Message)(err error){
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error("解析消息 1 失败"+ err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryGameRecordReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgamerecord.QueryGameRecordRsp{
		Error: 0,
		Records:[]*pbgamerecord.GameRecord{},
	}
	querydata := []*mgo.GameRecord{}
	if len(req.GameRecordIds) > 0 {
		querydata,err = mgo.QueryGameRecord(req.GameRecordIds)
		if err != nil {
			logrus.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
		}
		for _,v := range querydata{
			_data := &pbgamerecord.GameRecord{
				GameRecordId:v.GameRecordId,
				RoomId:v.RoomId,
				GameId:v.GameId,
				RoomType:v.RoomType,
				ClubId:v.ClubId,
				GameStartTime:v.GameStartTime,
				GameEndTime:v.GameEndTime,
				GamePlayers : []*pbgamerecord.GamePlayerInfo{},
			}
			for _,v1 := range v.GamePlayers {
				_data.GamePlayers = append(_data.GamePlayers,&pbgamerecord.GamePlayerInfo{
					UserId:v1.UserId,
					Name:v1.Name,
					BringinIntegral:v1.BringinIntegral,
					WinIntegral:v1.WinIntegral,
				})
			}
			rsp.Records = append(rsp.Records,_data)
		}
	}else{
		rsp.Error = 1
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	return
}

//请求查询游戏复盘记录
func (p *center)QueryGameRePlayRecord(ctx context.Context, args *codec.Message, reply *codec.Message)(err error){
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error("解析消息 1 失败"+ err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryGameRePlaydReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgamerecord.QueryGameRePlaydRsp{
		Error: 0,
		RePlayData: []byte{},
	}
	querydata := &mgo.GameRePlayData{}
	if len(req.GameRecordIds) > 0 {
		querydata,err = mgo.QueryGameRePlayRecord(req.GameRecordIds)
		if err != nil {
			logrus.Warn("查询复盘数据失败 err = " + err.Error())
			rsp.Error = 1
		}
		rsp.RePlayData = querydata.RePlayData
	}else{
		rsp.Error = 1
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息解析失败"+err.Error())
	}
	return
}

//查询俱乐部统计结果数据
func (p *center)QueryClubStatisticsReq(ctx context.Context, args *codec.Message, reply *codec.Message)(err error)  {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryClubStatisticsReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgamerecord.QueryClubStatisticsRsp{
		Error:0,
		StatisticsDatas:[]*pbgamerecord.StatisticsData{},
	}
	querydata := []*mgo.ClubStatisticsData{}
	if req.QueryType == 1 {
		querydata,err = mgo.QueryClubPlayStatistics(req.QueryClubId,req.QueryStartTime,req.QueryEndTime)
		if err != nil {
			logrus.Warn("查询俱乐部对局统计数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}else if req.QueryType == 2{
		querydata,err = mgo.QueryClubIntegralStatistics(req.QueryClubId,req.QueryStartTime,req.QueryEndTime)
		if err != nil {
			logrus.Warn("查询俱乐部积分统计数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}
	if querydata != nil {
		tmp := make(map[uint64]*mgo.StatisticsData)	//统计)
		for _,v := range querydata{
			for _,v1 := range v.Statistics{
				if d,ok := tmp[v1.UserId];!ok{
					tmp[v1.UserId] = &mgo.StatisticsData{
						UserId:v1.UserId,
						Name:v1.Name,
						Statistics:v1.Statistics,
					}
				}else{
					d.Statistics += v1.Statistics
				}
			}
		}
		csd := []interface{}{}
		for _,v2 := range tmp{
			csd = append(csd,v2)
		}

		//先做对局次数排名
		sort.Sort(csd, func(a interface{}, b interface{}) int8 {
			_a := a.(*mgo.StatisticsData)
			_b := b.(*mgo.StatisticsData)
			if _a.Statistics < _b.Statistics {
				return 1
			}else if _a.Statistics == _b.Statistics{
				return 0
			}else{
				return -1
			}
		})
		index := 0
		for _,u := range csd{
			if index > 10 {
				break
			}
			sdata := u.(*mgo.StatisticsData)
			rsp.StatisticsDatas = append(rsp.StatisticsDatas,&pbgamerecord.StatisticsData{
				UserId:sdata.UserId,
				Name:sdata.Name,
				Statistics:sdata.Statistics,
			})
			index++
		}
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败"+err.Error())
	}
	return
}
//--------------------------------------------------统计数据------------------------------------------------------------
//写入俱乐部统计数据到缓存
func WriteClubGameStatistics(gr  *mgo.WirteRecord)  {

	defer clublock.Unlock()
	clublock.Lock()

	if gr.RoomType != 1 {
		return
	}
	if ClubStatistics == nil{
		ClubStatistics = &ClubGameStatistics{
			ClubSD:make(map[int64]*ClubStatisticsData),
		}
	}
	if _,ok := ClubStatistics.ClubSD[gr.ClubId]; !ok{
		ClubStatistics.ClubSD[gr.ClubId] = &ClubStatisticsData{
			UserSD:make(map[uint64]*UserStatisticsData),
		}
	}
	for _,v := range gr.PlayerInfos {
		if _,ok := ClubStatistics.ClubSD[gr.ClubId].UserSD[v.UserId]; !ok{
			ClubStatistics.ClubSD[gr.ClubId].UserSD[v.UserId] = &UserStatisticsData{
				UserId:v.UserId,
				Name:v.Name,
				StatisticsIntegral :0,
				StatisticsPlay:0,
			}
		}
		ClubStatistics.ClubSD[gr.ClubId].UserSD[v.UserId].StatisticsIntegral += int64(v.WinIntegral)
		ClubStatistics.ClubSD[gr.ClubId].UserSD[v.UserId].StatisticsPlay ++
	}
}

//启动定时任务 暂定每天凌晨执行统计数据清零
func StartTimer_ResetClubStatistics()  {
	now := time.Now()
	next := now.Add(time.Hour * 24)
	next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	t := next.Sub(now)
	timer := time.NewTimer(t)
	go func() {
		 <-timer.C
		ResetClubGameStatistics()
	}()

}

//每天定时重置统计数据并写入到数据库中
func ResetClubGameStatistics()  {
	clublock.Lock()
	if ClubStatistics == nil{
		return
	}
	_ClubStatistics := ClubStatistics
	ClubStatistics = nil
	clublock.Unlock()
	for i,v1 := range _ClubStatistics.ClubSD {
		CSData :=  &mgo.ClubStatisticsData{
			ClubId:i,
			StatisticsTime:time.Now().Unix(),
			Statistics:[]*mgo.StatisticsData{},
		}

		usd := []interface{}{}
		for _,v2 := range v1.UserSD{
			usd = append(usd,v2)
		}

		//先做对局次数排名
		sort.Sort(usd, func(a interface{}, b interface{}) int8 {
			_a := a.(*UserStatisticsData)
			_b := b.(*UserStatisticsData)
			if _a.StatisticsPlay < _b.StatisticsPlay {
				return 1
			}else if _a.StatisticsPlay == _b.StatisticsPlay{
				return 0
			}else{
				return -1
			}
		})
		index := 0
		for _,u := range usd{
			if index > 10 {
				break
			}
			userdata := u.(*UserStatisticsData)
			CSData.Statistics = append(CSData.Statistics,&mgo.StatisticsData{
				UserId:userdata.UserId,
				Statistics:userdata.StatisticsPlay,
			})
			index++
		}
		err := mgo.AddClubPlayStatistics(CSData)
		if err != nil {
			panic("写入对局统计数据失败 err = " + err.Error())
		}

		//再排名积分
		sort.Sort(usd, func(a interface{}, b interface{}) int8 {
			_a := a.(*UserStatisticsData)
			_b := b.(*UserStatisticsData)
			if _a.StatisticsIntegral < _b.StatisticsIntegral {
				return 1
			}else if _a.StatisticsIntegral == _b.StatisticsIntegral{
				return 0
			}else{
				return -1
			}
		})
		index = 0
		for _,u := range usd{
			if index > 10 {
				break
			}
			userdata := u.(*UserStatisticsData)
			CSData.Statistics = append(CSData.Statistics,&mgo.StatisticsData{
				UserId:userdata.UserId,
				Name:userdata.Name,
				Statistics:userdata.StatisticsIntegral,
			})
			index++
		}
		err = mgo.AddClubIntegralStatistics(CSData)
		if err != nil {
			panic("写入统计数据失败 err = " + err.Error())
		}
	}
	StartTimer_ResetClubStatistics()		//重新计算下一次时间
}