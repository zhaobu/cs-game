package majiang

import (
	"game/configs"
	"game/util"
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"
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
	HandCards  []int32         `json:"handCards"`  //配牌数据
	DebugCards bool            `json:"debugCards"` //是否配牌
	stackCards map[int32]int32 //所有牌的统计
}

//三人麻将
var threePlayerCardDef = []int32{
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

//二人麻将
var twoPlayerCardDef = fourPlayerCardDef

var (
	testCards TestHandCards
)

type CardDef struct {
	*RoomLog //桌子日志
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
		self.Log.Error("玩家人数有问题")
		return nil
	}
	return card
}

//读取配牌
func (self *CardDef) GetDebugCards(gameName string, baseCard []int32, playercount int32) []int32 {
	//从配牌文件读取
	util.LoadJSON(configs.Conf.GameNode[gameName].GameTest, &testCards)
	if !testCards.DebugCards {
		return RandCards(baseCard)
	} else {
		//检查配牌是否合法,三人麻将没有万字牌
		if playercount == 3 {
			for _, v := range testCards.HandCards {
				if GetCardColor(v) == 1 {
					return RandCards(baseCard)
				}
			}
		}
	}

	//随机剩下的牌
	leftCards := RandCards(DelCards(nil, testCards.HandCards, baseCard))
	return append(leftCards, ReversaCards(testCards.HandCards)...)
}

//反转牌
func ReversaCards(cards []int32) []int32 {
	for i, j := 0, len(cards)-1; i < j; i, j = i+1, j-1 {
		cards[i], cards[j] = cards[j], cards[i]
	}
	return cards
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

func GetNextChair(chairId, playerCount int32) int32 {
	if chairId+1 == playerCount {
		return 0
	}
	return chairId + 1
}

//发牌
func (self *CardDef) DealCard(rawcards []int32, playercount, bankerID int32) (handCards [][]int32, leftCards []int32) {
	player_cards := make([][]int32, playercount)
	var leftNum = len(rawcards)                                                                 //剩下的牌数量
	for i, j := bankerID, int32(0); j < playercount; i, j = GetNextChair(i, playercount), j+1 { //保证配牌时,前14张发给庄家
		player_cards[i] = make([]int32, 0, 14)
		for index := 0; index < 13; index++ {
			player_cards[i] = append(player_cards[i], rawcards[leftNum-1])
			leftNum--
		}
		//庄家多摸一张牌
		if i == bankerID {
			player_cards[i] = append(player_cards[i], rawcards[leftNum-1])
			leftNum--
		}
	}

	self.Log.Warnf("玩家手牌为%+v", player_cards)
	return player_cards, rawcards[:leftNum] //返回所有玩家摸到的牌和随机牌库剩下的牌
}

//加
func Add_stack(m map[int32]int32, cards ...int32) (err error) {
	for _, card := range cards {
		if _, ok := m[card]; ok {
			m[card] = m[card] + 1
			if m[card] > 4 {
				err = fmt.Errorf("加牌%d时牌数量>4,stack=%s", card, string(debug.Stack()))
			}
		} else {
			m[card] = 1
		}
	}
	return
}

//减
func Sub_stack(m map[int32]int32, cards ...int32) (err error) {
	for _, card := range cards {
		if num, ok := m[card]; ok {
			m[card] = num - 1
			if num == 1 {
				delete(m, card)
			}
		} else {
			err = fmt.Errorf("减牌%d时牌数量为0,stack=%s", card, string(debug.Stack()))
		}
	}
	return
}

//统计牌数量(withOutHua为true时表示不统计花牌)
func CalStackCards(rawcards []int32, withOutHua bool) map[int32]int32 {
	var newcard = make(map[int32]int32)
	for _, v := range rawcards {
		if withOutHua && IsHuaCard(v) {
		} else {
			Add_stack(newcard, v)
		}
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

//是否合法
func IsVaildCard(card int32) bool {
	if card >= 11 && card <= 19 || card >= 21 && card <= 29 || card >= 31 && card <= 39 {
		return true
	} else if card >= 41 && card <= 47 || card >= 51 && card <= 58 {
		return true
	}
	return false
}

//从allCards中去掉cards,返回剩下的牌
func DelCards(cardsStack map[int32]int32, cards, allCards []int32) (leftCards []int32) {
	if cardsStack == nil {
		cardsStack = CalStackCards(cards, false)
	}
	for _, v := range allCards { //找到没有指定的牌,保存
		if num, ok := cardsStack[v]; ok {
			cardsStack[v]--
			if num == 1 {
				delete(cardsStack, v)
			}
		} else {
			leftCards = append(leftCards, v)
		}
	}
	return
}

//客户端配初始牌,然后写入文件
func (self *CardDef) DebugCardsFromClient(gameName string, debugCards []int32) {
	tmp := &TestHandCards{HandCards: debugCards, DebugCards: true}
	//写入到文件中
	util.WriteJSON(configs.Conf.GameNode[gameName].GameTest, tmp)
}

func GetCardColor(card int32) int32 {
	return card / 10
}

func GetCardValue(card int32) int32 {
	return card % 10
}
