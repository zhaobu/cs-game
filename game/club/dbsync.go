package main

import (
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"time"
)

func loadDB() {
	rsp, err := mgo.QueryAllClub()
	if err != nil {
		return
	}

	for _, c := range rsp {
		if c.Members == nil {
			c.Members = make(map[uint64]*mgo.ClubMember)
		}
		cc := &cacheClub{Club: c, desks: make(map[uint64]*pbcommon.DeskInfo)}
		clubMgr[c.ID] = cc

		for _, m := range c.Members {
			addUserJoinClub(m.UserID, c.ID)
		}
	}
}

func syncDB() {
	go func() {
		tick := time.NewTicker(time.Second * 1)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				muClubMgr.RLock()
				for _, v := range clubMgr {
					v.Lock()
					if v.noCommit {
						mgo.SaveClub(v.Club)
						v.noCommit = false
					}
					v.Unlock()
				}
				muClubMgr.RUnlock()
			}
		}
	}()
}
