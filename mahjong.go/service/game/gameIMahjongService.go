package game

import (
	"mahjong.go/library/core"
	"mahjong.go/mi/suggest"
)

// 麻将类型接口
type MahjongInterface interface {
	// 执行
	run()
	// 完成当前步骤
	next()
	// 获取麻将用户列表
	getUsers() map[int]*MahjongUser
	// 获取麻将用户列表
	getUser(int) *MahjongUser
	// 检测用户操作是否合法，即是否存在于waitqueue中
	checkUserOperation(int, *Operation) *core.Error
	// 用户回应操作
	replyWait(int, *Operation) *WaitInfo
	// 设置用户能进行的操作
	setWait(int, *WaitInfo)
	// 读取用户能进行的操作
	getWait(int) *WaitInfo
	// 获取重连时的可进行操作列表
	getRestoreOpreationlist(int) []*Operation
	// 清空操作队列
	cleanWait()
	// 是否已全部回应
	isAllReply() bool
	// 根据运算优先级，计算此次应该执行什么操作
	getReplyResult() *WaitMap
	// 计算用户能进行的操作
	calcOperation(int, int, int) []*Operation
	// 计算其他用户能进行什么操作
	calcAfterUserOperation(int) bool
	// 换牌
	userOperationExchange(int, []int) (bool, *core.Error)
	// 定缺
	userOperationLack(int, int) bool
	// 出牌
	userOperationPlay(int, int) *core.Error
	// 吃
	userOperationChow(int, []int) (bool, *core.Error)
	// 碰
	userOperationPong(int) *core.Error
	// 碰之后的操作
	userOperationAfterPong(int)
	// 明杠
	userOperationKong(userId int) *core.Error
	// 暗杠
	userOperationKongDark(userId int, pai int) *core.Error
	// 转弯杠或憨包杠
	userOperationKongTurn(userId int, tile int, opCode int) *core.Error
	// 报听
	userOperationBaoTing(userId, tile int) *core.Error
	// 胡
	userOperationWin(int, opCode int) *core.Error
	// 抢杠胡
	userOperationWinAfterKongTurn(int) *core.Error
	// 自摸
	userOperationWinSelf(userId int, opCode int) *core.Error
	// 过
	userOperationPass(int) int
	// pass cancel
	userOperationPassCancel(int) int
	// 过了之后补花
	userOperationFlowerExchange(int) int
	// 用户操作回滚
	userOperationRollback()
	// 获取 LastOperation
	getLastOperation() *Operation
	// 获取最后操作者
	getLastOperator() int
	// 获取前后鸡
	getChikenFB() []int
	// 获取前后鸡牌面
	getChikenFBTile() int
	// 获取前后鸡的位置
	getChikenFBIndex() int
	// 获取冲锋幺鸡用户Id
	getChikenChargeBam1() int
	// 获取冲锋乌骨鸡用户Id
	getChikenChargeDot8() int
	// 获取责任鸡用户Id
	getChikenResponsibility() int
	// 获取翻牌鸡的牌
	getChikenDrawTile() int
	// 读取骰子数
	getDiceList() [2]int
	// 读取用户此局的输赢积分
	getRoundScore(int) int
	// 读取麻将定缺信息
	getLackList() map[int]int
	// 获取牌墙剩余张数
	getWallTileCount() int
	// 获取最后打牌者id
	getLastPlayerId() int
	// 获取牌局开始时间
	getRoundCreateTime() int64
	// 获取回应操作初始化时间
	getReplyInitTime() int64
	// 获取结算信息
	getFrontData() map[int]*FrontScoreInfo
	// 获取胡信息
	getHInfo() HuInfo
	// 获取从前面抓了多少张
	getForward() int
	// 获取从后面抓了多少张
	getBackward() int
	// 是否金鸡
	isGoldBam1() bool
	// 获取回放操作列表
	getPlaybackOperationList() []*playbackOperation
	// 保存回放
	savePlaybackIntact()
	// 获取未回应操作的用户列表
	getUnReplyWaitUsers() []int
	// 获取已定缺的用户列表
	getLackedUsers() []int
	// 获取庄家id、连庄数
	getDealer() (int, int)
	// 获取选牌器
	getSelector() *suggest.MSelector
	// 是否处于换牌阶段
	isExchanging() bool
}
