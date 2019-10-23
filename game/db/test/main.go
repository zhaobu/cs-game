package main

import (
	"fmt"
	"game/db/mgo"
	pbgame "game/pb/game"
)

func main() {
	if err := mgo.Init("mongodb://127.0.0.1:27017/game"); err != nil {
		fmt.Println(err)
		return
	}
	// data, err := mgo.QueryUserPointcard(1)
	// if err == nil {
	// 	fmt.Println(data)
	// }
	code, data := mgo.ExchangePointcard(1, "11571734910")
	if code == 0 {
		userinfo, err := mgo.UpdateWealth(1, pbgame.FeeType_FTMasonry, int64(data.ExchangeNum))
		if err == nil {
			fmt.Println(userinfo)
		}
	}
	//fmt.Println(mgo.UpdateBindMobile(11199, "15019439545"))
	//fmt.Println(mgo.QueryUserInfo(99))
	//fmt.Println(mgo.QueryUserByMobile("15019439545"))
	// fmt.Println(mgo.UpdateBindMobile(0, "15019439545", "123456"))

}

// func testWealth() {
// 	//fmt.Println(mgo.UpdateWealth(14, 2, 8))
// 	//fmt.Println(mgo.UpdateWealthPre(14, 2, 2))
// 	fmt.Println(mgo.UpdateWealthPreSure(14, pbgame.FeeType_FTMasonry, 2))
// }

// func testClub() {
// 	fmt.Println(mgo.CreateClub("zztest1111", 1111, "notice1111", "arg1111"))

// 	fmt.Println(mgo.UpdateClub(4, "notice4", "arg4"))

// 	fmt.Println(mgo.RemoveClub(7))

// 	fmt.Println(mgo.JoinClub(8, 102))
// 	fmt.Println(mgo.JoinClub(9, 102))
// 	fmt.Println(mgo.ExitClub(5, 101))
// 	fmt.Println(mgo.QueryClubByID(1))

// 	xx, err := mgo.QueryClubByMember(10111)
// 	fmt.Println(err)
// 	for _, v := range xx {
// 		fmt.Println(v)
// 	}
// }
