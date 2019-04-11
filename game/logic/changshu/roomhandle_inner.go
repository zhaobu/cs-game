package main

import "go.uber.org/zap"

//玩家上下线
func (self *roomHandle) OnOffLine(uid uint64, online bool) {
	//检查玩家是否存在桌子
	d := getDeskByUID(uid)
	if d == nil {
		return
	}
	//检查玩家是否存在桌子信息
	if _, ok := d.deskPlayers[uid]; !ok {
		tlog.Error("OnOffLine() user in desk,but has no deskinfo", zap.Uint64("uid", uid))
		return
	}
	d.OnOffLine(uid, online)
	return
}
