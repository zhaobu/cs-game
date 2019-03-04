package main

import (
	"cy/game/db/mgo"
	"fmt"
)

func main() {
	if err := mgo.Init("mongodb://192.168.1.128:27017/game"); err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(mgo.CreateClub("zztest1111", 1111, "notice1111", "arg1111"))

	//fmt.Println(mgo.UpdateClub(4, "notice4", "arg4"))

	//fmt.Println(mgo.DestoryClub(7))

	//fmt.Println(mgo.JoinClub(8, 102))
	//fmt.Println(mgo.JoinClub(9, 102))
	// fmt.Println(mgo.ExitClub(5, 101))
	//fmt.Println(mgo.JoinClub2(8, 101))
	//fmt.Println(mgo.QueryClubByID(1))

	xx, err := mgo.QueryClubByMember(10111)
	fmt.Println(err)
	for _, v := range xx {
		fmt.Println(v)
	}
}
