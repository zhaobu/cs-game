package main

import (
	"context"
	"game/codec"
	"game/db/mgo"
	pbcommon "game/pb/common"
	pbgamerecord "game/pb/gamerecord"
	"game/util"
	sort "game/util/tools/Sort"
	"fmt"
	"github.com/robfig/cron"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	//ClubStatistics *ClubGameStatistics
	clublock       sync.RWMutex
)

//type ClubGameStatistics struct {
//	ClubSD map[int64]*ClubStatisticsData
//}
//
//type ClubStatisticsData struct {
//	UserSD map[uint64]*UserStatisticsData
//}
//
//type UserStatisticsData struct {
//	UserId             uint64 //用户Id
//	Name               string //姓名
//	StatisticsIntegral int64  //当天输赢积分统计
//	StatisticsPlay     int64  //当天本俱乐部次数统计
//}

func GameRecord_Init() {
	StartTimer_ResetClubStatistics()
}

//请求查询房间记录(点击战绩按钮或者俱乐部内查找)
func (p *center) QueryRoomRecordReq(ctx context.Context, args *codec.Message, reply *codec.Message) error {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error("解析消息 1 失败" + err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryRoomRecordReq)
	if !ok {
		err = fmt.Errorf("not *pbgamerecord.QueryRoomRecordReq")
		tlog.Error(err.Error())
		return err
	}
	log.Debugf("QueryRoomRecordReq请求参数:req:%s", util.PB2JSON(req, true))
	rsp := &pbgamerecord.QueryRoomRecordRsp{
		Error: 0,
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	querydata := []*mgo.RoomRecord{}
	if req.QueryType == 1 { //按userId查询
		rsp.QueryUserId = req.QueryUserId	//客户端需求 经查询的uid 返回给用户
		querydata, err = mgo.QueryUserRoomRecord(req.QueryUserId, req.QueryStartTime, req.QueryEndTime, req.CurPage, req.Limit)
		if err != nil {
			tlog.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	} else if req.QueryType == 2 { //按俱乐部id查询
		if req.UserIdentity == 1 || req.UserIdentity == 2 {	//普通会员
			querydata, err = mgo.QueryClubRoomRecord(req.QueryClubId, req.QueryStartTime, req.QueryEndTime, req.CurPage, req.Limit)
			if err != nil {
				tlog.Warn("查询用户数据失败 err = " + err.Error())
				rsp.Error = 1
			}
		}else{
			querydata, err = mgo.QueryClubRoomRecordByUser(req.Head.UserID,req.QueryClubId, req.QueryStartTime, req.QueryEndTime, req.CurPage, req.Limit)
			if err != nil {
				tlog.Warn("查询用户数据失败 err = " + err.Error())
				rsp.Error = 1
			}
		}
	} else if req.QueryType == 3 { //按俱乐部+房间号查询
		querydata, err = mgo.QueryClubRoomRecordByRoom(req.QueryClubId, req.QueryRoomId, req.CurPage, req.Limit)
		if err != nil {
			tlog.Warn("查询俱乐部房间数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	} else {
		rsp.Error = 1
	}
	rsp.Datas = make([]*pbgamerecord.RoomRecord, 0, len(querydata))
	for _, v := range querydata {
		_data := &pbgamerecord.RoomRecord{
			RoomRecordId:  v.RoomRecordId,
			GameStartTime: v.GameStartTime,
			DeskId:        v.DeskId,
			GameId:        v.GameId,
			ClubId:        v.ClubId,
			TotalInning:   v.TotalInning,
			PayType:       v.PayType,
			DeskInfo:      &pbgamerecord.GameAction{ActName: v.DeskInfo.ActName, ActValue: v.DeskInfo.ActValue},
			GamePlayers:   make([]*pbgamerecord.RoomPlayerInfo, 0, len(v.GamePlayers)),
			GameRecordIds: v.GameRecords,
		}
		for _, v1 := range v.GamePlayers {
			udata,err := mgo.QueryUserInfo(v1.UserId)
			playerinfo := &pbgamerecord.RoomPlayerInfo{
				UserId:     v1.UserId,
				Name:       v1.Name,
				TotalScore: v1.PreScore,
			}
			if err == nil{
				playerinfo.HeadUrl = udata.Profile
			}
			_data.GamePlayers = append(_data.GamePlayers,playerinfo )

		}
		if req.UserIdentity == 1 || req.UserIdentity == 2 {
			rsp.TotalFree += v.Fee * uint32(len(v.GamePlayers))
		}else{
			if v.PayType != 2 {
				if v.MasterUid == req.Head.UserID{
					rsp.TotalFree += v.Fee * uint32(len(v.GamePlayers))
				}
			}else{
				rsp.TotalFree += v.Fee
			}
		}
		rsp.Datas = append(rsp.Datas, _data)
	}
	rsp.TotalRooms = uint32(len(rsp.Datas))
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		tlog.Error("消息封装失败", zap.Error(err))
	}
	return err
}

//请求查询游戏详情记录(点击详情按钮)
func (p *center) QueryGameRecordReq(ctx context.Context, args *codec.Message, reply *codec.Message) error {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error("解析消息 1 失败" + err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryGameRecordReq)
	if !ok {
		err = fmt.Errorf("not *pbgamerecord.QueryGameRecordReq")
		tlog.Error(err.Error())
		return err
	}
	rsp := &pbgamerecord.QueryGameRecordRsp{
		Error: 0,
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	log.Debugf("QueryGameRecordReq请求参数req:%s", util.PB2JSON(req, true))
	querydata := []*mgo.GameRecord{}
	if querydata, err = mgo.QueryGameRecord(req.RoomRecordId); err != nil {
		tlog.Error("查询用户数据失败 err = " + err.Error())
		rsp.Error = 1
		return err
	}
	rsp.Records = make([]*pbgamerecord.GameRecord, 0, len(querydata))
	for _, v := range querydata {
		_data := &pbgamerecord.GameRecord{
			GameRecordId:  v.GameRecordId,
			Index:         v.Index,
			GameStartTime: v.GameStartTime,
			GameEndTime:   v.GameEndTime,
			GamePlayers:   make([]*pbgamerecord.GamePlayerInfo, 0, len(v.GamePlayers)),
		}
		for _, v1 := range v.GamePlayers {
			_data.GamePlayers = append(_data.GamePlayers, &pbgamerecord.GamePlayerInfo{
				UserId:   v1.UserId,
				Name:     v1.Name,
				Score:    v1.Score,
				PreScore: v1.PreScore,
			})
		}
		rsp.Records = append(rsp.Records, _data)
	}

	if err = codec.Pb2Msg(rsp, reply); err != nil {
		tlog.Error("消息封装失败", zap.Error(err))
	}
	log.Debugf("QueryGameRecordReq请求结果:rsp:%s", util.PB2JSON(rsp, true))
	return err
}

//请求查询游戏复盘记录(点击游戏回放按钮)
func (p *center) QueryGameRePlaydReq(ctx context.Context, args *codec.Message, reply *codec.Message) error {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error("解析消息 1 失败" + err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryGameRePlaydReq)
	if !ok {
		err = fmt.Errorf("not *pbgamerecord.QueryGameRePlaydReq")
		tlog.Error(err.Error())
		return err
	}
	rsp := &pbgamerecord.QueryGameRePlaydRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	log.Debugf("QueryGameRePlayRecord请求参数:req:%s", util.PB2JSON(req, true))
	querydata := &mgo.GameRecord{}
	if querydata, err = mgo.QueryGameRePlayRecord(req.GameRecordId); err != nil {
		tlog.Warn("查询复盘数据失败 err = " + err.Error())
		rsp.Error = 1
		return err
	}
	rsp.RePlayData = make([]*pbgamerecord.GameAction, 0, len(querydata.RePlayData))
	for _, v := range querydata.RePlayData {
		rsp.RePlayData = append(rsp.RePlayData, &pbgamerecord.GameAction{ActName: v.ActName, ActValue: v.ActValue})
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		tlog.Error("消息封装失败", zap.Error(err))
	}
	log.Debugf("QueryGameRePlayRecord请求结果:rsp:%s", util.PB2JSON(rsp, true))
	return err
}

//查询俱乐部统计结果数据
func (p *center) QueryClubStatisticsReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbgamerecord.QueryClubStatisticsReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		tlog.Error(err.Error())
		return
	}
	rsp := &pbgamerecord.QueryClubStatisticsRsp{
		Error:           0,
		StatisticsDatas: []*pbgamerecord.StatisticsData{},
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	querydata := []*mgo.ClubStatisticsData{}
	if req.QueryType == 1 {
		querydata, err = mgo.QueryClubPlayStatistics(req.QueryClubId, req.QueryStartTime, req.QueryEndTime)
		if err != nil {
			tlog.Warn("查询俱乐部对局统计数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	} else if req.QueryType == 2 {
		querydata, err = mgo.QueryClubIntegralStatistics(req.QueryClubId, req.QueryStartTime, req.QueryEndTime)
		if err != nil {
			tlog.Warn("查询俱乐部积分统计数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	}
	if querydata != nil {
		tmp := make(map[uint64]*mgo.StatisticsData) //统计)
		for _, v := range querydata {
			for _, v1 := range v.Statistics {
				if d, ok := tmp[v1.UserId]; !ok {
					tmp[v1.UserId] = &mgo.StatisticsData{
						UserId:     v1.UserId,
						Name:       v1.Name,
						Statistics: v1.Statistics,
					}
				} else {
					d.Statistics += v1.Statistics
				}
			}
		}
		csd := []interface{}{}
		for _, v2 := range tmp {
			csd = append(csd, v2)
		}

		sort.Sort(csd, func(a interface{}, b interface{}) int8 {
			_a := a.(*mgo.StatisticsData)
			_b := b.(*mgo.StatisticsData)
			if _a.Statistics < _b.Statistics {
				return 1
			} else if _a.Statistics == _b.Statistics {
				return 0
			} else {
				return -1
			}
		})
		index := 0
		for _, u := range csd {
			if index > 10 {
				break
			}
			sdata := u.(*mgo.StatisticsData)
			data := &pbgamerecord.StatisticsData{
				UserId:     sdata.UserId,
				Name:       sdata.Name,
				Statistics: sdata.Statistics,
			}
			udata,err := mgo.QueryUserInfo(sdata.UserId)
			if err == nil{
				data.HeadUrl = udata.Profile
			}
			rsp.StatisticsDatas = append(rsp.StatisticsDatas,data)
			index++
		}
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		panic("消息封装失败" + err.Error())
	}
	return
}

//--------------------------------------------------统计数据------------------------------------------------------------

//写入俱乐部统计数据到缓存
//func WriteClubGameStatistics(gr *mgo.WirteRecord) {
//	defer clublock.Unlock()
//	clublock.Lock()
//	if gr.CreateInfo.ClubId == 0 {
//		return
//	}
//	if ClubStatistics == nil {
//		ClubStatistics = &ClubGameStatistics{
//			ClubSD: make(map[int64]*ClubStatisticsData),
//		}
//	}
//	if _, ok := ClubStatistics.ClubSD[gr.CreateInfo.ClubId]; !ok {
//		ClubStatistics.ClubSD[gr.CreateInfo.ClubId] = &ClubStatisticsData{
//			UserSD: make(map[uint64]*UserStatisticsData),
//		}
//	}
//	for _, v := range gr.CurGameInfo.GamePlayers {
//		if _, ok := ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId]; !ok {
//			ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId] = &UserStatisticsData{
//				UserId:             v.UserId,
//				Name:               v.Name,
//				StatisticsIntegral: 0,
//				StatisticsPlay:     0,
//			}
//		}
//		ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId].StatisticsIntegral += int64(v.Score)
//		ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId].StatisticsPlay++
//	}
//}

//启动定时任务 暂定每天凌晨执行统计数据清零
func StartTimer_ResetClubStatistics() {
	c := cron.New()
	spec := "0 0 2 * * ?"
	c.AddFunc(spec, func() {
		ResetClubGameStatistics()
	})
	c.Start()
}

//每天定时重置统计数据并写入到数据库中
func ResetClubGameStatistics() {
 	mgo.CleraUserCurrDayStatistics()
	cds, err := mgo.QueryAndClearAllClubCurrDayStatistics()
	if err != nil{
		tlog.Error("查询俱乐部当日游戏缓存数据错误 err = " + err.Error())
		return
	}
	for _, v1 := range cds {
		CSData := &mgo.ClubStatisticsData{
			ClubId:         v1.ClubId,
			StatisticsTime: time.Now().Unix(),
			Statistics:     []*mgo.StatisticsData{},
		}

		usd := []interface{}{}
		for _, v2 := range v1.UserSD {
			usd = append(usd, v2)
		}

		//先做对局次数排名
		sort.Sort(usd, func(a interface{}, b interface{}) int8 {
			_a := a.(*mgo.UserStatisticsData)
			_b := b.(*mgo.UserStatisticsData)
			if _a.StatisticsPlay < _b.StatisticsPlay {
				return 1
			} else if _a.StatisticsPlay == _b.StatisticsPlay {
				return 0
			} else {
				return -1
			}
		})
		index := 0
		for _, u := range usd {
			if index > 10 {
				break
			}
			userdata := u.(*mgo.UserStatisticsData)
			CSData.Statistics = append(CSData.Statistics, &mgo.StatisticsData{
				UserId:     userdata.UserId,
				Name:       userdata.Name,
				Statistics: userdata.StatisticsPlay,
			})
			index++
		}
		err := mgo.AddClubPlayStatistics(CSData)
		if err != nil {
			tlog.Error("写入对局统计数据失败 err = " + err.Error())
			break
		}
		//再排名积分
		sort.Sort(usd, func(a interface{}, b interface{}) int8 {
			_a := a.(*mgo.UserStatisticsData)
			_b := b.(*mgo.UserStatisticsData)
			if _a.StatisticsIntegral < _b.StatisticsIntegral {
				return 1
			} else if _a.StatisticsIntegral == _b.StatisticsIntegral {
				return 0
			} else {
				return -1
			}
		})
		index = 0
		for _, u := range usd {
			if index > 10 {
				break
			}
			userdata := u.(*mgo.UserStatisticsData)
			CSData.Statistics = append(CSData.Statistics, &mgo.StatisticsData{
				UserId:     userdata.UserId,
				Name:       userdata.Name,
				Statistics: userdata.StatisticsIntegral,
			})
			index++
		}
		err = mgo.AddClubIntegralStatistics(CSData)
		if err != nil {
			tlog.Error("写入战绩数据库失败")
			break
		}
	}
}
