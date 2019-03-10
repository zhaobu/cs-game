package suggest

import (
	"sort"

	"github.com/fwhappy/util"
	"mahjong.go/mi/card"
	"mahjong.go/mi/step"
	"mahjong.go/mi/ting"
	"mahjong.go/mi/weight"
)

// GetSuggest 根据当前牌型，推荐一张牌
// 通用规则：先打缺，跟ai等级无关
func (ms *MSelector) GetSuggest() (int, []int) {
	// 通用规则
	// 先打缺
	// 优先推荐定缺的牌
	if ms.lack > 0 {
		// 用户手牌
		handTiles := util.MapToSlice(ms.handTiles)
		// 按1~9的顺序推荐
		sort.Ints(handTiles)
		for _, tile := range handTiles {
			if card.IsSameType(tile, ms.lack) {
				return tile, []int{}
			}
		}
	}

	// 如果当前牌未叫牌，则根据选牌算法计算
	switch ms.aiLevel {
	case AI_BRASS: // 英勇黄铜
		return ms.suggestByAIBrass()
	case AI_SLIVER: // 不屈白银
		return ms.suggestByAISliver()
	case AI_GOLD: // 荣耀黄金
		fallthrough
		// return ms.suggestByAIGold(s)
	case AI_PLATINUM: // 华贵铂金
		fallthrough
		// return ms.suggestByAIPlatinum(s)
	case AI_DIAMOND: // 璀璨钻石
		fallthrough
		// return ms.suggestByAIDiamond(s)
	case AI_MASTER: // 非凡大师
		fallthrough
		// return ms.suggestByAIMaster(s)
	case AI_KING: // 最强王者
		fallthrough
		// return ms.suggestByAIKing(s)
	default:
		// return ms.suggestByAIPlatinum(handTiles)
		return ms.GetSuggestTile()
	}

	/*

		// 明牌
		showTiles := ms.ShowShowTiles()

		// 如果有定缺的牌，则推荐缺的牌
		// 因为大的牌在右手边，所以优先推荐，便于用户操作
		if ms.lack > 0 {
			for i := len(handTiles) - 1; i >= 0; i-- {
				if card.IsSameType(handTiles[i], ms.lack) {
					return handTiles[i]
				}
			}
		}

		// 如果当前牌型已经叫牌，推荐剩余张数最多的牌
		if tingMap := ting.GetTingMap(handTiles, showTiles); len(tingMap) > 0 {
			selectTile := 0
			selectTileTingCnt := -1
			for tile, tingTiles := range tingMap {
				tingCnt := ms.getRemainTilesCnt(tingTiles)
				if tingCnt > selectTileTingCnt {
					selectTileTingCnt = tingCnt
					selectTile = tile
				}
			}
			return selectTile
		}

	*/
}

// GetSuggest 根据当前牌型，推荐一张牌
// 通用规则：先打缺，跟ai等级无关
func (ms *MSelector) GetSuggestOld() (int, []int) {
	// 通用规则
	// 先打缺
	// 优先推荐定缺的牌
	if ms.lack > 0 {
		// 用户手牌
		handTiles := util.MapToSlice(ms.handTiles)
		// 按1~9的顺序推荐
		sort.Ints(handTiles)
		for _, tile := range handTiles {
			if card.IsSameType(tile, ms.lack) {
				return tile, []int{}
			}
		}
	}

	// 如果当前牌未叫牌，则根据选牌算法计算
	switch ms.aiLevel {
	case AI_BRASS: // 英勇黄铜
		return ms.suggestByAIBrass()
	case AI_SLIVER: // 不屈白银
		return ms.suggestByAISliver()
	case AI_GOLD: // 荣耀黄金
		fallthrough
		// return ms.suggestByAIGold(s)
	case AI_PLATINUM: // 华贵铂金
		fallthrough
		// return ms.suggestByAIPlatinum(s)
	case AI_DIAMOND: // 璀璨钻石
		fallthrough
		// return ms.suggestByAIDiamond(s)
	case AI_MASTER: // 非凡大师
		fallthrough
		// return ms.suggestByAIMaster(s)
	case AI_KING: // 最强王者
		fallthrough
		// return ms.suggestByAIKing(s)
	default:
		// return ms.suggestByAIPlatinum(handTiles)
		return ms.GetSuggestTile3()
	}

	/*

		// 明牌
		showTiles := ms.ShowShowTiles()

		// 如果有定缺的牌，则推荐缺的牌
		// 因为大的牌在右手边，所以优先推荐，便于用户操作
		if ms.lack > 0 {
			for i := len(handTiles) - 1; i >= 0; i-- {
				if card.IsSameType(handTiles[i], ms.lack) {
					return handTiles[i]
				}
			}
		}

		// 如果当前牌型已经叫牌，推荐剩余张数最多的牌
		if tingMap := ting.GetTingMap(handTiles, showTiles); len(tingMap) > 0 {
			selectTile := 0
			selectTileTingCnt := -1
			for tile, tingTiles := range tingMap {
				tingCnt := ms.getRemainTilesCnt(tingTiles)
				if tingCnt > selectTileTingCnt {
					selectTileTingCnt = tingCnt
					selectTile = tile
				}
			}
			return selectTile
		}

	*/
}

