package card

import (
	"game/pb/game/ddz"
	"testing"
)

func Test_have(t *testing.T) {
	xx := NewSetCard([]byte{0x29, 0x39, 0x19})

	b := xx.Have([]byte{0x29, 0x39})
	if !b {
		t.Fatalf("dfdf")
	}
}

func Test_nothave(t *testing.T) {
	xx := NewSetCard([]byte{0x29, 0x39, 0x19})

	b := xx.Have([]byte{0x09})
	if b {
		t.Fatalf("dfdf")
	}
}

func Test_checkJokers(t *testing.T) {
	s := NewSetCard([]byte{0x4e, 0x4f})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtJokers || s.length != 0 || s.level != 0 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkBomb(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x35})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtBomb || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkSolo(t *testing.T) {
	s := NewSetCard([]byte{0x1d})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtSolo || s.length != 0 || s.level != 11 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkPair(t *testing.T) {
	s := NewSetCard([]byte{0x17, 0x37})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtPair || s.length != 0 || s.level != 5 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThree(t *testing.T) {
	s := NewSetCard([]byte{0x17, 0x37, 0x27})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThree || s.length != 0 || s.level != 5 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThreeSolo(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x14})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThreeSolo || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThreePair(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x14, 0x24})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThreePair || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkSoloChain(t *testing.T) {
	s := NewSetCard([]byte{0x04, 0x05, 0x06, 0x07, 0x08})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtSoloChain || s.length != 5 || s.level != 6 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkPairChain(t *testing.T) {
	s := NewSetCard([]byte{0x04, 0x05, 0x06, 0x14, 0x15, 0x16})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtPairChain || s.length != 3 || s.level != 4 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThreeChain(t *testing.T) {
	s := NewSetCard([]byte{0x04, 0x05, 0x14, 0x15, 0x24, 0x25})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThreeChain || s.length != 2 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThreeSoloChain(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x04, 0x14, 0x24, 0x06, 0x07})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThreeSoloChain || s.length != 2 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkThreePairChain(t *testing.T) {
	s := NewSetCard([]byte{0x07, 0x17, 0x27, 0x08, 0x18, 0x28, 0x03, 0x13, 0x04, 0x14})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtThreePairChain || s.length != 2 || s.level != 6 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkFour2(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x35, 0x04, 0x14})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtFour2 || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkFour2_(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x35, 0x03, 0x04})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtFour2 || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}

func Test_checkFour4(t *testing.T) {
	s := NewSetCard([]byte{0x05, 0x15, 0x25, 0x35, 0x03, 0x13, 0x04, 0x14})
	s.Type()
	if s.ct != pbgame_ddz.CardType_CtFour4 || s.length != 0 || s.level != 3 {
		t.Errorf("%v %v %v", s.ct, s.length, s.level)
	}
}
