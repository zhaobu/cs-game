package main

import (
	"game/db/mgo"
	"game/pb/common"
	"sync"
	"time"
)

type cacheClub struct {
	sync.RWMutex
	*mgo.Club
	desks    map[uint64]*pbcommon.DeskInfo
	f        func()
	noCommit bool
	lastquerytime time.Time		//上一次查询时间 用户俱乐部桌子同步处理
}

var (
	muClubMgr sync.RWMutex
	clubMgr   = make(map[int64]*cacheClub)
)

func newCacheClub() *cacheClub {
	c := &cacheClub{
		Club:  &mgo.Club{Members: make(map[uint64]*mgo.ClubMember)},
		desks: make(map[uint64]*pbcommon.DeskInfo),
		lastquerytime:time.Now(),
	}
	return c
}

func addClub(cc *cacheClub) {
	muClubMgr.Lock()
	clubMgr[cc.ID] = cc
	muClubMgr.Unlock()
}

func getClub(cid int64) *cacheClub {
	muClubMgr.RLock()
	defer muClubMgr.RUnlock()
	return clubMgr[cid]
}

func delClub(cid int64) {
	muClubMgr.Lock()
	delete(clubMgr, cid)
	muClubMgr.Unlock()
}