// 英勇黄铜
// 根据权重推荐
func (ms *MSelector) suggestByAIBrass() (int, []int) {
	handTiles := util.MapToSlice(ms.handTiles)
	sort.Ints(handTiles)
	// 根据权重推荐
	tilesWeight := weight.GetCardsWeight(handTiles, nil)
	suggestTile := ms.suggestByWeightAndRemain(tilesWeight)

	return suggestTile, []int{}
}

// 不屈白银
func (ms *MSelector) suggestByAISliver() (int, []int) {
	// 清空明牌和弃牌，计算剩余的牌
	ms.SetShowTilesMap(map[int]int{})
	ms.SetDiscardTilesMap(map[int]int{})
	ms.CalcRemaimTiles()

	return ms.GetSuggestTile()
}

// 华贵铂金
func (ms *MSelector) suggestByAIPlatinum(s []int) int {
	// fixme 暂时没空实现，还是根据权重选择
	tilesWeight := weight.GetCardsWeight(s, nil)
	return ms.suggestByWeightAndRemain(tilesWeight)

	/*
		// 计算当前手牌所处的阶段
		tilesStep := step.GetCardsStep(s)

		if tilesStep < 3 {
			tilesWeight := weight.GetCardsWeight(s, nil)
			return ms.suggestByWeightAndRemain(tilesWeight)
		}


			// 最多一类有效牌张数
			maxEffectTileCnt := 0
			// 最多一类有效牌列表
			maxEffectTiles := []int{}
			maxEffectList := map[int][]int{}
			maxEffectTotalWeights := map[int]int{}

			// 循环删除某一张手牌，计算一类有效牌的数量
			for _, playTile := range util.SliceUniqueInt(s) {
				tiles := util.SliceDel(s, playTile)
				// 计算删除后的牌阶，如果小于当前牌阶，跳过计算
				currentStep := step.GetCardsStep(tiles)
				if currentStep < tilesStep {
					continue
				}
				effects, weight := calcEffectsAndRemainWeight(tiles, currentStep)
				effectsLen := len(effects)
				if effectsLen > maxEffectTileCnt {
					maxEffectTileCnt = effectsLen
					maxEffectTiles = []int{playTile}
					maxEffectList = map[int][]int{}
					maxEffectList[playTile] = effects
					maxEffectTotalWeights = map[int]int{}
					maxEffectTotalWeights[playTile] = GetCardsWeightSum(effects)
				}

				step, effects, totalWeight := calcEffectsAndRemainWeight(tiles)
				if step >= unPlayStep {
					sort.Ints(effects)
					showDebug("打出:%v,手牌:%v, 一类有效牌:%v(%v)(remain:%v)------------", playTile, tiles, effects, len(effects), ms.getRemainTilesCnt(effects))

					effectsLen := len(effects)
					if effectsLen > maxEffectTileCnt {
						maxEffectTileCnt = effectsLen
						maxEffectTiles = []int{playTile}
						maxEffectList = map[int][]int{}
						maxEffectList[playTile] = effects
						maxEffectTotalWeights = map[int]int{}
						maxEffectTotalWeights[playTile] = totalWeight
					} else if len(effects) == maxEffectTileCnt {
						maxEffectTiles = append(maxEffectTiles, playTile)
						maxEffectList[playTile] = effects
						maxEffectTotalWeights[playTile] = totalWeight
					}
				}
			}

			// 如果存在相同的一类有效牌，则根据权重再取一次
			showDebug("maxEffectTiles:%v", maxEffectTiles)
			if len(maxEffectTiles) > 1 {
				showDebug("存在多张有效牌相同的打法，根据剩余牌权重重新筛选一次:%v", maxEffectTotalWeights)

				// 读取权重最大的牌
				maxWeightTiles, _ := getMaxValueSlice(maxEffectTotalWeights)
				showDebug("权重最大的牌:%v", maxWeightTiles)

				// 找出权重最小的牌中，关联牌最少的一张
				maxRemainCnt := -1
				var maxRemainTile int
				for _, tile := range maxWeightTiles {
					remainCnt := ms.getRemainTilesCnt([]int{tile})
					if maxRemainCnt < remainCnt {
						maxRemainCnt = remainCnt
						maxRemainTile = tile
					}
				}
				return maxRemainTile
			}

			return maxEffectTiles[0]
	*/
}

