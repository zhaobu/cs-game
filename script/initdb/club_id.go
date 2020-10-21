package main

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	clubid = make(map[int64]struct{})
)

func genClubID() {
	// 俱乐部ID从 1100000 到 1199999
	var i int64
	for ; i < 99999; i++ {
		clubid[i+1100000] = struct{}{}
	}
	mgoSess.DB("").C("club_id").RemoveAll(nil)

	begin := time.Now()
	bulk := mgoSess.DB("").C("club_id").Bulk()
	var cnt int64
	for k := range clubid {
		bulk.Insert(bson.M{"_id": k, "inuse": false})
		cnt++
		if cnt == 300 {
			bulk.Run()
			cnt = 0
		}
	}
	bulk.Run()

	fmt.Println("genClubID cost:", time.Now().Sub(begin))
}
