package desk

import (
	"cy/game/cache"
	"cy/game/db/mgo"
	"cy/game/pb/game/ddz"
	"cy/game/pb/hall"
	"sort"
	"time"
)

func (d *desk) enterCalc() {
	d.loge.Info("start calc")

	landlordWin := (d.currPlayerID == d.landlord)
	d.checkSpring(landlordWin)
	d.calc(landlordWin)
	d.gameOverInfo(1)

	if d.arg.Type == gameTypMatch {
		d.endGame()
		return
	}

	if d.arg.Type == gameTypFriend {
		d.flashWarRecord()

		if d.currLoopCnt >= d.arg.LoopCnt {
			d.toSiteDown(&d.warRecord)
			d.endGame()
			return
		}

		d.resetUserStatus()

		d.userGameStatusBroadcast()

		if err := d.f.Event("calc_end"); err != nil {
			d.loge.Error(err.Error())
		}

		d.currLoopCnt++
		d.enterWait()
	}
}

func (d *desk) flashWarRecord() {
	if len(d.warRecord.Rank) == 0 {
		for _, v := range d.sdPlayers {
			d.warRecord.Rank = append(d.warRecord.Rank, &pbgame_ddz.RankInfo{UserID: v.uid})
		}
	}

	if len(d.warRecord.Detail) > 0 {
		if d.warRecord.Detail[len(d.warRecord.Detail)-1].RoundID == d.currLoopCnt {
			return
		}
	}

	oneRound := &pbgame_ddz.OneRound{RoundID: d.currLoopCnt}
	for _, v := range d.sdPlayers {
		oneRound.Change = append(oneRound.Change, &pbgame_ddz.WealthChange{UserID: v.uid, Change: v.change})

		for _, v2 := range d.warRecord.Rank {
			if v2.UserID == v.uid {
				v2.RoundSum++
				if v.isWin {
					v2.WinSum++
				}
				v2.ChangeSum += v.change
				break
			}
		}
	}
	d.warRecord.Detail = append(d.warRecord.Detail, oneRound)

	sort.Slice(d.warRecord.Rank, func(i, j int) bool {
		return d.warRecord.Rank[i].ChangeSum > d.warRecord.Rank[j].ChangeSum
	})

	for idx, v := range d.warRecord.Rank {
		v.Order = uint32(idx)
	}
}

func (d *desk) endGame() {
	if err := d.f.Event("game_end"); err != nil {
		d.loge.Error(err.Error())
	}
	d.enterEnd()
}

func (d *desk) handleQueryWarRecord(uid uint64, _ *pbgame_ddz.QueryWarRecord) {
	d.toSiteDown(&d.warRecord)
}

func (d *desk) handleUserProposeBreakGame(uid uint64, _ *pbgame_ddz.UserProposeBreakGame) {
	if d.arg.Type != gameTypFriend || !d.isSitDownUser(uid) {
		return
	}

	if !(d.f.Is("SCall") || d.f.Is("SRob") || d.f.Is("SDouble") || d.f.Is("SPlay")) {
		return
	}

	if d.voteStartUserID != 0 {
		return
	}
	d.voteStartUserID = uid
	d.breakGameVoteStartTime = time.Now().UTC()

	for _, v := range d.sdPlayers {
		if v.uid == uid {
			v.agreeBreakGame = 1
			break
		}
	}

	d.toSiteDown(&pbgame_ddz.BreakGameVoteStart{UserID: uid, Time: timeOutBreakGameVote})

	d.breakGameTimer = time.AfterFunc(time.Second*timeOutBreakGameVote, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		for _, v := range d.sdPlayers {
			if v.agreeBreakGame == 0 { // 默认为同意
				v.agreeBreakGame = 1
				d.toSiteDown(&pbgame_ddz.BreakGameVoteBroadcast{UserID: v.uid, Agree: true})
			}
		}

		d.breakGameVoteEnd()
	})
}

func (d *desk) breakGameVoteEnd() {
	canBreakGame := true
	for _, v := range d.sdPlayers {
		if v.agreeBreakGame == 2 { // 有任意一人反对 就不能解散
			canBreakGame = false
			break
		}
	}

	voteEnd := &pbgame_ddz.BreakGameVoteEnd{}
	if canBreakGame {
		voteEnd.Code = 1
	} else {
		voteEnd.Code = 2
	}
	d.toSiteDown(voteEnd)

	if voteEnd.Code == 1 {
		if d.arg.Type == gameTypFriend && d.currLoopCnt == 1 { // 第1局还没打完，就解散了，预扣的要还回去
			d.preSureCreater()
		}

		d.gameOverInfo(2)
		d.flashWarRecord()
		d.toSiteDown(&d.warRecord)
		d.enterEnd()
		return
	}

	// clear
	d.voteStartUserID = 0
	for _, v := range d.sdPlayers {
		v.agreeBreakGame = 0
	}
}

