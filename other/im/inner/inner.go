package inner

import (
	"fmt"
)

func StoreKey(toid uint64) string {
	return fmt.Sprintf("uid:%d", toid)
}

func IdFromStoreKey(s string) (id string) {
	fmt.Sscanf(s, "uid:%s", &id)
	return
}

func SessionID(fromid, toid uint64, broadCast, multiCast bool) string {
	if broadCast {
		return fmt.Sprintf("b_%d", toid)
	}
	if multiCast {
		return fmt.Sprintf("m_%d", toid)
	}
	if fromid >= toid {
		return fmt.Sprintf("%d_%d", fromid, toid)
	}
	return fmt.Sprintf("%d_%d", toid, fromid)
}
