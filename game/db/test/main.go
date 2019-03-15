package main

import (
	"cy/game/db/mgo"
	"fmt"
)

func main() {
	if err := mgo.Init("mongodb://192.168.0.90:27017/game"); err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(mgo.UpdateBindMobile(11199, "15019439545"))
	//fmt.Println(mgo.QueryUserInfo(99))
	//fmt.Println(mgo.QueryUserByMobile("15019439545"))
	fmt.Println(mgo.UpdateBindMobile(0, "15019439545", "123456"))

}

func testWealth() {
	//fmt.Println(mgo.UpdateWealth(14, 2, 8))
	//fmt.Println(mgo.UpdateWealthPre(14, 2, 2))
	fmt.Println(mgo.UpdateWealthPreSure(14, 2, 2))
}

func testClub() {
	fmt.Println(mgo.CreateClub("zztest1111", 1111, "notice1111", "arg1111"))

	fmt.Println(mgo.UpdateClub(4, "notice4", "arg4"))

	fmt.Println(mgo.RemoveClub(7))

	fmt.Println(mgo.JoinClub(8, 102))
	fmt.Println(mgo.JoinClub(9, 102))
	fmt.Println(mgo.ExitClub(5, 101))
	fmt.Println(mgo.QueryClubByID(1))

	xx, err := mgo.QueryClubByMember(10111)
	fmt.Println(err)
	for _, v := range xx {
		fmt.Println(v)
	}
}