// 预扣还回去
func (d *desk) preSureCreater() {
	change := int64(0)
	goldChange := int64(0)
	masonryChange := int64(0)

	if d.arg.PaymentType == 1 {
		change = int64(d.arg.Fee)
	} else if d.arg.PaymentType == 2 {
		change = int64(d.arg.Fee / seatNumber)
	}

	if d.arg.FeeType == 1 {
		goldChange = change
	} else if d.arg.FeeType == 2 {
		masonryChange = change
	}

	info, err := mgo.UpdateWealthPreSure(d.createUserID, d.arg.FeeType, change)
	if err != nil {
		return
	}

	// 通知变化
	toGateNormal(d.loge, &pbhall.UserWealthChange{
		UserID:        d.createUserID,
		Gold:          info.Gold,
		GoldChange:    goldChange,
		Masonry:       info.Masonry,
		MasonryChange: masonryChange,
	}, d.createUserID)
}

func (d *desk) handleUserBreakGameVote(uid uint64, req *pbgame_ddz.UserBreakGameVote) {
	if d.voteStartUserID == 0 {
		return
	}

	for _, v := range d.sdPlayers {
		if v.uid == uid && v.agreeBreakGame == 0 {
			if req.Agree {
				v.agreeBreakGame = 1
			} else {
				v.agreeBreakGame = 2
			}
			d.toSiteDown(&pbgame_ddz.BreakGameVoteBroadcast{UserID: uid, Agree: req.Agree})

			votedCnt := 0
			for _, v := range d.sdPlayers {
				if v.agreeBreakGame != 0 {
					votedCnt++
				}
			}
			if votedCnt == seatNumber || !req.Agree {
				d.breakGameTimer.Stop()
				d.breakGameVoteEnd()
			}
			break
		}
	}
}

func (d *desk) enterEnd() {
	d.clearSdUser()
	deleteID2desk(d.id)

	cache.DeleteClubDeskRelation(d.id)
	cache.DelDeskInfo(d.id)
	cache.FreeDeskID(d.id)

}

func (d *desk) clearSdUser() {
	for dir, v := range d.sdPlayers {
		delete(d.sdPlayers, dir)
		deleteUser2desk(v.uid)
		cache.ExitGame(v.uid, gameName, gameID, d.id)
	}
}

func (d *desk) resetUserStatus() {
	for idx, v := range d.sdPlayers {
		d.sdPlayers[idx] = &playerInfo{
			uid:    v.uid,
			status: pbgame_ddz.UserGameStatus_UGSSitDown,
			info:   v.info,
			dir:    idx,
		}
	}
}

func (d *desk) userGameStatusBroadcast() {
	for _, v := range d.sdPlayers {
		d.toSiteDown(&pbgame_ddz.UserGameStatusBroadcast{
			UserID: v.uid,
			Status: v.status,
		})
	}
}

func (d *desk) checkSpring(landlordWin bool) {
	isSpring := uint32(1)

	if landlordWin {
		for _, v := range d.sdPlayers {
			if v.uid != d.landlord {
				if len(v.historyCard) != 0 {
					isSpring = 0
					break
				}
			}
		}
	} else {
		for _, v := range d.sdPlayers {
			if v.uid == d.landlord {
				if len(v.historyCard) > 1 {
					isSpring = 0
				}
				break
			}
		}
	}

	d.mul.spring = isSpring
}

func (d *desk) calc(landlordWin bool) {
	if landlordWin {
		for _, v := range d.sdPlayers {
			v.isWin = (v.uid == d.landlord)
		}
	} else {
		for _, v := range d.sdPlayers {
			v.isWin = (v.uid != d.landlord)
		}
	}

	for _, v := range d.sdPlayers {
		mul := d.mulUser(v.uid)
		if v.uid == d.landlord {
			mul *= 2
		}

		v.change = int64(d.arg.BaseScore * mul)
		if !v.isWin {
			v.change *= -1
		}

		if d.arg.Type == gameTypMatch {
			info, err := mgo.UpdateWealth(v.uid, d.arg.FeeType, v.change)
			if err != nil {
				d.loge.Error(err.Error())
				continue
			}

			v.info = info
		}
	}

	// 好友场是在第一局结束时扣除费用
	if d.arg.Type == gameTypFriend && d.currLoopCnt == 1 {
		// 预扣的先还原
		d.preSureCreater()

		// 实际扣
		if d.arg.PaymentType == 1 {
			_, err := mgo.UpdateWealth(d.createUserID, d.arg.FeeType, int64(d.arg.Fee)*-1)
			if err != nil {

			} else {

			}

		} else if d.arg.PaymentType == 2 {
			for _, v := range d.sdPlayers {
				info, err := mgo.UpdateWealth(v.uid, d.arg.FeeType, int64(d.arg.Fee/seatNumber)*-1)
				if err != nil {
					continue
				}
				v.info = info
			}
		}
	}
}

func (d *desk) gameOverInfo(endType uint32) {
	g := &pbgame_ddz.GameOverInfo{}
	g.EndType = endType
	g.IsSpring = (d.mul.spring == 1)
	for _, v := range d.sdPlayers {
		uei := &pbgame_ddz.UserEndInfo{
			UserID:    v.uid,
			EndStatus: 0, // TODO
			Cards:     v.currCard.Dump(),
			FeeType:   d.arg.FeeType,
			Change:    v.change,
			Mul:       d.mulUser(v.uid),
		}

		if d.arg.FeeType == 1 {
			uei.Curr = v.info.Gold
		} else if d.arg.FeeType == 2 {
			uei.Curr = v.info.Masonry
		}

		g.User = append(g.User, uei)
	}
	d.toSiteDown(g)
}
