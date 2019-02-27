package card

import (
	"cy/game/pb/game/ddz"
	"sort"
	"strings"
)

var (
	checkCardType = make(map[int][](func(*SetCard) bool))
)

func init() {
	checkCardType[1] = []func(*SetCard) bool{checkSolo}
	checkCardType[2] = []func(*SetCard) bool{checkPair, checkJokers}
	checkCardType[3] = []func(*SetCard) bool{checkThree}
	checkCardType[4] = []func(*SetCard) bool{checkThreeSolo, checkBomb}
	checkCardType[5] = []func(*SetCard) bool{checkSoloChain, checkThreePair}
	checkCardType[6] = []func(*SetCard) bool{checkSoloChain, checkPairChain, checkThreeChain, checkFour2}
	checkCardType[7] = []func(*SetCard) bool{checkSoloChain}
	checkCardType[8] = []func(*SetCard) bool{checkSoloChain, checkPairChain, checkThreeSoloChain, checkFour4}
	checkCardType[9] = []func(*SetCard) bool{checkSoloChain, checkThreeChain}
	checkCardType[10] = []func(*SetCard) bool{checkSoloChain, checkPairChain, checkThreePairChain}
	checkCardType[11] = []func(*SetCard) bool{checkSoloChain}
	checkCardType[12] = []func(*SetCard) bool{checkSoloChain, checkPairChain, checkThreeChain, checkThreeSoloChain}
	checkCardType[14] = []func(*SetCard) bool{checkPairChain}
	checkCardType[15] = []func(*SetCard) bool{checkThreeChain, checkThreePairChain}
	checkCardType[16] = []func(*SetCard) bool{checkPairChain, checkThreeSoloChain}
	checkCardType[18] = []func(*SetCard) bool{checkPairChain, checkThreeChain}
	checkCardType[20] = []func(*SetCard) bool{checkPairChain, checkThreeSoloChain, checkThreePairChain}
}

// SetCard 一组牌
type SetCard struct {
	seqs   []byte // 不会排序
	cards  []Card
	count  [16]int
	ct     pbgame_ddz.CardType
	length int
	level  int
	calced bool
}

func NewSetCard(seqs []byte) *SetCard {
	sc := &SetCard{}
	sc.seqs = make([]byte, len(seqs))
	copy(sc.seqs, seqs)
	sc.refash()

	return sc
}

func (sc *SetCard) refash() {
	sc.calced = false
	sc.cards = make([]Card, len(sc.seqs))
	sc.count = [16]int{}

	for i := 0; i < len(sc.seqs); i++ {
		sc.cards[i].Seq = sc.seqs[i]
		sc.cards[i].calc()
		sc.count[sc.cards[i].level]++
	}

	sort.Slice(sc.cards, func(i, j int) bool {
		return sc.cards[i].level > sc.cards[j].level
	})

	sc.calc()
}

func (sc *SetCard) String() string {
	b := strings.Builder{}
	for _, v := range sc.cards {
		b.WriteString(v.String() + " ")
	}
	return b.String()
}

func (sc SetCard) haveJoker() bool {
	if sc.count[14] == 1 && sc.count[15] == 1 {
		return true
	}
	return false
}

func (sc SetCard) IsEmpty() bool {
	return len(sc.seqs) == 0
}

func (sc SetCard) Len() int {
	return len(sc.seqs)
}

func (sc SetCard) Dump() []byte {
	t := make([]byte, len(sc.seqs))
	copy(t, sc.seqs)
	return t
}

func (sc *SetCard) Add(n []byte) {
	sc.seqs = append(sc.seqs, n...)
	sc.refash()
}

func (sc *SetCard) Del(n []byte) {
	for _, v := range n {
		for i := 0; i < len(sc.seqs); i++ {
			if sc.seqs[i] == v {
				sc.seqs[i] = sc.seqs[len(sc.seqs)-1]
				sc.seqs = sc.seqs[:len(sc.seqs)-1]
				break
			}
		}
	}
	sc.refash()
}

