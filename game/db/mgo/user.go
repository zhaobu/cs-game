package mgo

import (
	"cy/game/pb/common"
	"cy/game/util"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	mgoSess *mgo.Session
)

func Init(url string) (err error) {
	mgoSess, err = mgo.Dial(url)
	return
}

func UpdateWealth(uid uint64, feeType uint32, change int64) (uint64, error) {
	field := "gold"
	if feeType == 2 {
		field = "masonry"
	}

	result := bson.M{}
	_, err := mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{field: change}},
	}, result)
	if err != nil {
		return 0, err
	}
	r, ok := result[field].(int64)
	if !ok {
		return 0, fmt.Errorf("not int64")
	}
	if r < 0 {
		mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
			Upsert:    false,
			ReturnNew: true,
			Update:    bson.M{"$set": bson.M{field: int64(0)}},
		}, result)
		return 0, nil
	}
	return uint64(r), nil
}

func UpdateWealthPre(uid uint64, feeType uint32, change int64) (uint64, error) {
	field := "gold"
	if feeType == 2 {
		field = "masonry"
	}
	fieldPre := field + "pre"

	result := bson.M{}
	_, err := mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid, field: bson.M{"$gte": change}}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{field: change * -1, fieldPre: change}},
	}, result)
	if err != nil {
		return 0, err
	}

	r, ok := result[field].(int64)
	if !ok {
		return 0, fmt.Errorf("not int64")
	}

	return uint64(r), nil
}

// UpdateWealthPre 反操作
func UpdateWealthPreSure(uid uint64, feeType uint32, change int64) (uint64, error) {
	field := "gold"
	if feeType == 2 {
		field = "masonry"
	}
	fieldPre := field + "pre"

	result := bson.M{}
	_, err := mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid, fieldPre: bson.M{"$gte": change}}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{field: change, fieldPre: change * -1}},
	}, result)
	if err != nil {
		return 0, err
	}
	r, ok := result[field].(int64)
	if !ok {
		return 0, fmt.Errorf("not int64")
	}

	return uint64(r), nil
}

func QueryUserInfo(uid uint64) (info *pbcommon.UserInfo, err error) {
	info = &pbcommon.UserInfo{}
	result := bson.M{}
	err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).One(result)
	if err != nil {
		return nil, err
	}
	err = util.Bson2struct(result, info)
	return
}
