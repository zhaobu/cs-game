package main

import (
	"cy/im/cache"
	"fmt"
)

func main() {
	if err := cache.Init("192.168.0.213:6379"); err != nil {
		fmt.Println(err)
		return
	}

	// cache.UserOnline(123457, "def")
	// cache.UserOnline(123458, "ghi")
	// fmt.Println(cache.QueryUser(123457))
	// fmt.Println(cache.QueryUsers(123458, 123456))
	// cache.UserOffline(123456)

	// cache.UserEnterRoom(123456, 100)
	// cache.UserEnterRoom(123457, 100)
	// fmt.Println(cache.RoomUsers(100))
	// cache.UserExitRoom(123457, 100)
	// fmt.Println(cache.RoomUsers(100))

	// lastid, err := cache.LastReadID(123456, 123457)
	// fmt.Println(lastid, err)
	// cache.SetLastReadID(123456, 123457, lastid+1)
	// fmt.Println(cache.LastReadID(123456, 123457))

	// cache.ChangeUnreadCnt(123456, map[uint64]int64{1234569: 99})
	// fmt.Println(cache.UnreadCnt(123456))

	//fmt.Println(cache.UserFriend(123456))
	// cache.AddFriend(123456, 1)
	// cache.DelFriend(123456, 2)
	// cache.AddFriend(123456, 3)
	// fmt.Println(cache.UserFriend(123456))

	//fmt.Println(cache.AddFriendPending(123, 457))
	cache.DeleteFriendPending(457, 123)
}
