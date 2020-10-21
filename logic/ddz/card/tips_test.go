package card

import (
	"game/pb/game/ddz"
	"testing"
)

func Test_tipBomb(t *testing.T) {
	xx := NewSetCard([]byte{0x04, 0x14, 0x24, 0x34, 0x09, 0x19, 0x29, 0x39, 0x08})

	find := xx.Tips(NewSetCard([]byte{0x03, 0x13, 0x23, 0x33}))
	if len(find) != 2 {
		t.Fatalf("%v", find)
	}

	if find[0].ct != pbgame_ddz.CardType_CtBomb || find[0].level != 2 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtBomb || find[1].level != 7 {
		t.Fatalf("%v", find)
	}
}

func Test_tipSolo(t *testing.T) {
	xx := NewSetCard([]byte{0x03, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x09, 0x19, 0x29, 0x39})

	find := xx.Tips(NewSetCard([]byte{0x04}))
	if len(find) != 5 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtSolo || find[0].level != 3 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtSolo || find[1].level != 4 {
		t.Fatalf("%v", find)
	}
	if find[2].ct != pbgame_ddz.CardType_CtSolo || find[2].level != 5 {
		t.Fatalf("%v", find)
	}
	if find[3].ct != pbgame_ddz.CardType_CtSolo || find[3].level != 6 {
		t.Fatalf("%v", find)
	}
	if find[4].ct != pbgame_ddz.CardType_CtBomb || find[4].level != 7 {
		t.Fatalf("%v", find)
	}
}

func Test_tipPair(t *testing.T) {
	xx := NewSetCard([]byte{0x03, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x4e, 0x4f})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x14}))
	if len(find) != 3 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtPair || find[0].level != 5 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtPair || find[1].level != 6 {
		t.Fatalf("%v", find)
	}
	if find[2].ct != pbgame_ddz.CardType_CtJokers {
		t.Fatalf("%v", find)
	}
}

func Test_tipThree(t *testing.T) {
	xx := NewSetCard([]byte{0x03, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x09, 0x19, 0x29, 0x39})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x14, 0x24}))
	if len(find) != 1 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThree || find[0].level != 6 {
		t.Fatalf("%v", find)
	}
}

func Test_tipThreeSolo(t *testing.T) {

	xx := NewSetCard([]byte{0x03, 0x13, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x09, 0x19, 0x29, 0x39})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x14, 0x24, 0x02}))
	if len(find) != 1 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThreeSolo || find[0].level != 6 {
		t.Fatalf("%v", find)
	}
}

func Test_tipThreePair(t *testing.T) {

	xx := NewSetCard([]byte{0x03, 0x03, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x09, 0x19, 0x29, 0x39})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x14, 0x24, 0x02, 0x12}))
	if len(find) != 1 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThreePair || find[0].level != 6 {
		t.Fatalf("%v", find)
	}
}

func Test_tipSoloChain(t *testing.T) {

	xx := NewSetCard([]byte{0x01, 0x02, 0x03, 0x03, 0x05, 0x06, 0x07, 0x17, 0x08, 0x18, 0x28, 0x09, 0x19, 0x29, 0x39, 0x0a, 0x0b, 0x0c, 0x0d})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09}))
	if len(find) != 5 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtSoloChain || find[0].level != 8 || find[0].length != 6 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtSoloChain || find[1].level != 9 || find[1].length != 6 {
		t.Fatalf("%v", find)
	}
	if find[2].ct != pbgame_ddz.CardType_CtSoloChain || find[2].level != 10 || find[2].length != 6 {
		t.Fatalf("%v", find)
	}
	if find[3].ct != pbgame_ddz.CardType_CtSoloChain || find[3].level != 11 || find[3].length != 6 {
		t.Fatalf("%v", find)
	}
	if find[4].ct != pbgame_ddz.CardType_CtSoloChain || find[4].level != 12 || find[4].length != 6 {
		t.Fatalf("%v", find)
	}
}

func Test_tipPairChain(t *testing.T) {

	xx := NewSetCard([]byte{0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19}))
	if len(find) != 2 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtPairChain || find[0].level != 8 || find[0].length != 6 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtPairChain || find[1].level != 9 || find[1].length != 6 {
		t.Fatalf("%v", find)
	}
}

func Test_tipThreeChain(t *testing.T) {

	xx := NewSetCard([]byte{0x05, 0x06, 0x07, 0x08, 0x15, 0x16, 0x17, 0x18, 0x25, 0x26, 0x27, 0x28})

	find := xx.Tips(NewSetCard([]byte{0x04, 0x05, 0x14, 0x15, 0x24, 0x25}))
	if len(find) != 3 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThreeChain || find[0].level != 4 || find[0].length != 2 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtThreeChain || find[1].level != 5 || find[1].length != 2 {
		t.Fatalf("%v", find)
	}
	if find[2].ct != pbgame_ddz.CardType_CtThreeChain || find[2].level != 6 || find[2].length != 2 {
		t.Fatalf("%v", find)
	}
}

func Test_tipThreeSoloChain(t *testing.T) {
	xx := NewSetCard([]byte{0x05, 0x15, 0x25, 0x06, 0x16, 0x26, 0x08, 0x07, 0x17, 0x27})
	find := xx.Tips(NewSetCard([]byte{0x05, 0x15, 0x25, 0x04, 0x14, 0x24, 0x06, 0x07}))

	if len(find) != 2 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThreeSoloChain || find[0].level != 4 || find[0].length != 2 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtThreeSoloChain || find[1].level != 5 || find[1].length != 2 {
		t.Fatalf("%v", find)
	}
}

func Test_tipThreePairChain(t *testing.T) {
	xx := NewSetCard([]byte{0x05, 0x15, 0x25, 0x06, 0x16, 0x26, 0x07, 0x17, 0x27, 0x08, 0x18})
	find := xx.Tips(NewSetCard([]byte{0x05, 0x15, 0x25, 0x04, 0x14, 0x24, 0x06, 0x16, 0x07, 0x17}))

	if len(find) != 2 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtThreePairChain || find[0].level != 4 || find[0].length != 2 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtThreePairChain || find[1].level != 5 || find[1].length != 2 {
		t.Fatalf("%v", find)
	}
}

func Test_tipFour2(t *testing.T) {
	xx := NewSetCard([]byte{0x05, 0x15, 0x25, 0x35, 0x06, 0x07, 0x17, 0x27, 0x37})
	find := xx.Tips(NewSetCard([]byte{0x04, 0x14, 0x24, 0x34, 0x05, 0x15}))

	if len(find) != 2 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtFour2 || find[0].level != 3 {
		t.Fatalf("%v", find)
	}
	if find[1].ct != pbgame_ddz.CardType_CtFour2 || find[1].level != 5 {
		t.Fatalf("%v", find)
	}
}

func Test_tipFour4(t *testing.T) {
	xx := NewSetCard([]byte{0x05, 0x15, 0x25, 0x16, 0x06, 0x07, 0x17, 0x27, 0x37})
	find := xx.Tips(NewSetCard([]byte{0x04, 0x14, 0x24, 0x34, 0x05, 0x15, 0x06, 0x16}))

	if len(find) != 1 {
		t.Fatalf("%v", find)
	}
	if find[0].ct != pbgame_ddz.CardType_CtFour4 || find[0].level != 5 {
		t.Fatalf("%v", find)
	}
}
