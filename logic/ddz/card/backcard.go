package card

import (
	"game/pb/game/ddz"
)

func BackCardType2Mul(ct pbgame_ddz.BackCardType) uint32 {
	switch ct {
	case pbgame_ddz.BackCardType_BcCtThree:
		return 2
	case pbgame_ddz.BackCardType_BcCtSameColor:
		return 2
	case pbgame_ddz.BackCardType_BcCtChain:
		return 2
	case pbgame_ddz.BackCardType_BcCtJoker2:
		return 3
	case pbgame_ddz.BackCardType_BcCtJoker1:
		return 2
	default:
		return 1
	}
}

func (sc *SetCard) CalcBackCardType() pbgame_ddz.BackCardType {
	if sc.Len() != 3 {
		return pbgame_ddz.BackCardType_BcCtNormal
	}

	if sc.cards[0].number == sc.cards[1].number && sc.cards[0].number == sc.cards[2].number {
		return pbgame_ddz.BackCardType_BcCtThree
	}

	if sc.cards[0].color == sc.cards[1].color && sc.cards[0].color == sc.cards[2].color {
		return pbgame_ddz.BackCardType_BcCtSameColor
	}

	if sc.cards[0].level == sc.cards[1].level+1 && sc.cards[0].level == sc.cards[2].level+2 {
		return pbgame_ddz.BackCardType_BcCtChain
	}

	jokerCnt := 0
	for _, v := range sc.cards {
		if v.level > 13 {
			jokerCnt++
		}
	}

	if jokerCnt == 2 {
		return pbgame_ddz.BackCardType_BcCtJoker2
	} else if jokerCnt == 1 {
		return pbgame_ddz.BackCardType_BcCtJoker1
	}

	return pbgame_ddz.BackCardType_BcCtNormal
}
