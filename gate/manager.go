package main

import (
	"sync"

	"go.uber.org/zap"
)

const sessionMapNum = 32

type manager struct {
	sessionMaps [sessionMapNum]sessionMap
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*session // key uid
}

func newManager() *manager {
	manager := &manager{}
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*session)
	}
	return manager
}

func (manager *manager) GetSession(uid uint64) (sess *session, findOk bool) {
	smap := &manager.sessionMaps[uid%sessionMapNum]
	smap.RLock()
	sess, findOk = smap.sessions[uid]
	smap.RUnlock()
	return
}

func (manager *manager) SetSession(sess *session) {
	smap := &manager.sessionMaps[sess.uid%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	old, ok := smap.sessions[sess.uid]
	if ok {

		tlog.Info("set session", zap.Uint64("uid", sess.uid))
		old.repeatLogin(sess)
		old.stop()
	}
	smap.sessions[sess.uid] = sess
}

func (manager *manager) DelSessionByUID(uid uint64) {
	smap := &manager.sessionMaps[uid%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	old, ok := smap.sessions[uid]
	if ok {
		delete(smap.sessions, uid)
		old.stop()
	}
}

func (manager *manager) Iter(f func(uid uint64, sess *session)) {
	for idx := range manager.sessionMaps {
		manager.sessionMaps[idx].RLock()
		for uid, sess := range manager.sessionMaps[idx].sessions {
			go f(uid, sess)
		}
		manager.sessionMaps[idx].RUnlock()
	}
}
