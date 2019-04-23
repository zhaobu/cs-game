package majiang

import (
	"cy/game/configs"
	"cy/game/util"
	"math/rand"
	"runtime/debug"
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
	//  一条  二条  三条 四条  五条  六条 七条  八条 九条
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	21, 22, 23, 24, 25, 26, 27, 28, 29,
	//  一筒  二筒  三筒 四筒  五筒  六筒 七筒  八筒  九筒
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	31, 32, 33, 34, 35, 36, 37, 38, 39,
	//  东 南 西  北 中 發   白(白板算花牌)
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,
	41, 42, 43, 44, 45, 46, 47,

	//春夏秋冬,梅兰竹菊
	51, 52, 53, 54, 55, 56, 57, 58,
}

//配牌解析结构
type TestHandCards struct {
	HandCards  map[int32][]int32 `json:"handCards"`  //配牌数据
	DebugCards bool              `json:"debugCards"` //是否配牌
	stackCards map[int32]int32   //所有牌的统计
}

//二人麻将
var threePlayerCardDef = []int32{}

//二人麻将
var twoPlayerCardDef = []int32{}

var (
	log       *zap.SugaredLogger //majiang package的log
	testCards TestHandCards
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

//读取配牌
func (self *CardDef) DebugCards(gameName string, baseCard []int32, playercount int32) []int32 {
	//从配牌文件读取
	util.LoadJSON(configs.Conf.GameNode[gameName].GameTest, &testCards)
	if !testCards.DebugCards {
		return RandCards(baseCard)
	}
	testCards.stackCards = map[int32]int32{}
	debugCards := []int32{} //配的牌
	for i := playercount - 1; i >= 0; i-- {
		debugCards = append(debugCards, testCards.HandCards[i]...)
		Add_stack(testCards.stackCards, testCards.HandCards[i]...)
	}
	//去掉配的牌
	baseStacks := CalStackCards(baseCard)
	for k, v := range testCards.stackCards {
		num, ok := baseStacks[k]
		if !ok || v > num {
			log.Errorf("配牌时牌%d数量太多或者牌库不存在该牌", k)
			continue
		}
		baseStacks[k] -= v
		if baseStacks[k] == 0 {
			delete(baseStacks, k)
		}
	}
	//剩下的牌随机
	leftCards := []int32{}
	for k, v := range baseStacks {
		for i := int32(0); i < v; i++ {
			leftCards = append(leftCards, k)
		}
	}
	leftCards = RandCards(leftCards)

	return append(leftCards, debugCards...)
}

//洗牌
func RandCards(baseCard []int32) []int32 {
	array := make([]int32, len(baseCard)) //保证不会改变baseCard
	copy(array, baseCard)
	rand.Seed(time.Now().Unix())
	for i := len(array) - 1; i >= 0; i-- {
		p := randInt64(0, int64(i))
		a := array[i]
		array[i] = array[p]
		array[p] = a
	}
	return array
}

// randInt64 区间随机数
func randInt64(min, max int64) int64 {
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
			if m[card] > 4 {
				log.Errorf("加牌%d时牌数量>4,stack=%s", card, string(debug.Stack()))
			}
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
			log.Errorf("减牌%d时牌数量为0,stack=%s", card, string(debug.Stack()))
		}
	}
}

//统计牌数量
func CalStackCards(rawcards []int32) map[int32]int32 {
	var newcard = make(map[int32]int32)
	for _, v := range rawcards {
		Add_stack(newcard, v)
	}
	return newcard
}

func IsHuaCard(card int32) bool {
	return card >= 51 && card <= 58 || card == 47
}

func GetHuaCount(stackCards map[int32]int32) (res int32) {
	for k, v := range stackCards {
		if IsHuaCard(k) {
			res += v
		}
	}
	return
}
