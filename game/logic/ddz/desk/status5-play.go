package desk

import (
	"game/logic/ddz/card"
	"game/pb/game/ddz"
	"time"

	"github.com/sirupsen/logrus"
)

func (d *desk) enterPlay() {
	d.loge.Info("start outcard")
	d.operNotif()
}

// 操作通知
func (d *desk) operNotif() {
	cuid := d.currPlayerID
	cuinfo := d.sdPlayers[d.currPlayer]

	freeOut := d.isFreeOut(cuid)
	operTime := timeOutCanOut

	var bigerLast bool
	if !freeOut {
		bigerLast = cuinfo.currCard.HaveBiger(d.lastGiveCard)
		if !bigerLast {
			operTime = timeOutCanNotOut
		}
	}

	operMask := mask(freeOut, bigerLast)

	if cuinfo.isTrustee {
		d.regTimeOutPlay(timeOutTrustee)

		// 托管的人就不用接收OperNotif了

		d.toOther(&pbgame_ddz.OperNotif{
			UserID:    cuid,
			Mask:      0,             // 其他人看不到动作
			Time:      timeOutCanOut, // 其他人看到都是能出牌的时间
			IsFreeOut: freeOut,
		}, cuid)

		return
	}

	d.regTimeOutPlay(operTime)

	d.toOne(&pbgame_ddz.OperNotif{
		UserID:    cuid,
		Mask:      operMask,
		Time:      uint32(operTime),
		IsFreeOut: freeOut,
		Seq:       d.seq,
	}, cuid)

	d.toOther(&pbgame_ddz.OperNotif{
		UserID:    cuid,
		Mask:      0,             // 其他人看不到动作
		Time:      timeOutCanOut, // 其他人看到都是能出牌的时间
		IsFreeOut: freeOut,
	}, cuid)
}

func mask(freeOut, bigerLast bool) (m int32) {
	if freeOut {
		m = int32(pbgame_ddz.OperMask_OmOut)
	} else {
		if bigerLast {
			m = int32(pbgame_ddz.OperMask_OmOut | pbgame_ddz.OperMask_OmHint | pbgame_ddz.OperMask_OmPass)
		} else {
			m = int32(pbgame_ddz.OperMask_OmNoout)
		}
	}
	return
}

func (d *desk) isFreeOut(uid uint64) bool {
	return uid == d.lastGiveCardPlayerID
}

// 出牌超时timer
func (d *desk) regTimeOutPlay(operTime int) {
	d.reqTime = time.Now().UTC()
	seq := d.seq
	d.timer = tw.AfterFunc(time.Duration(operTime)*time.Second, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.seq == seq {
			if d.arg.Type != gameTypFriend {
				d.systemPlay()
			}
		}
	})
}

// 系统出牌 [超时、托管]
func (d *desk) systemPlay() {
	uid := d.currPlayerID
	freeOut := d.isFreeOut(uid)
	cc := d.sdPlayers[d.currPlayer].currCard

	out := cc.TimeOutOut(freeOut, d.lastGiveCard)

	uob := &pbgame_ddz.UserOperBroadcast{
		UserID:    uid,
		IsFreeOut: freeOut,
		IsTrust:   true,
		LeftCount: uint32(cc.Len()),
	}

	f := logrus.Fields{"uid": uid}
	if out != nil {
		uob.Oper = pbgame_ddz.OperMask_OmOut
		uob.Ct = out.Type()
		uob.Cards = out.Dump()
		f["cards"] = out.String()
	} else {
		uob.Oper = pbgame_ddz.OperMask_OmPass
	}

	f["oper"] = uob.Oper.String()
	f["ct"] = uob.Ct.String()
	f["isfreeout"] = uob.IsFreeOut

	d.loge.Infof("outcard timeout %d %+v", uid, f)

	d.passCheck(uid, d.currPlayer, out, uob.Oper)

	d.userOperBroadcast(uob)

	d.checkPlay()
}

// 用户操作
func (d *desk) handleUserOper(uid uint64, req *pbgame_ddz.UserOper) {
	reqOutCard := card.NewSetCard(req.Cards)

	d.loge.Infof("handleUserOper status:%s curruid:%d requid:%d oper:%s ct:%s out:%s",
		d.f.Current(), d.currPlayerID, uid, req.Oper.String(), req.Ct.String(), reqOutCard.String())

	if !d.f.Is("SPlay") {
		return
	}

	if d.currPlayerID != uid {
		return
	}

	if d.seq != req.Seq {
		d.toOne(&pbgame_ddz.ErrNotif{ErrType: 1}, uid)
		return
	}

	freeOut := d.isFreeOut(uid)

	if req.Oper == pbgame_ddz.OperMask_OmNoout || req.Oper == pbgame_ddz.OperMask_OmPass {
		req.Ct = pbgame_ddz.CardType_CtUnknown
		req.Cards = nil
		if freeOut {
			d.loge.Warnf("need out card %d", uid)
			d.toOne(&pbgame_ddz.ErrNotif{ErrType: 1}, uid)
			return
		}
	} else if req.Oper == pbgame_ddz.OperMask_OmOut {
		if !d.outCardCheck(uid, req) {
			d.toOne(&pbgame_ddz.ErrNotif{ErrType: 1}, uid)
			return
		}
	} else if req.Oper == pbgame_ddz.OperMask_OmHint {
		if !freeOut { // 跟牌才提示
			tcs := &pbgame_ddz.TipsCards{}
			find := d.sdPlayers[d.currPlayer].currCard.Tips(d.lastGiveCard)
			for _, v := range find {
				tcs.Tips = append(tcs.Tips, &pbgame_ddz.TipsCard{Ct: v.Type(), Cards: v.Dump()})
			}
			d.toOne(tcs, uid)
		}
		return
	} else {
		return
	}

	d.passCheck(uid, d.currPlayer, reqOutCard, req.Oper)

	cc := d.sdPlayers[d.currPlayer].currCard
	d.userOperBroadcast(&pbgame_ddz.UserOperBroadcast{
		UserID:    uid,
		Oper:      req.Oper,
		Ct:        req.Ct,
		Cards:     req.Cards,
		IsFreeOut: freeOut,
		IsTrust:   false,
		LeftCount: uint32(cc.Len()),
	})

	d.checkPlay()
}

