package main

import (
	"flag"
	"fmt"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

var (
	endpoint        = flag.String("endpoint", `https://zztest-1009.cn-hangzhou.ots.aliyuncs.com`, "endpoint")
	instanceName    = flag.String("instanceName", `zztest-1009`, "instanceName")
	accessKeyID     = flag.String("accessKeyID", `LTAIssLCxHELxHAq`, "accessKeyId")
	accessKeySecret = flag.String("accessKeySecret", `645bzZ5iJxPru921GNrvkYNIm2Uhnf`, "accessKeySecret")

	tsdbCli *tablestore.TableStoreClient
)

func main() {
	InitDB()

	//DeleteTableExclude("notexist")
	DeleteTable("add_friend_notif")
	DeleteTable("add_friend_result")

	//CreateTableChatMsg()
	CreateTableAddFriendNotif()
	CreateTableAddFriendResult()
}

func InitDB() {
	tsdbCli = tablestore.NewClient(*endpoint, *instanceName, *accessKeyID, *accessKeySecret)
}

func DeleteTable(tableName string) error {
	_, err := tsdbCli.DeleteTable(&tablestore.DeleteTableRequest{tableName})
	return err
}

func DeleteTableExclude(tableName string) error {
	listtables, err := tsdbCli.ListTable()
	if err != nil {
		return err
	}
	for _, name := range listtables.TableNames {
		if name != tableName {
			if err := DeleteTable(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateTableChatMsg() {
	const (
		tableNameChatMsg = "chat_msg"

		chatMsgKeyPartition = "pid"    // hash(chatMsgKeyStoreID)后取前6位
		chatMsgKeyStoreID   = "id"     // id:xxxxxx
		chatMsgKeyMsgID     = "msg_id" // 暂时用的UnixNano
		//chatMsgKeySessionID = "session_id" // sid:xxxxxx
	)

	tableMeta := new(tablestore.TableMeta)
	tableMeta.TableName = tableNameChatMsg
	tableMeta.AddPrimaryKeyColumn(chatMsgKeyPartition, tablestore.PrimaryKeyType_BINARY)
	tableMeta.AddPrimaryKeyColumn(chatMsgKeyStoreID, tablestore.PrimaryKeyType_STRING)
	tableMeta.AddPrimaryKeyColumn(chatMsgKeyMsgID, tablestore.PrimaryKeyType_INTEGER)
	//tableMeta.AddPrimaryKeyColumn(chatMsgKeySessionID, tablestore.PrimaryKeyType_STRING)
	//tableMeta.AddPrimaryKeyColumnOption("msg_id", tablestore.PrimaryKeyType_INTEGER, tablestore.AUTO_INCREMENT)

	tableOption := new(tablestore.TableOption)
	tableOption.TimeToAlive = -1
	tableOption.MaxVersion = 1

	reservedThroughput := new(tablestore.ReservedThroughput)
	reservedThroughput.Readcap = 0
	reservedThroughput.Writecap = 0

	createtableRequest := new(tablestore.CreateTableRequest)
	createtableRequest.TableMeta = tableMeta
	createtableRequest.TableOption = tableOption
	createtableRequest.ReservedThroughput = reservedThroughput

	_, err := tsdbCli.CreateTable(createtableRequest)
	if err != nil {
		fmt.Printf("Failed to create table %s with error: %v", tableMeta.TableName, err)
	} else {
		fmt.Printf("Create table %s finished\n", tableMeta.TableName)
	}
}

func CreateTableAddFriendNotif() {
	const (
		tableNameAddFriendNotif    = "add_friend_notif"
		addFriendNotifKeyPartition = "pid"
		addFriendNotifKeyStoreID   = "id"
		addFriendNotifKeyMsgID     = "msg_id"
	)

	tableMeta := new(tablestore.TableMeta)
	tableMeta.TableName = tableNameAddFriendNotif
	tableMeta.AddPrimaryKeyColumn(addFriendNotifKeyPartition, tablestore.PrimaryKeyType_BINARY)
	tableMeta.AddPrimaryKeyColumn(addFriendNotifKeyStoreID, tablestore.PrimaryKeyType_STRING)
	tableMeta.AddPrimaryKeyColumn(addFriendNotifKeyMsgID, tablestore.PrimaryKeyType_INTEGER)

	tableOption := new(tablestore.TableOption)
	tableOption.TimeToAlive = -1
	tableOption.MaxVersion = 1

	reservedThroughput := new(tablestore.ReservedThroughput)
	reservedThroughput.Readcap = 0
	reservedThroughput.Writecap = 0

	createtableRequest := new(tablestore.CreateTableRequest)
	createtableRequest.TableMeta = tableMeta
	createtableRequest.TableOption = tableOption
	createtableRequest.ReservedThroughput = reservedThroughput

	_, err := tsdbCli.CreateTable(createtableRequest)
	if err != nil {
		fmt.Printf("Failed to create table %s with error: %v", tableMeta.TableName, err)
	} else {
		fmt.Printf("Create table %s finished\n", tableMeta.TableName)
	}
}

func CreateTableAddFriendResult() {
	const (
		tableNameAddFriendResult    = "add_friend_result"
		addFriendResultKeyPartition = "pid"
		addFriendResultKeyStoreID   = "id"
		addFriendResultKeyMsgID     = "msg_id"
	)

	tableMeta := new(tablestore.TableMeta)
	tableMeta.TableName = tableNameAddFriendResult
	tableMeta.AddPrimaryKeyColumn(addFriendResultKeyPartition, tablestore.PrimaryKeyType_BINARY)
	tableMeta.AddPrimaryKeyColumn(addFriendResultKeyStoreID, tablestore.PrimaryKeyType_STRING)
	tableMeta.AddPrimaryKeyColumn(addFriendResultKeyMsgID, tablestore.PrimaryKeyType_INTEGER)

	tableOption := new(tablestore.TableOption)
	tableOption.TimeToAlive = -1
	tableOption.MaxVersion = 1

	reservedThroughput := new(tablestore.ReservedThroughput)
	reservedThroughput.Readcap = 0
	reservedThroughput.Writecap = 0

	createtableRequest := new(tablestore.CreateTableRequest)
	createtableRequest.TableMeta = tableMeta
	createtableRequest.TableOption = tableOption
	createtableRequest.ReservedThroughput = reservedThroughput

	_, err := tsdbCli.CreateTable(createtableRequest)
	if err != nil {
		fmt.Printf("Failed to create table %s with error: %v", tableMeta.TableName, err)
	} else {
		fmt.Printf("Create table %s finished\n", tableMeta.TableName)
	}
}