// 根据权重筛选
func (ms *MSelector) suggestByWeightAndRemain(tilesWeight map[int]int) int {
	// 读取权重最小的牌
	_, minWeightTiles := util.GetMapMinValue(tilesWeight)

	// 找出权重最小的牌中，关联牌最少的一张
	minRelationCnt := 1000000
	minRelationTile := 0
	for _, tile := range minWeightTiles {
		relationCnt := ms.getRemainTilesCnt(card.GetRelationTiles(tile))
		if relationCnt < minRelationCnt {
			minRelationCnt = relationCnt
			minRelationTile = tile
		}
	}
	return minRelationTile
}

// GetSuggestMap 获取推荐的列表
// 这里默认外面已经考虑过叫牌了，不再考虑叫牌的情况
// 有缺的时候，不给任何推荐，用户必须先打缺
// 如果有孤张，优先返回孤张
// 孤一对，不拆做提示
// 最多返回maxLength个结果，如果maxLength=0，表示返回所有
// maxLength: 最多返回多少个孤张
func (ms *MSelector) GetSuggestMap(maxLength int) map[int][]int {
	suggestMap := make(map[int][]int)
	// 优先判断是否有缺
	if ms.hasLack() {
		return suggestMap
	}
	// 用户手牌
	handTiles := util.MapToSlice(ms.handTiles)
	// 如果有孤张，则优先提示孤张
	// 孤张按权重来选，优先边张
	if gTiles := ms.getGuTiles(); len(gTiles) > 0 {
		for _, v := range weight.GetMinWeigthTiles(handTiles, gTiles, maxLength) {
			suggestMap[v] = []int{}
		}
		return suggestMap
	}
	// 按一类有效牌剩余数推荐
	// 获取所有的孤一对
	gpTiles := ms.getGuPairTiles()
	// 计算手牌当前牌阶
	currentStep := step.GetCardsStep(handTiles)

	// 循环删除一张手牌后，计算一类有效牌的数量
	for playTile := range ms.handTiles {
		// 孤对不推荐
		if util.IntInSlice(playTile, gpTiles) {
			continue
		}
		tiles := util.SliceDel(handTiles, playTile)
		// 如果打出后，牌阶比之前的还要低，肯定不能这么打
		playedStep := step.GetCardsStep(tiles)
		if playedStep < currentStep {
			continue
		}
		// 计算一类有效牌数量
		effects := calcEffects(tiles, currentStep)
		if len(effects) > 0 {
			suggestMap[playTile] = effects
		}
	}

	// 如果未找到一类有效牌，则继续查找二类有效牌
	if len(suggestMap) == 0 {
		// 循环删除一张手牌后，计算一类有效牌的数量
		for playTile := range ms.handTiles {
			sEffects := []int{}
			// 孤对不推荐
			if util.IntInSlice(playTile, gpTiles) {
				continue
			}
			tiles := util.SliceDel(handTiles, playTile)
			// 如果打出后，牌阶比之前的还要低，肯定不能这么打
			playedStep := step.GetCardsStep(tiles)
			if playedStep < currentStep {
				continue
			}

			// 获取跟playTile有关系的牌
			for _, rTile := range card.GetRelationTiles(playTile) {
				effects := calcEffects(append([]int{rTile}, tiles...), currentStep)
				if len(effects) > 0 {
					sEffects = append(sEffects, rTile)
				}
			}
			if len(sEffects) > 0 {
				suggestMap[playTile] = sEffects
			}
		}
	}

	// 如果找到了推荐的牌，则找出有效进张数最多的那些
	remainTiles := []int{}
	maxCnt := 0
	for tile, effects := range suggestMap {
		remainCnt := ms.getRemainTilesCnt(effects)
		if remainCnt > maxCnt {
			remainTiles = []int{tile}
			maxCnt = remainCnt
		} else if remainCnt == maxCnt {
			remainTiles = append(remainTiles, tile)
		}
	}
	if len(remainTiles) > maxLength {
		remainTiles = weight.GetMinWeigthTiles(handTiles, remainTiles, maxLength)
	}
	for k := range suggestMap {
		if !util.IntInSlice(k, remainTiles) {
			delete(suggestMap, k)
		}
	}

	return suggestMap
}

