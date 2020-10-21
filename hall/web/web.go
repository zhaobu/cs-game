package web

// RspCode RspCode
type RspCode int

const (
	Succ RspCode = iota
	Failed
	ArgInvalid
	NotFound
)

// UpdateWealthReq 更新财富
type UpdateWealthReq struct {
	UserID uint64 `json:"uid" form:"uid"` // 用户ID
	Typ    uint32 `json:"type" form:"type"`
	Change int64  `json:"change" form:"change"`
	Event  int    `json:"event"  form:"event"`
}

// BindAgentReq 绑定代理
type BindAgentReq struct {
	UserID uint64 `json:"uid" form:"uid"`
	Agent  string `json:"agent" form:"agent"`
}
