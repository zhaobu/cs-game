package main

import (
	"cy/game/db/mgo"
	"sync"
)

type userOtherInfo struct {
	sync.RWMutex
	UserID   uint64
	UserName string
	Profile  string
	Online   int32
	subFlag  bool
}

var (
	muUser            sync.RWMutex
	users             = make(map[uint64]*userOtherInfo)
	muUserJoinedClubs sync.RWMutex
	userJoinedClubs   = make(map[uint64]map[int64]struct{})
)

func mustGetUserOther(uid uint64) *userOtherInfo {
	uinfo, err := mgo.QueryUserInfo(uid)

	muUser.Lock()
	defer muUser.Unlock()

	u, find := users[uid]
	if find {
		return u
	}

	u = &userOtherInfo{
		UserID: uid,
	}

	if err == nil {
		u.UserName = uinfo.Name
		u.Profile = uinfo.Profile
	}

	users[uid] = u
	return u
}

// joinedclub
func getUserJoinedClubs(uid uint64) map[int64]struct{} {
	muUserJoinedClubs.RLock()
	defer muUserJoinedClubs.RUnlock()
	return userJoinedClubs[uid]
}

func addUserJoinClub(uid uint64, cid int64) {
	muUserJoinedClubs.Lock()
	if userJoinedClubs[uid] == nil {
		userJoinedClubs[uid] = make(map[int64]struct{})
	}
	userJoinedClubs[uid][cid] = struct{}{}
	muUserJoinedClubs.Unlock()
}

func delUserJoinClub(uid uint64, cid int64) {
	muUserJoinedClubs.Lock()
	delete(userJoinedClubs[uid], cid)
	muUserJoinedClubs.Unlock()
}
