package main

import (
	"cy/game/cache"
	"fmt"
)

func main() {
	cache.Init("192.168.1.128:6379", 1)

	//fmt.Println(cache.AddClubDeskRelation(10, 100))
	//fmt.Println(cache.AddClubDeskRelation(10, 102))
	//fmt.Println(cache.AddDeskInfo(&pbcommon.DeskInfo{ID: 100, ClubID: 10}))
	fmt.Println(cache.DeleteClubDeskRelation(100))

	//fmt.Println(cache.QueryClubDeskInfo(10))
	// cursor := "0"
	// var keys []string
	// for {
	// 	k, n, err := cache.ScanDeskInfo(cursor, "deskinfo:*")
	// 	if err != nil {
	// 		break
	// 	}

	// 	keys = append(keys, k...)

	// 	if n == "0" {
	// 		break
	// 	}

	// 	cursor = n
	// }
	// fmt.Println(keys)
}
