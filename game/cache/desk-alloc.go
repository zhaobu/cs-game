package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func AllocDeskID() (deskID uint64, err error) {
	cmdInt, err := redisCli.SCard("emptydesk").Result()
	if err != nil {
		return 0, err
	}
	if cmdInt == 0 {
		rand.Seed(time.Now().Unix())
		var num int = 1000
		codeMap := make(map[int64]bool, num)
		for {
			if len(codeMap) >= num {
				break
			}
			enter_code := rand.Int63n(999999-100000) + 100000
			if _, ok := codeMap[enter_code]; ok {
				continue
			}
			codeMap[enter_code] = true
			_, err := redisCli.SAdd("emptydesk", enter_code).Result()
			if err != nil {
				return 0, err
			}
		}
		// for {
		// 	enter_code := rand.Int63n(999999-100000) + 100000
		// 	reply, err := redis.Int(c.Do("SADD", "emptydesk", enter_code))
		// 	if err != nil {
		// 		return 0, err
		// 	}
		// 	if reply == 1 && num == 500 {
		// 		break
		// 	}
		// 	if num >= 30000 {
		// 		break
		// 	}
		// 	num++
		// }
	}
	cmdStr := redisCli.SPop("emptydesk")
	return cmdStr.Uint64()
}

func FreeDeskID(deskID uint64) (err error) {
	_, err = redisCli.SAdd("emptydesk", strconv.FormatUint(deskID, 10)).Result()
	return
}

func SCAN(pattern string, count int) (find []string) {
	if count < 1 || count > 50 {
		count = 50
	}

	var (
		cursor uint64
		n      int
		err    error
	)

	for {
		find, cursor, err = redisCli.Scan(cursor, pattern, int64(count)).Result()
		if err != nil {
			fmt.Printf("err=%s", err)
		}
		n += len(find)
		if cursor == 0 {
			break
		}
	}
	return
}
