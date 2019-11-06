package main

import (
	"context"
	"encoding/json"
	"fmt"
	"game/cache"
	"game/codec"
	pbcenter "game/pb/center"
	pbcommon "game/pb/common"
	pbinner "game/pb/inner"
	"game/util"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/go-redis/redis"
)

type attr struct {
	uid uint64
	t   time.Time // 进入匹配时间
}

type matchRoom struct {
	gameName string
	roomID   uint32
	complete chan map[uint64]struct{} // 匹配成功的
	mu       *sync.RWMutex
	players  map[uint64]*attr // 等待匹配的
}

func newMatchRoom(gameName string, roomID uint32) *matchRoom {
	r := matchRoom{}
	r.gameName = gameName
	r.roomID = roomID
	r.complete = make(chan map[uint64]struct{}, 1000)
	r.mu = &sync.RWMutex{}
	r.players = make(map[uint64]*attr)

	go r.doMatch()
	go r.doReadyGame()
	go r.doCheckLongTime()
	return &r
}

func (r *matchRoom) enter(uid uint64) pbcenter.MatchRspCode {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.players) > 50 {
		return pbcenter.MatchRspCode_Busy
	}

	r.players[uid] = &attr{uid: uid, t: time.Now().UTC()}
	return pbcenter.MatchRspCode_Queued
}

func (r *matchRoom) cancel(uid uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, find := r.players[uid]
	if !find {
		return false
	}

	delete(r.players, uid)
	cache.ExitMatch(uid)
	return true
}

func (r *matchRoom) doMatch() {
	const deskPlayerCnt = 3 // TODO
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			deskCnt := 0
			oneDesk := make(map[uint64]struct{})

			r.mu.Lock()
			for k := range r.players {
				oneDesk[k] = struct{}{}

				if len(oneDesk) == deskPlayerCnt {
					for v := range oneDesk {
						delete(r.players, v)
					}
					r.complete <- oneDesk
					oneDesk = make(map[uint64]struct{})

					deskCnt++
					if deskCnt >= 10 {
						break
					}
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *matchRoom) doReadyGame() {
	for uids := range r.complete {
		cli, err := getGameCli(r.gameName)
		if err != nil {
			clearMatchStatus(uids)
			continue
		}

		gameMatchSucc := &pbinner.GameMatchSucc{}
		for uid := range uids {
			gameMatchSucc.UserIDs = append(gameMatchSucc.UserIDs, uid)
		}
		gameMatchSucc.RoomId = r.roomID

		req := &codec.Message{}
		codec.Pb2Msg(gameMatchSucc, req)

		tlog.Info("GameMatchSucc", zap.Any("uids", gameMatchSucc.UserIDs), zap.String("gamename", r.gameName), zap.Uint32("roomid", r.roomID))
		_, err = cli.Go(context.Background(), "GameMatchSucc", req, nil, nil)
		if err != nil {
			clearMatchStatus(uids)
			tlog.Error(err.Error())
		}
	}
}

func (r *matchRoom) doCheckLongTime() {
	tick := time.NewTicker(time.Second * 30)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			r.deleteLongTime()
		}
	}
}

func (r *matchRoom) deleteLongTime() {
	deleted := make(map[uint64]struct{})

	r.mu.Lock()
	for _, v := range r.players {
		if time.Now().UTC().Sub(v.t) > time.Second*30 {
			delete(r.players, v.uid)
			deleted[v.uid] = struct{}{}
		}
	}
	r.mu.Unlock()

	for uid := range deleted {
		cache.ExitMatch(uid)

		msg := &codec.Message{}
		err := codec.Pb2Msg(&pbcenter.MatchTimeOut{}, msg)
		if err != nil {
			continue
		}

		var xx struct {
			Msg  *codec.Message
			Uids []uint64
		}
		xx.Msg = msg
		xx.Uids = append(xx.Uids, uid)

		data, err := json.Marshal(xx)
		if err != nil {
			continue
		}

		_, err = util.RedisXadd(redisCli, "backend_to_gate", msg.Name, data)
		if err != nil {
			tlog.Error(err.Error())
		}
	}
	return
}

var (
	waiter = make(map[string]*matchRoom) // key: GameName
	mu     sync.RWMutex

	redisCli *redis.Client
)

func enterMatch(uid uint64, gn string, rid uint32, matchRsp *pbcenter.MatchRsp) {
	code, inStatus, inGameName, inRoomID, err := cache.EnterMatch(uid, gn, rid)
	if err != nil {
		matchRsp.Code = pbcenter.MatchRspCode_InternalServerError
		matchRsp.StrCode = err.Error()
		tlog.Error(err.Error())
		return
	}

	matchRsp.Sec = int32(rand.Intn(60) + 30)
	matchRsp.Status = pbcommon.UserStatus(pbcommon.UserStatus_value[inStatus])
	matchRsp.GameName = inGameName
	matchRsp.RoomID = inRoomID

	if code == 3 {
		matchRsp.Code = pbcenter.MatchRspCode_StatuErr
		return
	} else if code == 2 {
		matchRsp.Code = pbcenter.MatchRspCode_Queued
		return
	}

	roomName := fmt.Sprintf("%s_%d", gn, rid)

	mu.Lock()
	defer mu.Unlock()

	if waiter[roomName] == nil {
		waiter[roomName] = newMatchRoom(gn, rid)
	}

	matchRsp.Code = waiter[roomName].enter(uid)
	return
}

func clearMatchStatus(uids map[uint64]struct{}) {
	for uid := range uids {
		cache.ExitMatch(uid)
	}
}
