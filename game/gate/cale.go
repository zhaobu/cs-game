package main

import (
	"cy/game/cache"
	"fmt"
)

func deskNoStart() {
	cursor := "0"

	for {
		//time.Sleep(time.Second * time.Duration(rand.Intn(60)+3*60))

		keys, next, err := cache.ScanDeskInfo(cursor, "deskinfo:*")
		if err != nil {
			continue
		}

		checkDeskInfo(keys)

		cursor = next
		if cursor == "0" {
			break
		}
	}
}

func checkDeskInfo(keys []string) {
	for _, v := range keys {
		var deskID uint64
		n, err := fmt.Sscanf(v, "deskinfo:%d", &deskID)
		if err != nil || n != 1 {
			continue
		}
		fmt.Println(cache.QueryDeskInfo(deskID))
	}
}