func (sc *SetCard) Type() pbgame_ddz.CardType {
	sc.calc()
	return sc.ct
}

func (sc *SetCard) Biger(other *SetCard) bool {
	sc.calc()
	other.calc()

	if sc.ct == pbgame_ddz.CardType_CtUnknown ||
		other.ct == pbgame_ddz.CardType_CtUnknown {
		return false
	}

	if sc.ct == other.ct {
		if sc.length == other.length && sc.level > other.level {
			return true
		}
		return false
	}
	if sc.ct <= pbgame_ddz.CardType_CtBomb && sc.ct < other.ct {
		return true
	}
	return false
}

func (sc *SetCard) Have(other []byte) bool {
	for _, seq := range other {
		var have bool
		for _, seq2 := range sc.seqs {
			if seq == seq2 {
				have = true
				break
			}
		}
		if !have {
			return false
		}
	}
	return true
}

func checkSolo(sc *SetCard) bool {
	if len(sc.cards) == 1 {
		sc.ct = pbgame_ddz.CardType_CtSolo
		sc.level = sc.cards[0].level
		return true
	}
	return false
}

func checkSoloChain(sc *SetCard) bool {
	length := len(sc.cards)
	if length >= 5 && length < 13 {
		level := sc.cards[0].level
		if level > number2Level(1) {
			return false
		}

		shouldLen := length
		if sc.continueCard(level, shouldLen, 1) {
			sc.ct = pbgame_ddz.CardType_CtSoloChain
			sc.length = shouldLen
			sc.level = level
			return true
		}
	}
	return false
}

func checkPair(sc *SetCard) bool {
	if len(sc.cards) == 2 && sc.cards[0].level == sc.cards[1].level {
		sc.ct = pbgame_ddz.CardType_CtPair
		sc.level = sc.cards[0].level
		return true
	}
	return false
}

func checkPairChain(sc *SetCard) bool {
	length := len(sc.cards)
	if length%2 == 0 && length >= 6 && length <= 20 {
		level := sc.cards[0].level
		if level > number2Level(1) {
			return false
		}

		shouldLen := length / 2
		if sc.continueCard(level, shouldLen, 2) {
			sc.ct = pbgame_ddz.CardType_CtPairChain
			sc.length = shouldLen
			sc.level = level
			return true
		}
	}
	return false
}

func checkThree(sc *SetCard) bool {
	if len(sc.cards) == 3 &&
		sc.cards[0].level == sc.cards[1].level &&
		sc.cards[0].level == sc.cards[2].level {
		sc.ct = pbgame_ddz.CardType_CtThree
		sc.level = sc.cards[0].level
		return true
	}
	return false
}

func checkThreeChain(sc *SetCard) bool {
	length := len(sc.cards)
	if length%3 == 0 && length >= 6 && length <= 18 {
		level := sc.cards[0].level
		if level > number2Level(1) {
			return false
		}

		shouldLen := length / 3
		if sc.continueCard(level, shouldLen, 3) {
			sc.ct = pbgame_ddz.CardType_CtThreeChain
			sc.length = shouldLen
			sc.level = level
			return true
		}
	}
	return false
}

func checkThreeSolo(sc *SetCard) bool {
	if len(sc.cards) == 4 {
		cnt1 := 0
		cnt3 := 0
		level := 0
		for idx, v := range sc.count {
			if v == 1 {
				cnt1++
			} else if v == 3 {
				cnt3++
				level = idx
			}
		}
		if cnt1 == 1 && cnt3 == 1 {
			sc.ct = pbgame_ddz.CardType_CtThreeSolo
			sc.level = level
			return true
		}
	}
	return false
}

