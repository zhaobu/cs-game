package mgo

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type ClubEmail struct {
	ID       int64 `bson:"_id,omitempty"`
	SendTime int64
	// 1JoinClub  2InviteJoinClub 3TransferMaster
	Typ     int32 // 0 userid为邮件接收者
	Content string
	// 0未处理 1已处理
	Flag    int32
	ClubID  int64
	UserID1 uint64 // 1 申请人 2 邀请人 3 老的群主
	UserID2 uint64 // 1 ----  2 被邀请人 3 新的群主
}

type userEmail struct {
	UserID uint64
	Email  *mgo.DBRef
}

func QueryEmail(id int64) (e *ClubEmail, err error) {
	e = &ClubEmail{}
	err = mgoSess.DB("").C("clubemail").FindId(id).One(e)
	return
}

func QueryClubEmail(cid int64) (e []*ClubEmail, err error) {
	e = make([]*ClubEmail, 0)
	err = mgoSess.DB("").C("clubemail").Find(bson.M{"clubid": cid}).All(&e)
	return
}

func ExistEmail(f bson.M) (bool, error) {
	n, err := mgoSess.DB("").C("clubemail").Find(f).Count()
	if err != nil {
		return false, err
	}
	fmt.Println(n, err)
	return n > 0, nil
}

func SetEmailFlag(id int64, flag int32) error {
	return mgoSess.DB("").C("clubemail").UpdateId(id, bson.M{"$set": bson.M{"flag": flag}})
}

func BatchSetEmailFlag(ids ...int64) error {
	_, err := mgoSess.DB("").C("clubemail").UpdateAll(bson.M{"_id": bson.M{"$in": ids}}, bson.M{"$set": bson.M{"flag": 1}})
	return err
}

func QueryUserEmail(uid uint64) (rsp []*ClubEmail, err error) {
	rsp = make([]*ClubEmail, 0)
	find := make([]*userEmail, 0)
	err = mgoSess.DB("").C("useremail").Find(bson.M{"userid": uid}).All(&find)
	if err != nil {
		return
	}

	for _, f := range find {
		ce := &ClubEmail{}
		if err := mgoSess.DB("").FindRef(f.Email).One(ce); err == nil {
			rsp = append(rsp, ce)
		}
	}
	return
}

func AddClubEmail(e *ClubEmail, uids ...uint64) (err error) {
	e.ID, err = incEmailID()
	if err != nil {
		return
	}

	err = mgoSess.DB("").C("clubemail").Insert(e)
	for _, uid := range uids {
		ue := &userEmail{
			UserID: uid,
			Email: &mgo.DBRef{
				Collection: "clubemail",
				Id:         e.ID,
			},
		}
		mgoSess.DB("").C("useremail").Insert(ue)
	}
	return
}

func incEmailID() (int64, error) {
	result := bson.M{}
	_, err := mgoSess.DB("").C("emailid").Find(nil).Apply(mgo.Change{
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