// GetSuggestTile 获取推荐的排
// 有缺的时候，先推荐定缺的牌，从1~9
func (ms *MSelector) GetSuggestTile() (int, []int) {
	// 用户手牌
	handTiles := util.MapToSlice(ms.handTiles)
	sort.Ints(handTiles)

	var suggestTiles []int

	_, tileEffects, secondEffects := ms.GetSuggestProgress()

	tileScores := map[int]int{}
	for playTile, effects := range tileEffects {
		// 获取一类有校牌张数
		var firstTileCnt int
		if len(effects) > 0 {
			firstTileCnt = ms.getRemainTilesCnt(effects)
		}

		// 获取二类有校牌张数
		var secondTileCnt int
		if secondEffects, ok := secondEffects[playTile]; ok && len(secondEffects) > 0 {
			secondTileCnt = ms.getRemainTilesCnt(secondEffects)
		}

		// 算分
		tileScores[playTile] = firstTileCnt*1000 + secondTileCnt
	}

	// 获取最高分有哪些牌
	_, suggestTiles = util.GetMapMaxValue(tileScores)

	// 有多少类牌
	uniqueTiles := util.SliceUniqueInt(handTiles)

	// 先找孤张
	// 孤张中有边张先出边张
	guTiles := ms.getSpecifiedGuTiles(suggestTiles)
	if len(guTiles) > 0 {
		tilesWeight := weight.GetCardsWeight(uniqueTiles, guTiles)
		_, minWeightTiles := util.GetMapMinValue(tilesWeight)
		sort.Ints(minWeightTiles)
		tile := minWeightTiles[0]
		return tile, tileEffects[tile]
	}

	// 继续找剩余牌中的边张
	tilesWeight := weight.GetCardsWeight(uniqueTiles, suggestTiles)
	_, minWeightTiles := util.GetMapMinValue(tilesWeight)
	sort.Ints(minWeightTiles)
	tile := minWeightTiles[0]
	if card.IsSide(tile) || card.IsSideNeighbor(tile) {
		return tile, tileEffects[tile]
	}

	// 没有孤张，继续找吊张
	// 吊张中有边张先出边张
	diaoTiles := ms.getSpecifiedDiaoTiles(suggestTiles)
	if len(diaoTiles) > 0 {
		tilesWeight := weight.GetCardsWeight(uniqueTiles, diaoTiles)
		_, minWeightTiles := util.GetMapMinValue(tilesWeight)
		sort.Ints(minWeightTiles)
		tile := minWeightTiles[0]
		return tile, tileEffects[tile]
	}

	// 什么都没找到，找一张最小的
	sort.Ints(suggestTiles)
	tile = suggestTiles[0]
	return tile, tileEffects[tile]
}

