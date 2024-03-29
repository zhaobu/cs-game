package main

import "sync"

var (
	muId2Desk   sync.RWMutex
	id2desk     = make(map[uint64]*Desk)
	muUser2desk sync.RWMutex
	user2desk   = make(map[uint64]*Desk)
)

func getDeskByID(deskId uint64) *Desk {
	muId2Desk.RLock()
	defer muId2Desk.RUnlock()

	return id2desk[deskId]
}

func updateID2desk(d *Desk) bool {
	muId2Desk.Lock()
	defer muId2Desk.Unlock()

	_, find := id2desk[d.deskId]
	if find {
		return false
	}
	id2desk[d.deskId] = d
	return true
}

func deleteID2desk(deskID uint64) {
	muId2Desk.Lock()
	defer muId2Desk.Unlock()

	delete(id2desk, deskID)
}

func getDeskByUID(uid uint64) *Desk {
	muUser2desk.RLock()
	defer muUser2desk.RUnlock()

	return user2desk[uid]
}

func updateUser2desk(d *Desk, uids ...uint64) {
	muUser2desk.Lock()
	defer muUser2desk.Unlock()

	for _, uid := range uids {
		user2desk[uid] = d
	}
}

func deleteUser2desk(uids ...uint64) {
	muUser2desk.Lock()
	defer muUser2desk.Unlock()

	for _, uid := range uids {
		delete(user2desk, uid)
	}
}
