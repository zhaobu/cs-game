package mgo

import "time"

const (
	UserWealthFlow = "userwealthflow" //财富流水记录表
)

type UserWealthFlowData struct {
	Id uint64 `bson:"_id"`
	Flows []*WealthFlowRecordData
}

//本局数据
type WealthFlowRecordData struct {
	RecordTime 	 int64             //开始时间 存储时间错
	SourceType   uint32
	SourceData   string
	CurrencyType uint32
	Change       int64
}

//写入流水数据
func WriteWealthRecordData(uId uint64,stype uint32,sdata string,ctype uint32 ,change int64) (err error) {
	UserRecor := &UserWealthFlowData{}
	if err = mgoSess.DB("").C(UserWealthFlow).FindId(uId).One(UserRecor); err != nil {
		UserRecor = &UserWealthFlowData{
			Id:uId,
			Flows:[]*WealthFlowRecordData{},
		}
	}
	UserRecor.Flows = append(UserRecor.Flows,&WealthFlowRecordData{
			RecordTime:time.Now().Unix(),
			SourceType:stype,
			SourceData:sdata,
			CurrencyType:ctype,
			Change:change,
		})
	_, err = mgoSess.DB("").C(UserWealthFlow).UpsertId(uId, UserRecor)
	if err != nil {
		return err
	}
	return
}