package dh

import (
	"math"
	"math/big"
	"math/rand"
	"time"
)

var (
	g        *big.Int
	p        *big.Int
	maxInt64 = big.NewInt(math.MaxInt64)
	rng      = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func init() {
	p, _ = big.NewInt(0).SetString(`0x83e9e6251f97702e9b5f2db1fc3497ff54df5f3fbfa11600c5eb082340c471d0683db07c5a4f9b9b4b89b6b094a6c34117fac2af85e3a891fe006903a528a56f`, 0)
	g, _ = big.NewInt(0).SetString(`0x2`, 0)
}

func GenDHpair() (dhPri *big.Int, dhPub *big.Int) {
	myPri := big.NewInt(0).Rand(rng, maxInt64)
	myPub := big.NewInt(0).Exp(g, myPri, p)
	return myPri, myPub
}

func CalcAgreeKey(myPri, otherPub *big.Int) *big.Int {
	agreeKey := big.NewInt(0).Exp(otherPub, myPri, p)
	return agreeKey
}