func checkThreePair(sc *SetCard) bool {
	if len(sc.cards) == 5 {
		cnt2 := 0
		cnt3 := 0
		level := 0
		for idx, v := range sc.count {
			if v == 2 {
				cnt2++
			} else if v == 3 {
				cnt3++
				level = idx
			}
		}
		if cnt2 == 1 && cnt3 == 1 {
			sc.ct = pbgame_ddz.CardType_CtThreePair
			sc.level = level
			return true
		}
	}
	return false
}

func checkThreeSoloChain(sc *SetCard) bool {
	length := len(sc.cards)
	if length%4 == 0 && length >= 8 && length <= 20 {
		cnt1 := 0
		cnt3 := 0

		for _, v := range sc.count {
			if v == 1 {
				cnt1++
			} else if v == 3 {
				cnt3++
			}
		}

		shouldLen := length / 4

		if cnt3 != shouldLen || cnt1 != shouldLen {
			return false
		}

		for idx, v := range sc.count {
			if v == 3 {
				if sc.continueCard(idx+shouldLen-1, shouldLen, 3) {
					sc.ct = pbgame_ddz.CardType_CtThreeSoloChain
					sc.length = shouldLen
					sc.level = idx + shouldLen - 1
					return true
				}
				break
			}
		}
	}
	return false
}

func checkThreePairChain(sc *SetCard) bool {
	length := len(sc.cards)
	if length%5 == 0 && length >= 10 && length <= 20 {
		cnt2 := 0
		cnt3 := 0

		for _, v := range sc.count {
			if v == 2 {
				cnt2++
			} else if v == 3 {
				cnt3++
			}
		}

		shouldLen := length / 5

		if cnt3 != shouldLen || cnt2 != shouldLen {
			return false
		}

		for idx, v := range sc.count {
			if v == 3 {
				if sc.continueCard(idx+shouldLen-1, shouldLen, 3) {
					sc.ct = pbgame_ddz.CardType_CtThreePairChain
					sc.length = shouldLen
					sc.level = idx + shouldLen - 1
					return true
				}
				break
			}
		}
	}
	return false
}

func checkFour2(sc *SetCard) bool {
	if len(sc.cards) == 6 {
		cnt4 := 0
		level := 0
		for idx, v := range sc.count {
			if v == 4 {
				cnt4++
				level = idx
				break
			}
		}
		if cnt4 == 1 {
			sc.ct = pbgame_ddz.CardType_CtFour2
			sc.level = level
			return true
		}
	}
	return false
}

func checkFour4(sc *SetCard) bool {
	if len(sc.cards) == 8 {
		cnt2 := 0
		cnt4 := 0
		level := 0
		for idx, v := range sc.count {
			if v == 4 {
				cnt4++
				level = idx
			} else if v == 2 {
				cnt2++
			}
		}
		if cnt4 == 1 && cnt2 == 2 {
			sc.ct = pbgame_ddz.CardType_CtFour4
			sc.level = level
			return true
		}
	}
	return false
}

func checkBomb(sc *SetCard) bool {
	if len(sc.cards) == 4 {
		for idx, v := range sc.count {
			if v == 4 {
				sc.ct = pbgame_ddz.CardType_CtBomb
				sc.level = idx
				return true
			}
		}
	}
	return false
}

func checkJokers(sc *SetCard) bool {
	if len(sc.cards) == 2 && sc.cards[0].Seq == 0x4f && sc.cards[1].Seq == 0x4e {
		sc.ct = pbgame_ddz.CardType_CtJokers
		return true
	}
	return false
}

func (sc *SetCard) calc() {
	if sc.calced {
		return
	}
	sc.calced = true

	for _, v := range checkCardType[len(sc.cards)] {
		if v(sc) {
			break
		}
	}
}

func (sc *SetCard) continueCard(begin, size, count int) bool {
	for i := begin; i > begin-size; i-- {
		if i >= len(sc.count) || i < 0 {
			return false
		}
		if sc.count[i] != count {
			return false
		}
	}
	return true
}
