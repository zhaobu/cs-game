package majiang

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

//card_def文件写对牌的定义

//四人麻将
var fourPlayerCardDef = []int32{
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
	//  东 南 西  北 中 發   白(白板算花牌)
	41, 42, 43, 44, 45, 46, 59,
	41, 42, 43, 44, 45, 46, 59,
	41, 42, 43, 44, 45, 46, 59,
	41, 42, 43, 44, 45, 46, 59,

	//春夏秋冬,梅兰竹菊
	51, 52, 53, 54, 55, 56, 57, 58,
}

//二人麻将
var threePlayerCardDef = []int32{}

//二人麻将
var twoPlayerCardDef = []int32{}

var (
	log *zap.SugaredLogger //majiang package的log
)

type CardDef struct {
}

func (self *CardDef) Init(logptr *zap.SugaredLogger) {
	log = logptr
}

func (self *CardDef) GetBaseCard(playerCount int32) []int32 {
	var card []int32
	if playerCount == 4 {
		card = make([]int32, len(fourPlayerCardDef))
		copy(card, fourPlayerCardDef)
	} else if playerCount == 3 {
		card = make([]int32, len(threePlayerCardDef))
		copy(card, threePlayerCardDef)
	} else if playerCount == 2 {
		card = make([]int32, len(twoPlayerCardDef))
		copy(card, twoPlayerCardDef)
	} else {
		log.Error("玩家人数有问题")
		return nil
	}
	return card
}

//洗牌
func (self *CardDef) RandCards(baseCard []int32) []int32 {
	array := make([]int32, len(baseCard)) //保证不会改变baseCard
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
func (self *CardDef) DealCard(rawcards []int32, playercount, bankerID int32) (handCards [][]int32, leftCards []int32) {
	player_cards := make([][]int32, playercount)
	var leftNum = len(rawcards) //剩下的牌数量
	for i := int32(0); i < playercount; i++ {
		//庄家多摸一张牌
		if i == bankerID {
			player_cards[i] = make([]int32, 14)
			player_cards[i][13] = rawcards[leftNum-1]
			leftNum--
		} else {
			player_cards[i] = make([]int32, 13)
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
func Add_stack(m map[int32]int32, cards ...int32) {
	for _, card := range cards {
		if _, ok := m[card]; ok {
			m[card] = m[card] + 1
		} else {
			m[card] = 1
		}
	}
}

//减
func Sub_stack(m map[int32]int32, cards ...int32) {
	for _, card := range cards {
		if num, ok := m[card]; ok {
			m[card] = num - 1
			if num == 1 {
				delete(m, card)
			}
		} else {
			log.Errorf("减牌%d时牌数量为0", card)
		}
	}
}

//统计牌数量
func (self *CardDef) StackCards(rawcards []int32) map[int32]int32 {
	var newcard = make(map[int32]int32)
	for _, v := range rawcards {
		Add_stack(newcard, v)
	}
	return newcard
}

func (self *CardDef) IsHuaCard(card int32) bool {
	return card >= 51 && card <= 59
}
