package mgo

import (
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"cy/game/util"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
)

var (
	mgoSess *mgo.Session
)

func Init(url string) (err error) {
	mgoSess, err = mgo.Dial(url)
	return
}

// UpdateWealth 更新财富 feeType 1:gold, 2:masonry  change有符号
func UpdateWealth(uid uint64, feeType uint32, change int64) (*pbcommon.UserInfo, error) {
	field := ""
	if feeType == 1 {
		field = "gold"
	} else if feeType == 2 {
		field = "masonry"
	} else {
		return nil, fmt.Errorf("bad feeType %d", feeType)
	}

	result := bson.M{}
	_, err := mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{field: change}},
	}, result)

	if err != nil {
		return nil, err
	}

	r, ok := result[field].(int64)
	if !ok {
		return nil, fmt.Errorf("%s not int64", field)
	}

	// 扣为负数则设置为0
	if r < 0 {
		_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
			Upsert:    false,
			ReturnNew: true,
			Update:    bson.M{"$set": bson.M{field: int64(0)}},
		}, result)
	}

	rsp := &pbcommon.UserInfo{}
	err = util.Bson2struct(result, rsp)
	return rsp, err
}

// UpdateWealthPre 预扣财富 feeType 1:gold, 2:masonry change需为正数
func UpdateWealthPre(uid uint64, feeType pbgame.FeeType, change int64) (*pbcommon.UserInfo, error) {
	if change <= 0 {
		return nil, fmt.Errorf("bad change %d", change)
	}

	field := "gold"
	if feeType == pbgame.FeeType_FTMasonry {
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
		return nil, err
	}

	rsp := &pbcommon.UserInfo{}
	err = util.Bson2struct(result, rsp)
	return rsp, err
}

// UpdateWealthPreSure UpdateWealthPre的反向操作
func UpdateWealthPreSure(uid uint64, feeType pbgame.FeeType, change int64) (*pbcommon.UserInfo, error) {
	if change <= 0 {
		return nil, fmt.Errorf("bad change %d", change)
	}

	field := "gold"
	if feeType == pbgame.FeeType_FTMasonry {
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
		return nil, err
	}

	rsp := &pbcommon.UserInfo{}
	err = util.Bson2struct(result, rsp)
	return rsp, err
}

// UpsertUserInfo 插入更新玩家信息
func UpsertUserInfo(u *pbcommon.UserInfo) (*pbcommon.UserInfo, bool, error) {
	coll := mgoSess.DB("").C("userinfo")

	var find = make(bson.M)
	err := coll.Find(bson.M{"wxid": u.WxID}).One(find)
	if err != nil {
		if err == mgo.ErrNotFound {
			var err2 error
			u.UserID, err2 = incUserID()
			if err2 != nil {
				return nil, false, err2
			}

			// 新玩家初始财富，必须要赋值，不能用客户端传过来的
			u.Gold = 5000
			u.Masonry = 8
			u.GoldPre = 0
			u.MasonryPre = 0
			u.RegisterTime = time.Now().Unix()	//注册时间
			bs, _ := util.Struct2bson(u)
			return u, true, coll.Insert(bs)
		}
		return nil, false, err
	}

	old := &pbcommon.UserInfo{}
	err = util.Bson2struct(find, old)
	if err != nil {
		return nil, false, err
	}

	// 更新的信息
	old.Longitude = u.Longitude
	old.Latitude = u.Latitude
	old.Name = u.Name
	old.Sex = u.Sex
	old.Profile = u.Profile

	bs, _ := util.Struct2bson(old)
	return old, false, coll.Update(bson.M{"wxid": old.WxID}, bs)
}

func incUserID() (uint64, error) {
	result := bson.M{}
	_, err := mgoSess.DB("").C("userid").Find(nil).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{"max": int64(1)}},
	}, result)
	if err != nil {
		return 0, err
	}

	r, _ := result["max"].(int64)
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

func QueryUserByMobile(mobile string) (info *pbcommon.UserInfo, err error) {
	info = &pbcommon.UserInfo{}
	result := bson.M{}
	err = mgoSess.DB("").C("userinfo").Find(bson.M{"mobile": mobile}).One(result)
	if err != nil {
		return nil, err
	}
	err = util.Bson2struct(result, info)
	return
}
func QueryUserByXianLiao(xlId string) (info *pbcommon.UserInfo, err error) {
	info = &pbcommon.UserInfo{}
	result := bson.M{}
	err = mgoSess.DB("").C("userinfo").Find(bson.M{"xlid": xlId}).One(result)
	if err != nil {
		return nil, err
	}
	err = util.Bson2struct(result, info)
	return
}




