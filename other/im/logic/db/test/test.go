package main

import (
	"cy/other/im/inner"
	
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func testWriteChatMsg() {
	var reqs []*db.ChatMsg

	nt := time.Now().UTC().UnixNano()
	var to uint64 = 10002
	var from uint64 = 10003
	content := []byte(`world`)

	reqs = append(reqs, &db.ChatMsg{
		StoreKey: inner.StoreKey(from),
		MsgID:    nt,

		SessionKey: inner.SessionID(from, to, false, false),
		To:         to,
		From:       from,
		GroupID:    0,
		Content:    content,
		Ct:         0,
		SentTime:   nt})

	nt++

	reqs = append(reqs, &db.ChatMsg{
		StoreKey: inner.StoreKey(to),
		MsgID:    nt,

		SessionKey: inner.SessionID(from, to, false, false),
		To:         to,
		From:       from,
		GroupID:    0,
		Content:    content,
		Ct:         0,
		SentTime:   nt})

	if err := db.BatchWriteChatMsg(reqs); err != nil {
		fmt.Println(err)
	}
}

func testReadChatMsg2() {

	storeKey := inner.StoreKey(10002)
	sessionKey := inner.SessionID(10002, 10003, false, false)

	r, err := db.RangeGetBySessionKey(storeKey, sessionKey, 0, 3)
	if err != nil {
		fmt.Println(err)
		return
	}
	for idx, v := range r {
		fmt.Printf("%d %+v\n", idx, v)
	}
	return
}

func init() {
	Log.SetFormatter(&Log.JSONFormatter{})
	logName := fmt.Sprintf("logic_%d_%d.Log", os.Getpid(), time.Now().Unix())
	file, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		Log.SetOutput(file)
	} else {
		Log.SetOutput(os.Stdout)
	}
}

func main() {

	//db.InitTS()

	return
}