// GetSuggestTile3 获取推荐的排
// 有缺的时候，先推荐定缺的牌，从1~9
func (ms *MSelector) GetSuggestTile3() (int, []int) {
	// 用户手牌
	handTiles := util.MapToSlice(ms.handTiles)
	sort.Ints(handTiles)

	// 有效牌map
	tileEffects := make(map[int][]int, len(ms.handTiles))

	// 计算手牌当前牌阶
	currentStep := step.GetCardsStep(handTiles)

	// 有效牌列表 key: 打出的牌; value: 实际有效进张数
	effectsMap := make(map[int]int)
	// 循环删除一张手牌后，计算一类有效牌的数量
	uniqueTiles := util.SliceUniqueInt(handTiles)
	// 会导致降阶的牌
	deStepTiles := []int{}
	sort.Ints(uniqueTiles)
	for _, playTile := range uniqueTiles {
		tiles := util.SliceDel(handTiles, playTile)
		// 计算会导致降阶的牌
		deStep := step.GetCardsStep(tiles)
		if deStep < currentStep {
			deStepTiles = append(deStepTiles, playTile)
		}
		// 计算一类有效牌数量
		effects := calcEffects(tiles, currentStep)
		// 记录有多少有效牌
		tileEffects[playTile] = effects
		if len(effects) > 0 {
			// 计算实际有效进张数
			effectsMap[playTile] = ms.getRemainTilesCnt(effects)
		}
	}
	// 如果找到了一类有效牌，需要实际有效进张数最多的那张牌
	var suggestTiles []int
	if len(effectsMap) > 0 {
		_, suggestTiles = util.GetMapMaxValue(effectsMap)
	} else {
		suggestTiles = util.SliceUniqueInt(handTiles)
		if len(suggestTiles) != len(deStepTiles) {
			suggestTiles = util.SliceDel(suggestTiles, deStepTiles...)
		}
	}

	// 先找孤张
	// 孤张中有边张先出边张
	guTiles := ms.getSpecifiedGuTiles(suggestTiles)
	if len(guTiles) > 0 {
		tilesWeight := weight.GetCardsWeight(uniqueTiles, guTiles)
		_, minWeightTiles := util.GetMapMinValue(tilesWeight)
		sort.Ints(minWeightTiles)
		tile := minWeightTiles[0]
		return tile, tileEffects[tile]
	}

	// 继续找剩余牌中的边张
	tilesWeight := weight.GetCardsWeight(uniqueTiles, suggestTiles)
	_, minWeightTiles := util.GetMapMinValue(tilesWeight)
	sort.Ints(minWeightTiles)
	tile := minWeightTiles[0]
	if card.IsSide(tile) || card.IsSideNeighbor(tile) {
		return tile, tileEffects[tile]
	}

	// 没有孤张，继续找吊张
	// 吊张中有边张先出边张
	diaoTiles := ms.getSpecifiedDiaoTiles(suggestTiles)
	if len(diaoTiles) > 0 {
		tilesWeight := weight.GetCardsWeight(uniqueTiles, diaoTiles)
		_, minWeightTiles := util.GetMapMinValue(tilesWeight)
		sort.Ints(minWeightTiles)
		tile := minWeightTiles[0]
		return tile, tileEffects[tile]
	}

	// 什么都没找到，找一张最小的
	sort.Ints(suggestTiles)
	tile = suggestTiles[0]
	return tile, tileEffects[tile]
}

