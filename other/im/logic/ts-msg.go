package main

import (
	"crypto/md5"
	. "cy/other/im/common/logger"
	"cy/other/im/inner"
	"flag"
	"strconv"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

const (
	tableNameChatMsg = "chat_msg"

	chatMsgKeyPartition = "pid"    // hash(chatMsgKeyID)后取前6位
	chatMsgKeyStoreID   = "id"     // id:xxxxxx
	chatMsgKeyMsgID     = "msg_id" // 暂时用的UnixNano
)

var (
	endpoint        = flag.String("endpoint", `https://zztest-1009.cn-hangzhou.ots.aliyuncs.com`, "endpoint")
	instanceName    = flag.String("instanceName", `zztest-1009`, "instanceName")
	accessKeyID     = flag.String("accessKeyID", `LTAIssLCxHELxHAq`, "accessKeyId")
	accessKeySecret = flag.String("accessKeySecret", `645bzZ5iJxPru921GNrvkYNIm2Uhnf`, "accessKeySecret")

	tsdbCli *tablestore.TableStoreClient
)

func InitTS() {
	tsdbCli = tablestore.NewClient(*endpoint, *instanceName, *accessKeyID, *accessKeySecret)
}

type ChatMsg struct {
	StoreKey string
	MsgID    int64

	SessionKey string
	To         uint64
	From       uint64
	GroupID    uint64 // 房间ID 或者 世界ID 暂时没用到（TO够用了）
	Content    []byte
	Ct         int64

	SentTime int64
}

func BatchWriteChatMsg(reqs []*ChatMsg) (err error) {
	if len(reqs) == 0 {
		return
	}

	Log.Infof("db.BatchWriteChatMsg,reqs=%v", reqs)

	batchWriteReq := &tablestore.BatchWriteRowRequest{}
	for _, req := range reqs {
		putRowChange := new(tablestore.PutRowChange)
		putRowChange.TableName = tableNameChatMsg

		putPk := new(tablestore.PrimaryKey)

		h := md5.New()
		h.Write([]byte(req.StoreKey))
		pkey := h.Sum(nil)[:6]
		h.Reset()

		putPk.AddPrimaryKeyColumn(chatMsgKeyPartition, pkey)
		putPk.AddPrimaryKeyColumn(chatMsgKeyStoreID, req.StoreKey)
		putPk.AddPrimaryKeyColumn(chatMsgKeyMsgID, req.MsgID)

		putRowChange.PrimaryKey = putPk

		// colum value
		putRowChange.AddColumn("session_key", req.SessionKey)
		putRowChange.AddColumn("to_id", strconv.FormatUint(req.To, 10))
		putRowChange.AddColumn("from_id", strconv.FormatUint(req.From, 10))
		putRowChange.AddColumn("group_id", strconv.FormatUint(req.GroupID, 10))
		putRowChange.AddColumn("content", req.Content)
		putRowChange.AddColumn("ct", req.Ct)
		putRowChange.AddColumn("send_time", req.SentTime)

		putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
		putRowChange.SetReturnPk()

		batchWriteReq.AddRowChange(putRowChange)
	}

	response, err := tsdbCli.BatchWriteRow(batchWriteReq)
	if err != nil {
		return err
	}
	// todo check all succeed
	for _, rows := range response.TableToRowsResult {
		for _, r := range rows {
			if !r.IsSucceed {
				// TODO
			}
		}
	}
	return nil
}

func RangeGetMsgRecord(storeKey string, startMsgid, endMsgid int64, limit int32) (result []*ChatMsg, err error) {
	defer func() {
		Log.Info("db.RangeGetMsgRecord,storeKey=%s,startMsgid=%d,endMsgid=%d,limit=%d,result=%v,err=%s", storeKey, startMsgid, endMsgid, limit, result, err)
	}()

	h := md5.New()
	_, err = h.Write([]byte(storeKey))
	if err != nil {
		return nil, err
	}
	pkey := h.Sum(nil)[:6]
	h.Reset()

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyPartition, pkey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyStoreID, storeKey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyMsgID, startMsgid)

	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn(chatMsgKeyPartition, pkey)
	endPK.AddPrimaryKeyColumn(chatMsgKeyStoreID, storeKey)
	if endMsgid == 0 || endMsgid < startMsgid {
		endPK.AddPrimaryKeyColumnWithMaxValue(chatMsgKeyMsgID)
	} else {
		endPK.AddPrimaryKeyColumn(chatMsgKeyMsgID, endMsgid)
	}

	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = tableNameChatMsg

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = limit

	getRangeRequest := &tablestore.GetRangeRequest{}
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	// condition := tablestore.NewSingleColumnCondition("from_id", tablestore.CT_EQUAL, "10003")
	// getRangeRequest.RangeRowQueryCriteria.Filter = condition

	getRangeResp, err := tsdbCli.GetRange(getRangeRequest)
	for {
		if err != nil {

		}
		if len(getRangeResp.Rows) > 0 {
			for _, row := range getRangeResp.Rows {
				rr := &ChatMsg{}

				for _, pk := range row.PrimaryKey.PrimaryKeys {
					if pk.ColumnName == chatMsgKeyStoreID {
						rr.StoreKey = pk.Value.(string)
					} else if pk.ColumnName == chatMsgKeyMsgID {
						rr.MsgID = pk.Value.(int64)
					}
				}

				for _, col := range row.Columns {
					if col.ColumnName == "session_key" {
						rr.SessionKey = col.Value.(string)
					} else if col.ColumnName == "to_id" {
						rr.To, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "from_id" {
						rr.From, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "group_id" {
						rr.GroupID, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "content" {
						rr.Content = col.Value.([]byte)
					} else if col.ColumnName == "ct" {
						rr.Ct = col.Value.(int64)
					} else if col.ColumnName == "send_time" {
						rr.SentTime = col.Value.(int64)
					}
				}

				if int32(len(result)) >= limit {
					return
				}
				result = append(result, rr)

			}
			if getRangeResp.NextStartPrimaryKey == nil {
				break
			} else {
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = tsdbCli.GetRange(getRangeRequest)
			}
		} else {
			break
		}
	}
	return
}

func RangeGetBySessionKey(storeKey, sessionKey string, startMsgid int64, limit int32) (result []*ChatMsg, err error) {
	defer func() {
		Log.Info("db.RangeGetBySessionKey,storeKey=%s,sessionKey=%s,startMsgid=%d,limit=%d,result=%v,err=%s", storeKey, sessionKey, startMsgid, limit, result, err)
	}()

	h := md5.New()
	_, err = h.Write([]byte(storeKey))
	if err != nil {
		return nil, err
	}
	pkey := h.Sum(nil)[:6]
	h.Reset()

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyPartition, pkey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyStoreID, storeKey)
	startPK.AddPrimaryKeyColumn(chatMsgKeyMsgID, startMsgid)

	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn(chatMsgKeyPartition, pkey)
	endPK.AddPrimaryKeyColumn(chatMsgKeyStoreID, storeKey)
	endPK.AddPrimaryKeyColumnWithMaxValue(chatMsgKeyMsgID)

	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = tableNameChatMsg

	rangeRowQueryCriteria.StartPrimaryKey = endPK
	rangeRowQueryCriteria.EndPrimaryKey = startPK
	rangeRowQueryCriteria.Direction = tablestore.BACKWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = limit

	getRangeRequest := &tablestore.GetRangeRequest{}
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	c1 := tablestore.NewSingleColumnCondition("session_key", tablestore.CT_EQUAL, sessionKey)
	uid := inner.IdFromStoreKey(storeKey)
	c2 := tablestore.NewSingleColumnCondition("from_id", tablestore.CT_NOT_EQUAL, uid)
	cf := tablestore.NewCompositeColumnCondition(tablestore.LO_AND)
	cf.AddFilter(c1)
	cf.AddFilter(c2)

	getRangeRequest.RangeRowQueryCriteria.Filter = cf

	getRangeResp, err := tsdbCli.GetRange(getRangeRequest)

	for {
		if err != nil {

		}
		if len(getRangeResp.Rows) > 0 {
			for _, row := range getRangeResp.Rows {
				rr := &ChatMsg{}

				for _, pk := range row.PrimaryKey.PrimaryKeys {
					if pk.ColumnName == chatMsgKeyStoreID {
						rr.StoreKey = pk.Value.(string)
					} else if pk.ColumnName == chatMsgKeyMsgID {
						rr.MsgID = pk.Value.(int64)
					}
				}

				for _, col := range row.Columns {
					if col.ColumnName == "session_key" {
						rr.SessionKey = col.Value.(string)
					} else if col.ColumnName == "to_id" {
						rr.To, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "from_id" {
						rr.From, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "group_id" {
						rr.GroupID, _ = strconv.ParseUint(col.Value.(string), 10, 64)
					} else if col.ColumnName == "content" {
						rr.Content = col.Value.([]byte)
					} else if col.ColumnName == "ct" {
						rr.Ct = col.Value.(int64)
					} else if col.ColumnName == "send_time" {
						rr.SentTime = col.Value.(int64)
					}
				}

				if int32(len(result)) >= limit {
					return
				}
				result = append(result, rr)
			}
			if getRangeResp.NextStartPrimaryKey == nil {
				break
			} else {

				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = tsdbCli.GetRange(getRangeRequest)
			}
		} else {
			break
		}
	}
	return
}
