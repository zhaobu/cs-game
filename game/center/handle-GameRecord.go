package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	pbcommon "cy/game/pb/common"
	pbgamerecord "cy/game/pb/gamerecord"
	sort "cy/game/util/tools/Sort"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ClubStatistics *ClubGameStatistics
	clublock       sync.RWMutex
)

type ClubGameStatistics struct {
	ClubSD map[int64]*ClubStatisticsData
}

type ClubStatisticsData struct {
	UserSD map[uint64]*UserStatisticsData
}

type UserStatisticsData struct {
	UserId             uint64 //用户Id
	Name               string //姓名
	StatisticsIntegral int64  //当天输赢积分统计
	StatisticsPlay     int64  //当天本俱乐部次数统计
}

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
	log.Debugf("QueryRoomRecordReq请求参数:req:%v", req)
	rsp := &pbgamerecord.QueryRoomRecordRsp{
		Error: 0,
	}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	querydata := []*mgo.RoomRecord{}
	if req.QueryType == 1 { //按userId查询
		querydata, err = mgo.QueryUserRoomRecord(req.QueryUserId, req.QueryStartTime, req.QueryEndTime, req.CurPage, req.Limit)
		if err != nil {
			tlog.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
		}
	} else if req.QueryType == 2 { //按俱乐部id查询
		querydata, err = mgo.QueryClubRoomRecord(req.QueryClubId, req.QueryStartTime, req.QueryEndTime, req.CurPage, req.Limit)
		if err != nil {
			tlog.Warn("查询用户数据失败 err = " + err.Error())
			rsp.Error = 1
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
			RoomRule:      &pbgamerecord.GameAction{ActName: v.RoomRule.ActName, ActValue: v.RoomRule.ActValue},
			GamePlayers:   make([]*pbgamerecord.RoomPlayerInfo, 0, len(v.GamePlayers)),
			GameRecordIds: v.GameRecords,
		}
		for _, v1 := range v.GamePlayers {
			_data.GamePlayers = append(_data.GamePlayers, &pbgamerecord.RoomPlayerInfo{
				UserId:     v1.UserId,
				Name:       v1.Name,
				TotalScore: v1.TotalScore,
				ChairId:    v1.ChairId,
			})
		}
		rsp.Datas = append(rsp.Datas, _data)
	}
	err = codec.Pb2Msg(rsp, reply)
	if err != nil {
		tlog.Error("消息封装失败", zap.Error(err))
	}
	log.Debugf("QueryRoomRecordReq请求结果:rsp:%v", rsp)
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
	log.Debugf("QueryGameRecordReq请求参数req:%v", req)
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
				UserId:  v1.UserId,
				ChairId: v1.ChairId,
				Score:   v1.Score,
			})
		}
		rsp.Records = append(rsp.Records, _data)
	}

	if err = codec.Pb2Msg(rsp, reply); err != nil {
		tlog.Error("消息封装失败", zap.Error(err))
	}
	log.Debugf("QueryGameRecordReq请求结果:rsp:%v", rsp)
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
	log.Debugf("QueryGameRePlayRecord请求参数:req:%v", req)
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
	log.Debugf("QueryGameRePlayRecord请求结果:rsp:%v", rsp)
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

		//先做对局次数排名
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
			rsp.StatisticsDatas = append(rsp.StatisticsDatas, &pbgamerecord.StatisticsData{
				UserId:     sdata.UserId,
				Name:       sdata.Name,
				Statistics: sdata.Statistics,
			})
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
func WriteClubGameStatistics(gr *mgo.WirteRecord) {
	defer clublock.Unlock()
	clublock.Lock()
	if gr.CreateInfo.ClubId == 0 {
		return
	}
	if ClubStatistics == nil {
		ClubStatistics = &ClubGameStatistics{
			ClubSD: make(map[int64]*ClubStatisticsData),
		}
	}
	if _, ok := ClubStatistics.ClubSD[gr.CreateInfo.ClubId]; !ok {
		ClubStatistics.ClubSD[gr.CreateInfo.ClubId] = &ClubStatisticsData{
			UserSD: make(map[uint64]*UserStatisticsData),
		}
	}
	for _, v := range gr.CurGameInfo.GamePlayers {
		if _, ok := ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId]; !ok {
			ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId] = &UserStatisticsData{
				UserId:             v.UserId,
				Name:               v.Name,
				StatisticsIntegral: 0,
				StatisticsPlay:     0,
			}
		}
		ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId].StatisticsIntegral += int64(v.Score)
		ClubStatistics.ClubSD[gr.CreateInfo.ClubId].UserSD[v.UserId].StatisticsPlay++
	}
}

//启动定时任务 暂定每天凌晨执行统计数据清零
func StartTimer_ResetClubStatistics() {
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
func ResetClubGameStatistics() {
	//if ClubStatistics == nil {
	//	return
	//}
	//_ClubStatistics := ClubStatistics
	//ClubStatistics = nil
	cds, err := mgo.QueryAndClearAllClubCurrDayStatistics()
	if err != nil {
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
				_a := a.(*UserStatisticsData)
				_b := b.(*UserStatisticsData)
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
				userdata := u.(*UserStatisticsData)
				CSData.Statistics = append(CSData.Statistics, &mgo.StatisticsData{
					UserId:     userdata.UserId,
					Statistics: userdata.StatisticsPlay,
				})
				index++
			}
			err := mgo.AddClubPlayStatistics(CSData)
			if err != nil {
				log.Errorf("写入对局统计数据失败 err = " + err.Error())
				break
			}

			//再排名积分
			sort.Sort(usd, func(a interface{}, b interface{}) int8 {
				_a := a.(*UserStatisticsData)
				_b := b.(*UserStatisticsData)
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
				userdata := u.(*UserStatisticsData)
				CSData.Statistics = append(CSData.Statistics, &mgo.StatisticsData{
					UserId:     userdata.UserId,
					Name:       userdata.Name,
					Statistics: userdata.StatisticsIntegral,
				})
				index++
			}
			err = mgo.AddClubIntegralStatistics(CSData)
			if err != nil {
				log.Errorf("写入战绩数据库失败")
				break
			}
		}
	} else {
		log.Errorf("查询当日俱乐部游戏统计数据错误", err)
	}
	StartTimer_ResetClubStatistics() //重新计算下一次时间
}
