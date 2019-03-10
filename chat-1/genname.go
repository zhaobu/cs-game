package main

import (
	"math/rand"
	"time"
)

var (
	charSet []byte = []byte(`abcdefghijklnmopqrstuvwxyzABCDEFGHIJKLNMOPQRSTUVWXYZ0123456789`)
)

func init() {
	rand.Seed(time.Now().Unix())
}

func genRandName(size int) string {
	r := make([]byte, size)
	for i := 0; i < size; i++ {
		r[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(r)
}
