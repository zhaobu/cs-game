package main

import (
	"game/cache"
	"game/db/mgo"
	pbcommon "game/pb/common"
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
		if ds, err := cache.QueryClubDeskInfo(c.ID); err == nil {
			for _, d := range ds {
				cc.desks[d.ID] = d
			}
		}
		clubMgr[c.ID] = cc

		for _, m := range c.Members {
			addUserJoinClub(m.UserID, c.ID)
		}
	}
}

//游戏服务器异常之后 需要重新同步桌子信息
func synchroClubdeskinfo(ClubID int64) {
	if v, ok := clubMgr[ClubID]; ok {
		v.desks = make(map[uint64]*pbcommon.DeskInfo)
		if ds, err := cache.QueryClubDeskInfo(v.ID); err == nil {
			for _, d := range ds {
				v.desks[d.ID] = d
			}
		}
	}
}

func syncDB() {
	go func() {
		tick := time.NewTicker(time.Second * 3)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				muClubMgr.RLock()
				for _, v := range clubMgr {
					v.Lock()
					if v.noCommit {
						err := mgo.SaveClub(v.Club)
						if err == nil && v.f != nil {
							go v.f()
							v.f = nil
						}
						v.noCommit = false
					}
					v.Unlock()
				}
				muClubMgr.RUnlock()
			}
		}
	}()
}
