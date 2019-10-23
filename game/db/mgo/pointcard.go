package mgo

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	UserPointcardTable = "pointcard" //用户点卡数据

)

//游戏创建详情
type UserPointcardData struct {
	UserId     uint64           `bson:"_id"` //用户Id
	Pointcards []*PointcardData //点卡数据
}

//点卡数据
type PointcardData struct {
	PcId           string //点卡Id
	OrderId        string //订单号
	Buytime        int64  //购买事件
	ExchangeNum    uint32 //兑换数
	ExchangeState  uint8  //点卡状态 0未兑换 1已兑换 2过期
	ExchangeTime   int64  //兑换事件
	ExchangeUserId uint64 //兑换用户Id
}

func QueryUserPointcard(uId uint64) (data *UserPointcardData, err error) {
	data = &UserPointcardData{}
	err = mgoSess.DB("").C(UserPointcardTable).Find(bson.M{"_id": uId}).One(data)
	return
}

//添加用户点卡数据
func AddUserPointcard(uId uint64, orderId string, buytime int64, exchangenum uint32) {
	pdata := PointcardData{
		PcId:           fmt.Sprintf("%d%d", uId, buytime),
		OrderId:        orderId,
		Buytime:        buytime,
		ExchangeNum:    exchangenum,
		ExchangeState:  0,
		ExchangeTime:   0,
		ExchangeUserId: 0,
	}

	//更新每局记录id和总分详情
	change := mgo.Change{
		Update:    bson.M{"$push": bson.M{"pointcards": pdata}},
		ReturnNew: false,
		Remove:    false,
		Upsert:    true,
	}
	mgoSess.DB("").C(UserPointcardTable).Find(bson.M{"_id": uId}).Apply(change, nil)
}

//兑换点卡 0 成功 1 不存在  2点卡已失效
func ExchangePointcard(uId uint64, pcId string) (code int32, data *PointcardData) {
	pcdata := &UserPointcardData{}
	if err := mgoSess.DB("").C(UserPointcardTable).Find(bson.M{"pointcards.pcid": pcId}).One(pcdata); err != nil { //不返回游戏回放数据
		return 1, nil
	}

	for _, v := range pcdata.Pointcards {
		if v.PcId == pcId {
			if v.ExchangeState == 1 {
				return 2, nil
			} else {
				//更新记录状态
				change := mgo.Change{
					Update:    bson.M{"$set": bson.M{"pointcards.$.exchangestate": 1, "pointcards.$.exchangetime": time.Now().Unix(), "pointcards.$.exchangeuserid": uId}},
					ReturnNew: false,
					Remove:    false,
					Upsert:    true,
				}
				mgoSess.DB("").C(UserPointcardTable).Find(bson.M{"_id": uId, "pointcards.pcid": pcId}).Apply(change, nil)
				return 0, v
			}
		}
	}
	return 1, nil
}
