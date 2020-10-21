package main

import (
	"fmt"
	"math/rand"
	"time"
)

func genDeskID(startDeskID, endDeskID int) {

	redisCli.FlushDB()
	var deskIDs []interface{}
	//deskIDs = append(deskIDs, "emptydesk")
	for i := startDeskID; i <= endDeskID; i++ {
		deskIDs = append(deskIDs, fmt.Sprintf("%d", i))
	}

	//随机打乱
	r := rand.New(rand.NewSource(time.Now().Unix()))
	randdeskIDs := []interface{}{}
	for _, i := range r.Perm(len(deskIDs)) {
		randdeskIDs = append(randdeskIDs, deskIDs[i])
	}
	_, err := redisCli.SAdd("emptydesk", randdeskIDs...).Result()
	if err != nil {
		panic(err.Error())
	}
	//reply, err := redis.Strings(c.Do("SPOP", "emptydesk", "1"))
	//if err != nil {
	//	panic(err.Error())
	//}
	//fmt.Printf("获取空座子%v",reply)
}