func (ms *MSelector) GetSuggestProgress() (suggestTile int, firstEffectsMap, secondEffectsMap map[int][]int) {
	firstEffectsMap = make(map[int][]int)
	firstEffectsRemainMap := make(map[int]int)
	firstEffectsRemainMax := 0

	secondEffectsMap = make(map[int][]int)

	// 用户手牌
	handTiles := util.MapToSlice(ms.handTiles)
	handTilesLen := len(handTiles)
	// 计算手牌当前牌阶
	currentStep := step.GetCardsStep(handTiles)

	// 循环删除一张手牌后，计算一类有效牌的数量
	uniqueTiles := util.SliceUniqueInt(handTiles)
	sort.Ints(uniqueTiles)
	for _, playTile := range uniqueTiles {
		firstEffectsMap[playTile] = []int{}

		// 剩余手牌
		tiles := util.SliceDel(handTiles, playTile)

		// 计算打出后的牌阶
		playedStep := step.GetCardsStep(tiles)
		// 不能打会降阶的牌
		if playedStep < currentStep {
			continue
		}

		// 计算一类有效牌
		firstEffects := calcEffects(tiles, playedStep)
		firstEffectsMap[playTile] = firstEffects

		// 计算一类有效进张数和最大的一类有效进张数
		firstEffectsRemainMap[playTile] = ms.getRemainTilesCnt(firstEffects)
		if firstEffectsRemainMap[playTile] > firstEffectsRemainMax {
			firstEffectsRemainMax = firstEffectsRemainMap[playTile]
		}
	}

	// 4.5阶：4阶已叫牌
	// 3阶的时候，遍历所有的一类有效牌，找出有没有可能有4.5阶存在

	mayTingStep := 0
	switch handTilesLen {
	case 14:
		mayTingStep = 3
	case 11:
		mayTingStep = 2
	case 8:
		mayTingStep = 1
	}

	if currentStep == mayTingStep {
		canTing := false
		canTingMap := make(map[int][]int)
		canTingRemainMap := make(map[int]int)
		canTingRemainMax := 0

		for playTile, firstEffects := range firstEffectsMap {
			canTingMap[playTile] = []int{}
			if len(firstEffects) == 0 {
				continue
			}
			// 剩余手牌
			tiles := util.SliceDel(handTiles, playTile)
			// 计算一类有校牌中，有哪些一类有效牌可导致听牌
			canTingTiles := calcCanTingTiles(tiles, firstEffects)
			if len(canTingTiles) > 0 {
				canTing = true
				canTingMap[playTile] = canTingTiles
				canTingRemainMap[playTile] = ms.getRemainTilesCnt(canTingTiles)
				if canTingRemainMap[playTile] > canTingRemainMax {
					canTingRemainMax = canTingRemainMap[playTile]
				}
			}
		}

		// 如果有可能导致听牌的一类有校牌，则4.5阶的才算一类有效牌
		if canTing {
			firstEffectsMap = canTingMap
			firstEffectsRemainMap = canTingRemainMap
			firstEffectsRemainMax = canTingRemainMax
		}

	}

	// 找到所有相同一类有效进张数的牌，继续去比二类有效进张数
	for _, playTile := range uniqueTiles {
		secondEffectsMap[playTile] = []int{}

		// 跳过一类剩余有效进张数小的
		if firstEffectsRemainMap[playTile] != firstEffectsRemainMax {
			continue
		}

		// 剩余手牌
		tiles := util.SliceDel(handTiles, playTile)
		secondEffects := []int{}
		for _, replaceTile := range util.SliceUniqueInt(tiles) {
			remainTiles := util.SliceDel(tiles, replaceTile)
			for _, relationTile := range card.GetRelationTiles(remainTiles...) {
				// 跳过一类有效牌
				if util.IntInSlice(relationTile, firstEffectsMap[playTile]) {
					continue
				}

				// 跳过已经是二类的了，不再重复计算
				if util.IntInSlice(relationTile, secondEffects) {
					continue
				}

				// 补进来，手牌变成13张
				appendTiles := append(remainTiles, relationTile)

				// 如果降阶了，不能算二类
				playedStep := step.GetCardsStep(appendTiles)
				// 不能打会降阶的牌
				if playedStep < currentStep {
					continue
				}

				effects := calcEffects(appendTiles, playedStep)
				remainCnt := ms.GetRemainTilesCnt(effects)
				// 剩余张数，超过一类有效进张的，算二类有效牌
				if remainCnt > firstEffectsRemainMap[playTile] {
					secondEffects = append(secondEffects, relationTile)
				}
			}
		}
		secondEffectsMap[playTile] = secondEffects
	}

	return
}

