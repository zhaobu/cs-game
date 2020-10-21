package card

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	oneCards = []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d,
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d,
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
		0x4e, 0x4f,
	}

	descCards = map[byte]string{
		0x01: "黑桃A", 0x02: "黑桃2", 0x03: "黑桃3", 0x04: "黑桃4", 0x05: "黑桃5", 0x06: "黑桃6", 0x07: "黑桃7", 0x08: "黑桃8", 0x09: "黑桃9", 0x0a: "黑桃10", 0x0b: "黑桃J", 0x0c: "黑桃Q", 0x0d: "黑桃K",
		0x11: "红桃A", 0x12: "红桃2", 0x13: "红桃3", 0x14: "红桃4", 0x15: "红桃5", 0x16: "红桃6", 0x17: "红桃7", 0x18: "红桃8", 0x19: "红桃9", 0x1a: "红桃10", 0x1b: "红桃J", 0x1c: "红桃Q", 0x1d: "红桃K",
		0x21: "梅花A", 0x22: "梅花2", 0x23: "梅花3", 0x24: "梅花4", 0x25: "梅花5", 0x26: "梅花6", 0x27: "梅花7", 0x28: "梅花8", 0x29: "梅花9", 0x2a: "梅花10", 0x2b: "梅花J", 0x2c: "梅花Q", 0x2d: "梅花K",
		0x31: "方块A", 0x32: "方块2", 0x33: "方块3", 0x34: "方块4", 0x35: "方块5", 0x36: "方块6", 0x37: "方块7", 0x38: "方块8", 0x39: "方块9", 0x3a: "方块10", 0x3b: "方块J", 0x3c: "方块Q", 0x3d: "方块K",
		0x4e: "小王", 0x4f: "大王",
	}
)

func init() {
	rand.Seed(time.Now().Unix())
}

func cardColor(seq byte) int {
	return int((seq & 0xf0) >> 4)
}

func cardNumber(seq byte) int {
	return int(seq & 0xf)
}

func cardLevel(seq byte) int {
	return number2Level(cardNumber(seq))
}

func number2Level(n int) int {
	if n >= 3 && n <= 13 {
		return n - 2
	} else if n == 1 || n == 2 {
		return n + 11
	} else if n == 14 || n == 15 {
		return n
	}
	panic(fmt.Sprintf("bad number %d", n))
}

func level2Number(l int) int {
	if l >= 1 && l <= 11 {
		return l + 2
	} else if l == 12 || l == 13 {
		return l - 11
	} else if l == 14 || l == 15 {
		return l
	}
	panic(fmt.Sprintf("bad level %d", l))
}

type Card struct {
	Seq    byte
	color  int
	number int
	level  int
}

func (c Card) String() string {
	return descCards[c.Seq]
}

// 计算牌型
func (c *Card) calc() {
	if _, find := descCards[c.Seq]; !find {
		panic(fmt.Errorf("bad seq 0x%0x", c.Seq))
	}

	c.color = cardColor(c.Seq)
	c.number = cardNumber(c.Seq)
	c.level = cardLevel(c.Seq)
}

func RandDdzCard() (dir0, dir1, dir2, back *SetCard) {
	tmp := make([]byte, 54)
	copy(tmp, oneCards)

	for i := 0; i < 50; i++ {
		k := rand.Intn(54)
		j := rand.Intn(54)

		t := tmp[k]
		tmp[k] = tmp[j]
		tmp[j] = t
	}

	return NewSetCard(tmp[0:17]), NewSetCard(tmp[17:34]), NewSetCard(tmp[34:51]), NewSetCard(tmp[51:54])
}
