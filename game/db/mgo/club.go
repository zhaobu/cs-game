package mgo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type ClubMember struct {
	UserID   uint64
	Identity int32
	Agree    bool
	Relation []uint64	//改成员关系列表
}

type DeskSetting struct {
	GameName        string
	GameArgMsgName  string
	GameArgMsgValue []byte
	Enable          bool
}

type Club struct {
	ID              int64
	MasterUserID    uint64
	Profile         string
	Name            string
	IsAutoCreate    bool
	IsCustomGameArg bool
	IsMasterPay     bool
	Notice          string
	CurrDayDestoryDeskNum int		//当天解散房间次数
	LastDestoryDeskNumTime int64	//最后解散桌子时间
	Members         map[uint64]*ClubMember
	GameArgs        []*DeskSetting
}

type clubId struct {
	ID    int64 `bson:"_id"`
	Inuse bool
}

// var (
// 	muClubID sync.RWMutex
// 	clubIds  = make(map[int64]struct{})
// )

// func initClubId() {
// 	//tmp := make([]*clubId, 0)
// 	var tmp []*clubId
// 	mgoSess.DB("").C("club_id").Find(bson.M{"inuse": false}).All(&tmp)
// 	for _, v := range tmp {
// 		clubIds[v.ID] = struct{}{}
// 	}
// }

// func AllocClubID() (id int64, err error) {
// 	muClubID.Lock()
// 	defer muClubID.Unlock()

// 	if len(clubIds) == 0 {
// 		return 0, fmt.Errorf("clubid exhaust")
// 	}

// 	for k := range clubIds {
// 		id = k
// 		break
// 	}

// 	err = mgoSess.DB("").C("club_id").UpdateId(id, bson.M{"$set": bson.M{"inuse": true}})
// 	if err != nil {
// 		return
// 	}
// 	delete(clubIds, id)

// 	return
// }

func AllocClubID() (id int64, err error) {
	var tmp clubId
	_, err = mgoSess.DB("").C("club_id").Find(bson.M{"inuse": false}).Apply(mgo.Change{
		Update:    bson.M{"$set": bson.M{"inuse": true}},
		Upsert:    false,
		Remove:    false,
		ReturnNew: true,
	}, &tmp)
	return tmp.ID, err
}

func IncClubID() (int64, error) {
	result := bson.M{}
	_, err := mgoSess.DB("").C("clubid").Find(nil).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{"max": int64(1)}},
	}, result)
	if err != nil {
		return 0, err
	}

	r, _ := result["max"].(int64)
	return r, nil
}

func RemoveClub(id int64) error {
	return mgoSess.DB("").C("clubinfo").Remove(bson.M{"id": id})
}

func SaveClub(req *Club) (err error) {
	_, err = mgoSess.DB("").C("clubinfo").Upsert(bson.M{"id": req.ID}, req)
	return
}

func QueryAllClub() (rsp []*Club, err error) {
	rsp = make([]*Club, 0)
	err = mgoSess.DB("").C("clubinfo").Find(nil).Sort("id").All(&rsp)
	return
}

func QueryClubByID(id int64) (c *Club, err error) {
	c = &Club{}
	err = mgoSess.DB("").C("clubinfo").Find(bson.M{"id": id}).One(c)
	return
}

//校验创建房间接口 code 0 校验成功 1俱乐部不存在 2用户不是本俱乐部的 3用户无权创建房间
func CheckCreateClubRoom(cId int64,uId uint64)(code int32){
	c := &Club{}
	err := mgoSess.DB("").C("clubinfo").Find(bson.M{"id": cId}).One(c)
	if err != nil{
		return 1
	}else{
		for _,u:=range c.Members{
			if u.UserID == uId{
				if u.Identity != 4 {	//不是黑名单用户
					return 0
				}else{
					return 3
				}
			}
		}
		return 2
	}
}

//校验加入房间接口 code 0 校验成功 1俱乐部不存在 2用户不是本俱乐部的 3用户无权加入房间 4房间中有亲属成员存在
func CheckJoinClubRoom(uId uint64, cId int64,roomusers []uint64)(code int32){
	c := &Club{}
	err := mgoSess.DB("").C("clubinfo").Find(bson.M{"id": cId}).One(c)
	if err != nil{
		return 1
	}else{
		for _,u:=range c.Members{
			if u.UserID == uId{
				if u.Identity != 4 {	//不是黑名单用户
					if roomusers != nil{
						for _,v := range u.Relation{
							for _,v1 := range roomusers {
								if v == v1 {
									return 4
								}
							}
						}
					}
					return 0
				}else{
					return 3
				}
			}
		}
		return 2
	}
}
