package main

import (
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	"cy/game/pb/common"
	"cy/game/pb/game"
	"cy/game/pb/inner"

	"bytes"
	"context"

	"github.com/sirupsen/logrus"
)

func flashDesk(n *pbinner.DeskChangeNotif) {
	logrus.Infof("DeskChangeNotif %+v\n", n)

	cc := getClub(n.ClubID)
	if cc == nil {
		logrus.Warnf("can not find club %d", n.ClubID)
		return
	}

	// 更新cache_desk
	var deskInfo *pbcommon.DeskInfo

	switch n.ChangeTyp {
	case 1, 2:
		var err error
		deskInfo, err = cache.QueryDeskInfo(n.DeskID)
		if err != nil {
			logrus.Warnf("QueryDeskInfo %s", err.Error())
			return
		}
		cc.Lock()
		cc.desks[deskInfo.ID] = deskInfo
		cc.Unlock()
	case 3:
		cc.Lock()
		delete(cc.desks, n.DeskID)
		cc.Unlock()
	}

	cc.RLock()
	isAutoCreate := cc.IsAutoCreate
	cc.RUnlock()

	if isAutoCreate {
		// 更新、删除的时候 可能要自动建房间
		var needCheck bool
		if n.ChangeTyp == 3 {
			needCheck = true
		} else if n.ChangeTyp == 2 && deskInfo.Status == "2" {
			needCheck = true
		}
		if needCheck {
			checkAutoCreate(n.ClubID)
		}
	}

	go sendClubChangeInfo(n.ClubID, clubChangeTypUpdate, 0)

}

func checkAutoCreate(cid int64) {
	logrus.Infof("checkAutoCreate %d", cid)
	var needCreateGameArgs []*mgo.DeskSetting

	cc := getClub(cid)
	if cc == nil {
		return
	}

	cc.RLock()

	var gameArgs []*mgo.DeskSetting
	for _, a := range cc.GameArgs {
		if a.Enable {
			gameArgs = append(gameArgs, a)
		}
	}

	for _, a := range gameArgs {
		var have bool
		for _, b := range cc.desks {
			if a.GameName == b.GameName && bytes.Equal(a.GameArgMsgValue, b.ArgValue) {
				have = true
				break
			}
		}

		if !have {
			needCreateGameArgs = append(needCreateGameArgs, a)
			if pb, err := protobuf.Unmarshal(a.GameArgMsgName, a.GameArgMsgValue); err == nil {
				logrus.Infof("will create desk arg: %+v\n", pb)
			}
		}
	}

	masterUserID := cc.MasterUserID

	cc.RUnlock()

	if len(needCreateGameArgs) > 0 {
		createDesk(needCreateGameArgs, cid, masterUserID)
	}

}

func createDesk(setting []*mgo.DeskSetting, cid int64, masterUserID uint64) {
	logrus.Infof("%d createDesk %d", cid, len(setting))

	for _, s := range setting {
		cli, err := getGameCli(s.GameName)
		if err != nil {
			continue
		}

		reqRCall := &codec.Message{}
		codec.Pb2Msg(&pbgame.MakeDeskReq{
			Head:            &pbcommon.ReqHead{UserID: masterUserID},
			GameName:        s.GameName,
			GameArgMsgName:  s.GameArgMsgName,
			GameArgMsgValue: s.GameArgMsgValue,
			ClubID:          cid,
		}, reqRCall)
		reqRCall.UserID = masterUserID
		rspRCall := &codec.Message{}

		cli.Go(context.Background(), "MakeDeskReq", reqRCall, rspRCall, nil)
	}
}