func (d *desk) userOperBroadcast(req *pbgame_ddz.UserOperBroadcast) {
	for _, v := range d.sdPlayers {
		req.Mul = d.mulUser(v.uid)
		d.toOne(req, v.uid)
	}
}

func (d *desk) outCardCheck(uid uint64, req *pbgame_ddz.UserOper) bool {
	if len(req.Cards) == 0 {
		return false
	}

	haveC := d.sdPlayers[d.currPlayer].currCard
	outC := card.NewSetCard(req.Cards)
	outCt := outC.Type()

	if outCt == pbgame_ddz.CardType_CtUnknown || outCt != req.Ct {
		d.loge.Warnf("bad card type uid:%d out:%s outCt:%s reqCt:%s", uid, outC.String(), outCt.String(), req.Ct.String())
		return false
	}

	if !haveC.Have(req.Cards) {
		d.loge.Warnf("not have card uid:%d out:%s have:%s ", uid, outC.String(), haveC.String())
		return false
	}

	if !d.isFreeOut(uid) && !outC.Biger(d.lastGiveCard) {
		d.loge.Warnf("not biger uid:%d out:%s last:%s", uid, outC.String(), d.lastGiveCard.String())
		return false
	}

	return true
}

func (d *desk) passCheck(uid uint64, uidx int, outC *card.SetCard, oper pbgame_ddz.OperMask) {
	d.seq++
	info := d.sdPlayers[uidx]
	info.lastOper = oper
	info.lastCard = outC

	if oper == pbgame_ddz.OperMask_OmOut && outC != nil && outC.Len() != 0 {
		cType := outC.Type()
		if cType == pbgame_ddz.CardType_CtJokers || cType == pbgame_ddz.CardType_CtBomb {
			d.mul.bomb++
		}

		info.lastCardType = cType

		d.lastGiveCardPlayerID = uid
		d.lastGiveCard = outC

		info.historyCard = append(info.historyCard, outC) // 记录出牌历史
		info.currCard.Del(outC.Dump())
	}
}

func (d *desk) checkPlay() {
	cc := d.sdPlayers[d.currPlayer].currCard

	if cc.IsEmpty() {
		if err := d.f.Event("play_end"); err != nil {
			d.loge.Error(err.Error())
		}
		d.enterCalc()
		return
	}

	d.turnNext()

	d.clearOneLastCard(d.currPlayerID)
	if d.lastGiveCardPlayerID == d.currPlayerID { // 新的一轮
		d.clearLastCard()
	}

	d.operNotif()
}

func (d *desk) clearOneLastCard(uid uint64) {
	for _, v := range d.sdPlayers {
		if v.uid == uid {
			v.lastOper = pbgame_ddz.OperMask_OmUnknown
			v.lastCard = nil
			v.lastCardType = pbgame_ddz.CardType_CtUnknown
			return
		}
	}
}

func (d *desk) clearLastCard() {
	for _, v := range d.sdPlayers {
		v.lastOper = pbgame_ddz.OperMask_OmUnknown
		v.lastCard = nil
		v.lastCardType = pbgame_ddz.CardType_CtUnknown
	}
}

// 用户获取倍数
func (d *desk) handleMultipleReq(uid uint64, req *pbgame_ddz.MultipleReq) {
	if !(d.f.Is("SCall") || d.f.Is("SRob") || d.f.Is("SPlay")) {
		return
	}

	m := &pbgame_ddz.MultipleRsp{
		CallRob: (1 << d.mul.callRob),
		Back:    d.mul.back,
		Bomb:    (1 << d.mul.bomb),
		Spring:  (1 << d.mul.spring),
	}

	for _, v := range d.sdPlayers {
		udi := &pbgame_ddz.UserDoubleInfo{UserID: v.uid}
		udi.D = d.mulUser(v.uid)
		m.Double = append(m.Double, udi)
	}

	d.toOne(m, uid)
}

// 用户托管
func (d *desk) handleUserTrustee(uid uint64, req *pbgame_ddz.UserTrustee) {
	if !(d.f.Is("SDouble") || d.f.Is("SPlay")) {
		return
	}

	d.changeTrustee(uid, req.IsTrustee)

	// 托管且正在等待此人出牌，则立即执行系统出牌
	if req.IsTrustee && uid == d.currPlayerID && d.f.Is("SPlay") {
		d.systemPlay()
	}
}

func (d *desk) changeTrustee(uid uint64, en bool) {
	for _, v := range d.sdPlayers {
		if v.uid == uid {
			v.isTrustee = en
			d.toSiteDown(&pbgame_ddz.TrusteeBroadcast{
				UserID:    uid,
				IsTrustee: en,
			})
			return
		}
	}
}

func (d *desk) handleChatReq(uid uint64, req *pbgame_ddz.ChatReq) {
	d.toSiteDown(&pbgame_ddz.ChatBroadcast{
		SenderUserID: req.SenderUserID,
		RecverUserID: req.RecverUserID,
		MsgID:        req.MsgID,
		FileURL:      req.FileURL,
	})
}
