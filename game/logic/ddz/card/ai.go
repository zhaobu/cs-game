package card

// TimeOutOut 超时出牌
func (sc *SetCard) TimeOutOut(isFreeOut bool, lastGiveCard *SetCard) (outCard *SetCard) {
	if isFreeOut {
		len := sc.Len()
		if len > 0 {
			outSeq := []byte{sc.cards[len-1].Seq}
			outCard = NewSetCard(outSeq)
		}
		return
	}
	find := sc.Tips(lastGiveCard)
	if len(find) == 0 {
		return nil
	}
	return find[0]
}
