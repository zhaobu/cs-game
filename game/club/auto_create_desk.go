package main

import (
	"context"
	"game/cache"
	"game/codec"
	"game/codec/protobuf"
	"game/db/mgo"
	pbcommon "game/pb/common"
	pbgame "game/pb/game"
	pbinner "game/pb/inner"
	"hash/crc32"

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
	IsProofe := cc.IsProofe
	cc.RUnlock()

	if isAutoCreate && !IsProofe {
		// 更新、删除的时候 可能要自动建房间
		var needCheck bool
		if n.ChangeTyp == 3 {
			needCheck = true
		} else if n.ChangeTyp == 2 && deskInfo.Status == "2" {
			needCheck = true
		}
		if needCheck {
			setting, cid, masterUserID := checkAutoCreate(n.ClubID)
			if len(setting) > 0 {
				createDesk(setting, cid, masterUserID)
			}
		}
	}

	go sendClubChangeInfo(n.ClubID, clubChangeTypUpdateNoTips, 0)
}

func checkAutoCreate(_cid int64) (setting []*mgo.DeskSetting, cid int64, masterUserID uint64) {
	logrus.Infof("checkAutoCreate %d", _cid)
	var needCreateGameArgs []*mgo.DeskSetting

	cc := getClub(_cid)
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
		have := false
		//fmt.Printf("开始查找 a.GameName=%s  a.GameArgMsgValue = %v a.CreateVlaueHash=%d\n", a.GameName,a.GameArgMsgValue,crc32.ChecksumIEEE(a.GameArgMsgValue))
		for _, b := range cc.desks {
			//fmt.Printf("检测俱乐部房间 a.GameName=%s b.GameName=%s\n", a.GameName, b.GameName)
			//fmt.Printf("检测俱乐部房间 a.Status=%s \n",b.Status)
			//fmt.Printf("检测俱乐部房间\n a.CreateVlaueHash =%d  b.CreateVlaueHash =%d  \n",crc32.ChecksumIEEE(a.GameArgMsgValue),b.CreateVlaueHash)
			//fmt.Printf("a GameArgMsgValue = %d a CreateVlaueHash = %d \n",crc32.ChecksumIEEE(a.GameArgMsgValue),a.GameArgMsgValue)
			if a.GameName == b.GameName && b.Status == "1" && b.CreateVlaueHash == uint64(crc32.ChecksumIEEE(a.GameArgMsgValue)) { //判断是否有空的对应玩法的桌子
				//fmt.Printf("找到一个相同的空房间 %d\n",b.CreateVlaueHash)
				have = true
				break
			}
		}

		if !have {
			//fmt.Printf("需要创建一个房间 %d \n",crc32.ChecksumIEEE(a.GameArgMsgValue))
			needCreateGameArgs = append(needCreateGameArgs, a)
			if pb, err := protobuf.Unmarshal(a.GameArgMsgName, a.GameArgMsgValue); err == nil {
				logrus.Infof("will create desk arg: %+v\n", pb)
			}
		}
		//fmt.Printf("结束查找-----------------------------------------------------------------------------------\n")
	}

	defer cc.RUnlock()
	//if len(needCreateGameArgs) > 0 {
	//	createDesk(needCreateGameArgs, _cid, masterUserID)
	//}
	return needCreateGameArgs, _cid, cc.MasterUserID
}

func createDesk(setting []*mgo.DeskSetting, cid int64, masterUserID uint64) {
	logrus.Infof("%d createDesk %d", cid, len(setting))
	go func() {
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
				ClubMasterUid:   masterUserID,
			}, reqRCall)
			reqRCall.UserID = masterUserID
			rspRCall := &codec.Message{}
			cli.Call(context.Background(), "MakeDeskReq", reqRCall, rspRCall)
			// time.Sleep(time.Millisecond * 10)
		}
	}()
}

//销毁桌子
func destoryDesk(uId uint64, desks ...*pbcommon.DeskInfo) {
	for _, s := range desks {
		cli, err := getGameCli(s.GameName)
		if err != nil {
			continue
		}
		reqRCall := &codec.Message{}
		codec.Pb2Msg(&pbgame.DestroyDeskReq{
			Head:   &pbcommon.ReqHead{UserID: uId},
			DeskID: s.ID,
			Type:   pbgame.DestroyDeskType_DestroyTypeClub,
		}, reqRCall)
		reqRCall.UserID = uId
		rspRCall := &codec.Message{}
		cli.Call(context.Background(), "DestroyDeskReq", reqRCall, rspRCall)
	}
}
