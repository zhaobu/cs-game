package notif

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/sirupsen/logrus"
)

const (
	tableNameAddFriendNotif    = "add_friend_notif"
	addFriendNotifKeyPartition = "pid"
	addFriendNotifKeyStoreID   = "id"
	addFriendNotifKeyMsgID     = "msg_id"
)

type AddFriendNotif struct {
	StoreKey   string
	MsgID      int64
	Target     uint64
	Source     uint64
	Msg        string
	InviteTime int64
}

func BatchWriteAddFriendNotif(cli *tablestore.TableStoreClient, reqs []*AddFriendNotif) (err error) {
	if len(reqs) == 0 {
		return
	}

	logrus.WithFields(logrus.Fields{
		"reqs": reqs,
		"err":  err,
	}).Info("db.BatchWriteAddFriendNotif")

	batchWriteReq := &tablestore.BatchWriteRowRequest{}
	for _, req := range reqs {
		putRowChange := new(tablestore.PutRowChange)
		putRowChange.TableName = tableNameAddFriendNotif

		putPk := new(tablestore.PrimaryKey)

		h := md5.New()
		h.Write([]byte(req.StoreKey))
		pkey := h.Sum(nil)[:6]
		h.Reset()

		putPk.AddPrimaryKeyColumn(addFriendNotifKeyPartition, pkey)
		putPk.AddPrimaryKeyColumn(addFriendNotifKeyStoreID, req.StoreKey)
		putPk.AddPrimaryKeyColumn(addFriendNotifKeyMsgID, req.MsgID)

		putRowChange.PrimaryKey = putPk

		// colum value
		putRowChange.AddColumn("target", strconv.FormatUint(req.Target, 10))
		putRowChange.AddColumn("source", strconv.FormatUint(req.Source, 10))
		putRowChange.AddColumn("msg", req.Msg)
		putRowChange.AddColumn("invite_time", req.InviteTime)

		putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
		//putRowChange.SetReturnPk()

		batchWriteReq.AddRowChange(putRowChange)
	}

	response, err := cli.BatchWriteRow(batchWriteReq)
	if err != nil {
		return err
	}
	// todo check all succeed
	for _, rows := range response.TableToRowsResult {
		for _, r := range rows {
			if !r.IsSucceed {
				fmt.Println(r.Error)
			}
		}
	}
	return nil
}

func RangeGetAddFriendNotif(cli *tablestore.TableStoreClient, storeKey string, startMsgid int64, limit int32) (result []*AddFriendNotif, err error) {
	defer func() {
		logrus.WithFields(logrus.Fields{
			"storeKey":   storeKey,
			"startMsgid": startMsgid,
			"limit":      limit,
			"result":     result,
			"err":        err,
		}).Info("db.RangeGetAddFriendNotif")
	}()

	h := md5.New()
	_, err = h.Write([]byte(storeKey))
	if err != nil {
		return nil, err
	}
	pkey := h.Sum(nil)[:6]
	h.Reset()

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn(addFriendNotifKeyPartition, pkey)
	startPK.AddPrimaryKeyColumn(addFriendNotifKeyStoreID, storeKey)
	startPK.AddPrimaryKeyColumn(addFriendNotifKeyMsgID, startMsgid)
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn(addFriendNotifKeyPartition, pkey)
	endPK.AddPrimaryKeyColumn(addFriendNotifKeyStoreID, storeKey)
	endPK.AddPrimaryKeyColumnWithMaxValue(addFriendNotifKeyMsgID)

	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = tableNameAddFriendNotif

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = limit

	getRangeRequest := &tablestore.GetRangeRequest{}
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria
	getRangeResp, err := cli.GetRange(getRangeRequest)

	for {
		if err != nil {
			//fmt.Println("get range failed with error:", err)
		}
		if len(getRangeResp.Rows) > 0 {
			for _, row := range getRangeResp.Rows {
				rr := &AddFriendNotif{}

				for _, pk := range row.PrimaryKey.PrimaryKeys {
					if pk.ColumnName == addFriendNotifKeyStoreID {
						rr.StoreKey = pk.Value.(string)
					} else if pk.ColumnName == addFriendNotifKeyMsgID {
						rr.MsgID = pk.Value.(int64)
					}
				}

				for _, col := range row.Columns {
					if col.ColumnName == "target" {
						rr.Target, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "source" {
						rr.Source, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "msg" {
						rr.Msg = col.Value.(string)
					} else if col.ColumnName == "invite_time" {
						rr.InviteTime = col.Value.(int64)
					}
				}

				result = append(result, rr)
			}
			if getRangeResp.NextStartPrimaryKey == nil {
				break
			} else {
				//fmt.Println("next pk is :", getRangeResp.NextStartPrimaryKey.PrimaryKeys[0].Value, getRangeResp.NextStartPrimaryKey.PrimaryKeys[1].Value, getRangeResp.NextStartPrimaryKey.PrimaryKeys[2].Value)
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = cli.GetRange(getRangeRequest)
			}
		} else {
			break
		}
	}
	return
}

func DeleteAddFriendNotif(cli *tablestore.TableStoreClient, storeKey string, msgid int64) (err error) {

	defer func() {
		logrus.WithFields(logrus.Fields{
			"storeKey": storeKey,
			"msgid":    msgid,
			"err":      err,
		}).Info("db.DeleteAddFriendNotif")
	}()

	deleteRowReq := new(tablestore.DeleteRowRequest)
	deleteRowReq.DeleteRowChange = new(tablestore.DeleteRowChange)
	deleteRowReq.DeleteRowChange.TableName = tableNameAddFriendNotif
	deletePk := new(tablestore.PrimaryKey)

	h := md5.New()
	h.Write([]byte(storeKey))
	pkey := h.Sum(nil)[:6]
	h.Reset()

	deletePk.AddPrimaryKeyColumn(addFriendNotifKeyPartition, pkey)
	deletePk.AddPrimaryKeyColumn(addFriendNotifKeyStoreID, storeKey)
	deletePk.AddPrimaryKeyColumn(addFriendNotifKeyMsgID, msgid)
	deleteRowReq.DeleteRowChange.PrimaryKey = deletePk
	deleteRowReq.DeleteRowChange.SetCondition(tablestore.RowExistenceExpectation_EXPECT_EXIST)
	_, err = cli.DeleteRow(deleteRowReq)

	return err
}
