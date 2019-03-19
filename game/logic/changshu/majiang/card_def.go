package majiang

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

//card_def文件写对牌的定义

//四人麻将
var fourPlayerCardDef = []uint8{
	//  一万二万三万四万五万六万七八万九万
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	11, 12, 13, 14, 15, 16, 17, 18, 19,
	//  一筒  二筒  三筒 四筒  五筒  六筒 七筒  八筒  九筒
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	//  一条  二条  三条 四条  五条  六条 七条  八条 九条
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	//  东 南 西  北 中 發 白
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,

	//春夏秋冬,梅兰竹菊
	51, 52, 53, 54, 55, 56, 57, 58,
}

//二人麻将
var threePlayerCardDef = []uint8{}

//二人麻将
var twoPlayerCardDef = []uint8{}

var (
	log *logrus.Entry //majiang package的log
)

type CardDef struct {
}

func (self *CardDef) Init(logptr *logrus.Entry) {
	log = logptr
}

func (self *CardDef) GetBaseCard(playerCount int32) []uint8 {
	var card []uint8
	if playerCount == 4 {
		card = make([]uint8, len(fourPlayerCardDef))
		copy(card, fourPlayerCardDef)
	} else if playerCount == 3 {
		card = make([]uint8, len(threePlayerCardDef))
		copy(card, threePlayerCardDef)
	} else if playerCount == 2 {
		card = make([]uint8, len(twoPlayerCardDef))
		copy(card, twoPlayerCardDef)
	} else {
		log.Error("玩家人数有问题")
		return nil
	}
	return card
}

//洗牌
func (self *CardDef) RandCards(baseCard []uint8) []uint8 {
	array := make([]uint8, len(baseCard)) //保证不会改变baseCard
	copy(array, baseCard)
	rand.Seed(time.Now().Unix())
	for i := len(array) - 1; i >= 0; i-- {
		p := self.randInt64(0, int64(i))
		a := array[i]
		array[i] = array[p]
		array[p] = a
	}
	return array
}

// randInt64 区间随机数
func (self *CardDef) randInt64(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}

//发牌
func (self *CardDef) DealCard(rawcards []uint8, playercount, bankerID int32) (handCards [][]uint8, leftCards []uint8) {
	player_cards := make([][]uint8, playercount)
	var leftNum = len(rawcards) //剩下的牌数量
	for i := int32(0); i < playercount; i++ {
		//庄家多摸一张牌
		if i == bankerID {
			player_cards[i] = make([]uint8, 14)
			player_cards[i][13] = rawcards[leftNum-1]
			leftNum--
		} else {
			player_cards[i] = make([]uint8, 13)
		}
		for index := 0; index < 13; index++ {
			player_cards[i][index] = rawcards[leftNum-1]
			leftNum--
		}
	}

	log.Warnf("玩家手牌为%+v", player_cards)
	return player_cards, rawcards[:leftNum] //返回所有玩家摸到的牌和随机牌库剩下的牌
}

//加
func (self *CardDef) add_stack(m map[uint8]int32, card uint8) {
	if _, ok := m[card]; ok {
		m[card] = m[card] + 1
	} else {
		m[card] = 1
	}
}

//减
func (self *CardDef) sub_stack(m map[uint8]int32, card uint8) {
	num, ok := m[card]
	if ok == false {
		log.Errorf("减牌%d时牌数量为0", card)
	} else if num == 1 {
		delete(m, card)
	} else {
		m[card] = num - 1
	}
}

//统计牌数量
func (self *CardDef) StackCards(rawcards []uint8) map[uint8]int32 {
	var newcard = make(map[uint8]int32)
	for _, v := range rawcards {
		self.add_stack(newcard, v)
	}
	return newcard
}
