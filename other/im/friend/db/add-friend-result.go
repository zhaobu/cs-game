package db
import (
	"crypto/md5"
	. "cy/other/im/common/logger"
	"fmt"
	"strconv"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

const (
	tableNameAddFriendResult    = "add_friend_result"
	addFriendResultKeyPartition = "pid"
	addFriendResultKeyStoreID   = "id"
	addFriendResultKeyMsgID     = "msg_id"
)

type AddFriendResult struct {
	StoreKey   string
	MsgID      int64
	Target     uint64
	Source     uint64
	Msg        string
	InviteTime int64
	Code       int64
}

func BatchWriteAddFriendResult(cli *tablestore.TableStoreClient, reqs []*AddFriendResult) (err error) {
	if len(reqs) == 0 {
		return
	}

	defer func() {
		Log.Info("db.BatchWriteAddFriendResult,reqs=%v,err=%s", reqs, err)
	}()

	batchWriteReq := &tablestore.BatchWriteRowRequest{}
	for _, req := range reqs {
		putRowChange := new(tablestore.PutRowChange)
		putRowChange.TableName = tableNameAddFriendResult

		putPk := new(tablestore.PrimaryKey)

		h := md5.New()
		h.Write([]byte(req.StoreKey))
		pkey := h.Sum(nil)[:6]
		h.Reset()

		putPk.AddPrimaryKeyColumn(addFriendResultKeyPartition, pkey)
		putPk.AddPrimaryKeyColumn(addFriendResultKeyStoreID, req.StoreKey)
		putPk.AddPrimaryKeyColumn(addFriendResultKeyMsgID, req.MsgID)

		putRowChange.PrimaryKey = putPk

		// colum value
		putRowChange.AddColumn("target", strconv.FormatUint(req.Target, 10))
		putRowChange.AddColumn("source", strconv.FormatUint(req.Source, 10))
		putRowChange.AddColumn("msg", req.Msg)
		putRowChange.AddColumn("invite_time", req.InviteTime)
		putRowChange.AddColumn("code", req.Code)

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

func RangeGetAddFriendResult(cli *tablestore.TableStoreClient, storeKey string, startMsgid int64, limit int32) (result []*AddFriendResult, err error) {
	defer func() {
		Log.Infof("db.RangeGetAddFriendResult:storeKey=%s,startMsgid=%d,limit=%d,result=%v,err=%s", storeKey, startMsgid, limit, limit, err)
	}()

	h := md5.New()
	_, err = h.Write([]byte(storeKey))
	if err != nil {
		return nil, err
	}
	pkey := h.Sum(nil)[:6]
	h.Reset()

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn(addFriendResultKeyPartition, pkey)
	startPK.AddPrimaryKeyColumn(addFriendResultKeyStoreID, storeKey)
	startPK.AddPrimaryKeyColumn(addFriendResultKeyMsgID, startMsgid)
	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn(addFriendResultKeyPartition, pkey)
	endPK.AddPrimaryKeyColumn(addFriendResultKeyStoreID, storeKey)
	endPK.AddPrimaryKeyColumnWithMaxValue(addFriendResultKeyMsgID)

	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = tableNameAddFriendResult

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
				rr := &AddFriendResult{}

				for _, pk := range row.PrimaryKey.PrimaryKeys {
					if pk.ColumnName == addFriendResultKeyStoreID {
						rr.StoreKey = pk.Value.(string)
					} else if pk.ColumnName == addFriendResultKeyMsgID {
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
					} else if col.ColumnName == "code" {
						rr.Code = col.Value.(int64)
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

func DeleteAddFriendResult(cli *tablestore.TableStoreClient, storeKey string, msgid int64) (err error) {
	defer func() {
		Log.Infof("db.DeleteAddFriendResult:storeKey=%s,msgid=%d,err=%s", storeKey, msgid, err)
	}()

	deleteRowReq := new(tablestore.DeleteRowRequest)
	deleteRowReq.DeleteRowChange = new(tablestore.DeleteRowChange)
	deleteRowReq.DeleteRowChange.TableName = tableNameAddFriendResult
	deletePk := new(tablestore.PrimaryKey)

	h := md5.New()
	h.Write([]byte(storeKey))
	pkey := h.Sum(nil)[:6]
	h.Reset()

	deletePk.AddPrimaryKeyColumn(addFriendResultKeyPartition, pkey)
	deletePk.AddPrimaryKeyColumn(addFriendResultKeyStoreID, storeKey)
	deletePk.AddPrimaryKeyColumn(addFriendResultKeyMsgID, msgid)
	deleteRowReq.DeleteRowChange.PrimaryKey = deletePk
	deleteRowReq.DeleteRowChange.SetCondition(tablestore.RowExistenceExpectation_EXPECT_EXIST)
	_, err = cli.DeleteRow(deleteRowReq)

	return err
}