func updateUserOneField(uid uint64, fieldName string, newValue string) (info *pbcommon.UserInfo, err error) {
	result := bson.M{}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$set": bson.M{fieldName: newValue}},
	}, result)

	if err != nil {
		return nil, err
	}

	info = &pbcommon.UserInfo{}
	err = util.Bson2struct(result, info)
	return
}

func updateUserManyField(uid uint64, kv map[string]string) (info *pbcommon.UserInfo, err error) {
	bm := bson.M{}
	for k, v := range kv {
		bm[k] = v
	}

	result := bson.M{}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$set": bm},
	}, result)

	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", result)

	info = &pbcommon.UserInfo{}
	err = util.Bson2struct(result, info)
	return
}

func UpdateIDCardAndName(uid uint64, idCard, cnName string) (info *pbcommon.UserInfo, err error) {
	return updateUserManyField(uid, map[string]string{"identitycard": idCard, "identitycardname": cnName})
}

func UpdateBindMobile(uid uint64, mobile, password string) (info *pbcommon.UserInfo, err error) {
	return updateUserManyField(uid, map[string]string{"mobile": mobile, "password": password})
}

func BindXianLiaoID(uid uint64, xlid string) (info *pbcommon.UserInfo, err error) {
	return updateUserManyField(uid, map[string]string{"xlid": xlid})
}

func UpdateAgentID(uid uint64, agentID string) (info *pbcommon.UserInfo, err error) {
	return updateUserOneField(uid, "agent", agentID)
}

func UpdateUserLocation(uid uint64,Longitude, Latitude float64,Place string)(err error){
	result := bson.M{}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$set": bson.M{"longitude": Longitude,"latitude":Latitude,"place":Place}},
	}, result)
	return err
}

func UpdateUserIP(uid uint64,Ip string)error{
	_,err := updateUserOneField(uid, "ip", Ip)
	return err
}

func UpdateSessionID(uid uint64, sessionID string) (info *pbcommon.UserInfo, err error) {
	return updateUserOneField(uid, "sessionid", sessionID)
}

//添加用户财富 对net端口开放
func UpdateWealthByNet(uid uint64, feeType uint32, change int64) (user *pbcommon.UserInfo, err error) {
	field := ""
	if feeType == 1 {
		field = "gold"
	} else if feeType == 2 {
		field = "masonry"
	} else {
		return nil, fmt.Errorf("修改的货币类型错误 %d", feeType)
	}
	user, err = QueryUserInfo(uid)
	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败 err=%s", err.Error())
	}
	oldnum := int64(0)
	newnum := int64(0)
	if feeType == 1 {
		oldnum = int64(user.Gold)
	} else {
		oldnum = int64(user.Masonry)
	}
	newnum = oldnum + change
	if newnum < 0 {
		return nil, fmt.Errorf("余额不足")
	}
	result := bson.M{}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{field: change}},
	}, result)
	user = &pbcommon.UserInfo{}
	err = util.Bson2struct(result, user)
	return user, nil
}

//设置用户红名接口
func SetUserRedName(uid uint64, isredname bool) (user *pbcommon.UserInfo, err error) {
	result := bson.M{}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$set": bson.M{"isredname": isredname}},
	}, result)
	if err != nil {
		return nil, err
	}
	user = &pbcommon.UserInfo{}
	err = util.Bson2struct(result, user)
	return user, nil
}

//设置用户邮寄记录信息
func SetUserGamePlay(uid uint64, iswin bool) (err error) {
	result := bson.M{}
	var change mgo.Change
	if iswin {
		change = mgo.Change{
			Upsert:    true,
			ReturnNew: true,
			Update:    bson.M{"$inc": bson.M{"totalcase": 1, "wincasecase": 1}},
		}
	} else {
		change = mgo.Change{
			Upsert:    true,
			ReturnNew: true,
			Update:    bson.M{"$inc": bson.M{"totalcase": 1}},
		}
	}
	_, err = mgoSess.DB("").C("userinfo").Find(bson.M{"userid": uid}).Apply(change, result)
	return err
}
