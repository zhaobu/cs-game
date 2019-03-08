package game

import (
	"strings"

	"github.com/fwhappy/util"
	"mahjong.go/mi/card"
	"mahjong.go/mi/ting"
)

// MahjongUser 麻将用户信息
type MahjongUser struct {
	// 用户id
	UserId int
	// 用户客户端版本
	Version string
	// 房间内的位置, 0 ~ 3
	Index int
	// 总分，在用户加入房间时，根据房间类型，填充不同的值
	Score int
	// 是否庄家
	IsDealer bool
	// 手牌容器
	HandTileList *card.CMap
	// 弃牌容器
	DiscardTileList *card.DiscardContainer
	// 明牌容器
	ShowCardList *card.ShowCardList
	// 听牌容器
	MTC *ting.MTContainer
	// 杠的类型，用于热炮判断
	KongCode int
	// 杠后抓牌标志（是否处于杠后抓牌状态）
	DrowAfterKongFlag bool
	// 赢的牌（包括胡的牌、自摸的牌）, 用在showhandTile中
	WinTile int
	// 是否已定缺，这个字段只是用在断线重连时，告知用户自己是否已定缺，无其他用途
	LackTile int
	// 用户选择要换的牌
	ExchangeOutTiles []int
	// 用户换到牌表
	ExchangeInTiles []int
	// 用户鸡的状态
	KitchenStatus int // 0 不输不赢(黄牌或者烧鸡) 1:赢鸡 2:输鸡
	// 胡牌类型:0未胡牌;1自模胡;2胡牌;3抢杠胡;4热炮胡;5杠后开花
	WinStatus int
	// 胡牌牌型
	WinType int
	// 点炮类型:0未点炮;1点炮;2:热炮;3:被抢杠
	PaoStatus int
	// 操作时间汇总
	ReplyTimeCnt int
	// 过胡状态
	SkipWin []int
	// 过碰状态
	SkipPong []int
	// 是否原缺
	InitLack bool
	// 局内消息序列
	MSC *MSeqContainer

	// 额外的抓牌好牌率
	DrawEffectExtraRate int

	// 会员等级经验加成
	MemberAddExp int
}

// 初始化麻将用户
func newMahjongUser(userId int, version string) *MahjongUser {
	user := &MahjongUser{}
	user.UserId = userId
	user.Version = version
	user.HandTileList = card.NewCMap()
	user.ShowCardList = card.NewShowCardList()
	user.DiscardTileList = card.NewDiscardContainer()
	user.MTC = ting.NewMTContainer()
	user.DrowAfterKongFlag = false
	user.IsDealer = false
	user.SkipWin = make([]int, 0)
	user.SkipPong = make([]int, 0)
	user.InitLack = false
	user.MSC = NewMSeqContainer()
	user.ExchangeOutTiles = make([]int, 0)
	user.ExchangeInTiles = make([]int, 0)
	return user
}

// 给用户手牌排序
// 排序规则，非定缺的排左边，定缺的排右边，rightTile放最后
func (this *MahjongUser) sortHandTile(rightTile int) []int {
	var leftSlice, rightSlice []int

	// 将手牌转成排序好的数组
	handTileSlice := util.SliceCopy(this.HandTileList.ToSortedSlice())

	// 从手牌中移除掉需要放在最后的牌
	if rightTile > 0 {
		handTileSlice = util.SliceDel(handTileSlice, rightTile)
	}
	// 将手牌按照定缺情况，分成左右两个数组
	for _, tile := range handTileSlice {
		if this.isLackTile(tile) {
			rightSlice = append(rightSlice, tile)
		} else {
			leftSlice = append(leftSlice, tile)
		}
	}
	// 将定缺的牌拼到后面
	if len(rightSlice) > 0 {
		leftSlice = append(leftSlice, rightSlice...)
	}
	// 将最后的牌拼上
	if rightTile > 0 {
		leftSlice = append(leftSlice, rightTile)
	}
	return leftSlice
}

// 判断某张牌是不是缺的牌
func (mu *MahjongUser) isLackTile(tile int) bool {
	return mu.LackTile > 0 && mu.LackTile/10 == tile/10
}

// 统计用户拥有的牌
// 只有设置了满堂鸡，才算上弃牌的
func (mu *MahjongUser) getMergedTileMap(statDiscard bool) map[int]int {
	tiles := make(map[int]int)
	// 手牌
	for tile, cnt := range mu.HandTileList.GetTileMap() {
		tiles[tile] = cnt
	}
	// 明牌
	for _, tile := range mu.ShowCardList.GetAllTiles() {
		tiles[tile]++
	}
	// 弃牌
	if statDiscard {
		for _, tile := range mu.DiscardTileList.GetTiles() {
			tiles[tile]++
		}
	}
	return tiles
}

// 是否有“缺”的牌
func (mu *MahjongUser) hasLackTile() bool {
	if mu.LackTile > 0 {
		for _, tile := range mu.HandTileList.GetUnique() {
			if IsSameTileKind(tile, mu.LackTile) {
				return true
			}
		}
	}
	return false
}

// 获取“缺的牌的数量”
func (mu *MahjongUser) getLackCount(lack int) int {
	var cnt = 0
	if mu.LackTile > 0 {
		for _, tile := range mu.HandTileList.ToSlice() {
			if IsSameTileKind(tile, mu.LackTile) {
				cnt++
			}
		}
	}
	return cnt
}

// 检查并设置用户的原缺状态
func (mu *MahjongUser) checkInitLack() bool {
	var hasBam, hasDot, hasCrak = false, false, false
	for _, tile := range mu.HandTileList.GetUnique() {
		if card.IsBAM(tile) {
			hasBam = true
		} else if card.IsCrak(tile) {
			hasCrak = true
		} else if card.IsDot(tile) {
			hasDot = true
		}
	}
	return !hasBam || !hasDot || !hasCrak
}

// EnableSeq 是否支持带序列的消息
// 2.3.0以上的客户端才支持
func (mu *MahjongUser) EnableSeq() bool {
	return mu.Version == "" || mu.Version == "latest" || strings.Compare(mu.Version, "2.3.0") >= 0
}

// SendOperationPush 给用户发送OperationPush
func (mu *MahjongUser) SendOperationPush(opList []*Operation) {
	SendMessageByUserId(mu.UserId, mu.MSC.AddWOperation(opList).ToImPacket(mu.EnableSeq()))
}

// SendUserOperationPush 给用户发送UserOperation
func (mu *MahjongUser) SendUserOperationPush(uOperation *UserOperation) {
	SendMessageByUserId(mu.UserId, mu.MSC.AddUOperation(uOperation).ToImPacket(mu.EnableSeq()))
}

// SendClientOperationPush 给用户发送ClientOperation
func (mu *MahjongUser) SendClientOperationPush(cOperation *ClientOperation) {
	SendMessageByUserId(mu.UserId, mu.MSC.AddCOperation(cOperation).ToImPacket(mu.EnableSeq()))
}
