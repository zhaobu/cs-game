package net

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

const (
	NetKey string = "jo0REfekb*1sMhM6"
)

var(
	netAddr string
)

func Init(Addr string){
	netAddr = Addr
}


func md5V(str string) string  {
	h := md5.New()
	h.Write([]byte(str))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}