// GetSuggestLack 计算推荐定缺
// 我们原来是根据哪门派少定义哪一门，但实际这个定缺是有缺陷的；
// 根据计算三门牌的得分，最后定缺得分最少的一门牌；
// 评估得分 = 牌数得分 + 牌型得分
// 牌数得分 = 每张牌3分，有几张牌加计分
// 牌型得分 = 杠10分>刻子7分>顺子3分>对子3分>挨张2分
// 牌型得分中一张牌只能计算一次；
func (ms *MSelector) GetSuggestLack() int {
	// 已定缺
	if ms.lack > 0 {
		return ms.lack
	}

	// 统计手牌中，万条筒牌的数量
	craks, bams, dots := card.KindCards(ms.ShowHandTiles()...)
	// fmt.Printf("craks:%v\n", craks)
	// fmt.Printf("bams:%v\n", bams)
	// fmt.Printf("dots:%v\n", dots)
	// 计算万、条、筒的分数
	craksScore := weight.GetCardsScore(craks)
	bamsScore := weight.GetCardsScore(bams)
	dotsScore := weight.GetCardsScore(dots)

	// fmt.Printf("cardsScore:%v\n", craksScore)
	// fmt.Printf("bamsScore:%v\n", bamsScore)
	// fmt.Printf("dotsScore:%v\n", dotsScore)

	if craksScore <= bamsScore && craksScore <= dotsScore {
		return card.MAHJONG_CRAK1
	} else if bamsScore < craksScore && bamsScore <= dotsScore {
		return card.MAHJONG_BAM1
	} else {
		return card.MAHJONG_DOT1
	}
}

// GetValidTileCnt 获取当前牌型，可进牌的张数
func (ms *MSelector) GetValidTileCnt(handTiles []int) int {
	var count int
	//手牌及张数
	handTileCnt := len(handTiles)
	// 如果是2、5、8、11、14，需要循环去掉一张后，计算最大的可进张数
	if handTileCnt == 2 || handTileCnt == 5 || handTileCnt == 8 || handTileCnt == 11 || handTileCnt == 14 {
		// 循环删除一张手牌后，计算一类有效牌的数量
		uniqueTiles := util.SliceUniqueInt(handTiles)
		sort.Ints(uniqueTiles)
		for _, playTile := range uniqueTiles {
			tiles := util.SliceDel(handTiles, playTile)
			// 计算一类有效牌数量
			effects := calcEffects(tiles, 0)
			cnt := ms.getRemainTilesCnt(effects)
			if cnt > count {
				count = cnt
			}
		}
	} else {
		effects := calcEffects(handTiles, 0)
		count = ms.getRemainTilesCnt(effects)
	}
	return count
}

// CalcEffects 获取推荐的排
// 有缺的时候，先推荐定缺的牌，从1~9
func (ms *MSelector) CalcEffects() []int {
	// 优先推荐定缺的牌
	// 用户手牌
	baseTiles := []int{}
	if ms.lack > 0 {
		for _, tile := range util.MapToSlice(ms.handTiles) {
			if !card.IsSameType(tile, ms.lack) {
				baseTiles = append(baseTiles, tile)
			}
		}
	} else {
		baseTiles = util.MapToSlice(ms.handTiles)
	}
	sort.Ints(baseTiles)

	return calcEffects(baseTiles, 0)
}

func calcCanTingTiles(tiles, maybeTiles []int) []int {
	canTingTiles := []int{}
	for _, tile := range maybeTiles {
		originTiles := util.SliceCopy(tiles)
		calcTiles := append(originTiles, tile)
		sort.Ints(calcTiles)

		m := ting.GetTingMap(calcTiles, nil)
		if canTing := len(m) > 0; canTing {
			canTingTiles = append(canTingTiles, tile)
		}
	}
	return canTingTiles
}
