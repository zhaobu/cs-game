package main

import (
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"sync"
)

type cacheClub struct {
	sync.RWMutex
	*mgo.Club
	desks    map[uint64]*pbcommon.DeskInfo
	noCommit bool
}

var (
	muClubMgr sync.RWMutex
	clubMgr   = make(map[int64]*cacheClub)
)

func newCacheClub() *cacheClub {
	c := &cacheClub{
		Club:  &mgo.Club{Members: make(map[uint64]*mgo.ClubMember)},
		desks: make(map[uint64]*pbcommon.DeskInfo),
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
