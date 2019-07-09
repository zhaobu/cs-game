package main

// 信息变动 基本信息、游戏参数、
// 成员变动 数量、身份、
// 桌子变动 数量、属性

type clubChangeTyp int

const (
	clubChangeTypJoin   clubChangeTyp = 1
	clubChangeTypUpdate clubChangeTyp = 2
	clubChangeTypExit   clubChangeTyp = 3
	clubChangeTypRemove clubChangeTyp = 4
	clubChanageTypeDeskUpdata clubChangeTyp  = 5
	// 被踢出？
)

const (
	identityMaster = 1 // 群主
	identityAdmin  = 2 // 管理员
	identityNormal = 3 // 普通成员
	identityBlack  = 4 // 黑名单
	identityInReview = 5 //审核中
)

const (
	emailTypTitle          = 0
	emailTypJoinClub       = 1
	emailTypInviteJoinClub = 2
	emailTypTransferMaster = 3
)
