package card

import (
	"cy/game/pb/game/ddz"
)

// Tips 提示
func (sc *SetCard) Tips(last *SetCard) (find []*SetCard) {
	last.calc()
	if last.ct == pbgame_ddz.CardType_CtUnknown {
		return
	}

	switch last.ct {
	case pbgame_ddz.CardType_CtJokers:
		return
	case pbgame_ddz.CardType_CtBomb:
		return sc.tipBomb(last.level)
	case pbgame_ddz.CardType_CtSolo:
		return sc.tipSolo(last.level)
	case pbgame_ddz.CardType_CtPair:
		return sc.tipPair(last.level)
	case pbgame_ddz.CardType_CtThree:
		return sc.tipThree(last.level)
	case pbgame_ddz.CardType_CtThreeSolo:
		return sc.tipThreeSolo(last.level)
	case pbgame_ddz.CardType_CtThreePair:
		return sc.tipThreePair(last.level)
	case pbgame_ddz.CardType_CtSoloChain:
		return sc.tipSoloChain(last.level, last.length)
	case pbgame_ddz.CardType_CtPairChain:
		return sc.tipPairChain(last.level, last.length)
	case pbgame_ddz.CardType_CtThreeChain:
		return sc.tipThreeChain(last.level, last.length)
	case pbgame_ddz.CardType_CtThreeSoloChain:
		return sc.tipThreeSoloChain(last.level, last.length)
	case pbgame_ddz.CardType_CtThreePairChain:
		return sc.tipThreePairChain(last.level, last.length)
	case pbgame_ddz.CardType_CtFour2:
		return sc.tipFour2(last.level)
	case pbgame_ddz.CardType_CtFour4:
		return sc.tipFour4(last.level)
	}
	return
}

func (sc *SetCard) HaveBiger(other *SetCard) bool {
	return len(sc.Tips(other)) > 0
}

func (sc *SetCard) getSeqsByLevel(lv int, cnt int) (seqs []byte, ok bool) {
	for _, v := range sc.cards {
		if v.level == lv {
			seqs = append(seqs, v.Seq)
			if len(seqs) == cnt {
				ok = true
				return
			}
		}
	}
	return
}

func (sc *SetCard) haveContinueCard(begin, size, count int) bool {
	for i := begin; i > begin-size; i-- {
		if i >= len(sc.count) || i < 0 {
			return false
		}
		if sc.count[i] < count {
			return false
		}
	}
	return true
}

func (sc *SetCard) biggerTip(ct pbgame_ddz.CardType) (find []*SetCard) {
	if ct > pbgame_ddz.CardType_CtBomb {
		find = sc.tipBomb(-1)
	}
	if sc.haveJoker() {
		find = append(find, NewSetCard([]byte{0x4e, 0x4f}))
	}
	return
}

func (sc *SetCard) tipBomb(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 4 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 4)
			find = append(find, NewSetCard(seqs))
		}
	}
	return
}

func (sc *SetCard) tipSolo(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 1 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 1)
			find = append(find, NewSetCard(seqs))
		}
	}
	for idx, v := range sc.count {
		if v == 2 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 1)
			find = append(find, NewSetCard(seqs))
		}
	}
	for idx, v := range sc.count {
		if v == 3 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 1)
			find = append(find, NewSetCard(seqs))
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtSolo); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipPair(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 2 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 2)
			find = append(find, NewSetCard(seqs))
		}
	}
	for idx, v := range sc.count {
		if v == 3 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 2)
			find = append(find, NewSetCard(seqs))
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtPair); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThree(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 3 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 3)
			find = append(find, NewSetCard(seqs))
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThree); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThreeSolo(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 3 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 3)
			xx := NewSetCard(seqs)
			seqs2 := sc.littleSolo([]int{idx})
			if seqs2 != nil {
				xx.Add(seqs2)
				find = append(find, xx)
			}
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThreeSolo); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThreePair(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 3 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 3)
			xx := NewSetCard(seqs)
			seqs2 := sc.littlePair([]int{idx})
			if seqs2 != nil {
				xx.Add(seqs2)
				find = append(find, xx)
			}
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThreePair); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipSoloChain(bigLevel int, len int) (find []*SetCard) {
	for start := bigLevel + 1; start < 13; start++ {
		if sc.haveContinueCard(start, len, 1) {
			s := NewSetCard([]byte{})
			for i := start; i > (start - len); i-- {
				seqs, _ := sc.getSeqsByLevel(i, 1)
				s.Add(seqs)
			}
			find = append(find, s)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtSoloChain); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipPairChain(bigLevel int, len int) (find []*SetCard) {
	for start := bigLevel + 1; start < 13; start++ {
		if sc.haveContinueCard(start, len, 2) {
			s := NewSetCard([]byte{})
			for i := start; i > (start - len); i-- {
				seqs, _ := sc.getSeqsByLevel(i, 2)
				s.Add(seqs)
			}
			find = append(find, s)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtPairChain); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThreeChain(bigLevel int, len int) (find []*SetCard) {
	for start := bigLevel + 1; start < 13; start++ {
		if sc.haveContinueCard(start, len, 3) {
			s := NewSetCard([]byte{})
			for i := start; i > (start - len); i-- {
				seqs, _ := sc.getSeqsByLevel(i, 3)
				s.Add(seqs)
			}
			find = append(find, s)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThreeChain); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThreeSoloChain(bigLevel int, len int) (find []*SetCard) {
	for start := bigLevel + 1; start < 13; start++ {
		if sc.haveContinueCard(start, len, 3) {
			s := NewSetCard([]byte{})
			for i := start; i > (start - len); i-- {
				seqs, _ := sc.getSeqsByLevel(i, 3)
				s.Add(seqs)
			}

			notLevel := []int{}
			for zz := 0; zz < len; zz++ {
				notLevel = append(notLevel, start-zz)
			}
			for soloIdx := 0; soloIdx < len; soloIdx++ {
				soloX := sc.littleSolo(notLevel)
				if soloX == nil {
					return
				}
				s.Add(soloX)
				notLevel = append(notLevel, cardLevel(soloX[0]))
			}
			find = append(find, s)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThreeSoloChain); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipThreePairChain(bigLevel int, len int) (find []*SetCard) {
	for start := bigLevel + 1; start < 13; start++ {
		if sc.haveContinueCard(start, len, 3) {
			s := NewSetCard([]byte{})
			for i := start; i > (start - len); i-- {
				seqs, _ := sc.getSeqsByLevel(i, 3)
				s.Add(seqs)
			}

			notLevel := []int{}
			for zz := 0; zz < len; zz++ {
				notLevel = append(notLevel, start-zz)
			}
			for soloIdx := 0; soloIdx < len; soloIdx++ {
				soloX := sc.littlePair(notLevel)
				if soloX == nil {
					return
				}
				s.Add(soloX)
				notLevel = append(notLevel, cardLevel(soloX[0]))
			}
			find = append(find, s)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtThreePairChain); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipFour2(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 4 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 4)
			xx := NewSetCard(seqs)

			seqs2 := sc.littlePair([]int{idx})
			if seqs2 != nil {
				xx.Add(seqs2)
				find = append(find, xx)
			} else {
				seq3 := sc.littleSolo([]int{idx})
				if seq3 == nil || len(seq3) != 1 {
					return
				}
				idx3 := cardLevel(seq3[0])
				seq4 := sc.littleSolo([]int{idx, idx3})
				if seq4 == nil {
					return
				}
				xx.Add(seq3)
				xx.Add(seq4)
				find = append(find, xx)
			}
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtFour2); bf != nil {
		find = append(find, bf...)
	}
	return
}

func (sc *SetCard) tipFour4(bigLevel int) (find []*SetCard) {
	for idx, v := range sc.count {
		if v == 4 && idx > bigLevel {
			seqs, _ := sc.getSeqsByLevel(idx, 4)
			xx := NewSetCard(seqs)

			seqs2 := sc.littlePair([]int{idx})
			if seqs2 == nil {
				return
			}
			idx2 := cardLevel(seqs2[0])
			seqs3 := sc.littlePair([]int{idx, idx2})
			if seqs3 == nil {
				return
			}

			xx.Add(seqs2)
			xx.Add(seqs3)
			find = append(find, xx)
		}
	}
	if bf := sc.biggerTip(pbgame_ddz.CardType_CtFour4); bf != nil {
		find = append(find, bf...)
	}
	return
}

// 找一个最小且不为notlevel单张
func (sc *SetCard) littleSolo(notLevel []int) []byte {
	for idx, v := range sc.count {
		if v == 0 {
			continue
		}

		find := false
		for _, v2 := range notLevel {
			if v2 == idx {
				find = true
				break
			}
		}
		if find {
			continue
		}

		seqs, _ := sc.getSeqsByLevel(idx, 1)
		return seqs
	}
	return nil
}

// 找一个最小且不为notlevel对子
func (sc *SetCard) littlePair(notLevel []int) []byte {
	for idx, v := range sc.count {
		if v < 2 {
			continue
		}

		find := false
		for _, v2 := range notLevel {
			if v2 == idx {
				find = true
				break
			}
		}
		if find {
			continue
		}

		seqs, _ := sc.getSeqsByLevel(idx, 2)
		return seqs
	}
	return nil
}
