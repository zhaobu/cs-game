package api

import (
	"cy/game/db/mgo"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// BindAgentReq 绑定代理
type BindAgentReq struct {
	UserID uint64 `json:"UserId" form:"UserId"`
	Agent  string `json:"Agent" form:"Agent"`
	Sign   string `json:"Sign"  form:"Sign"`
}

//绑定代理
func BuildAgentReq(c *gin.Context) {
	req := BindAgentReq{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%s%s", req.UserID,req.Agent, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!"})
		return
	}
	_, err := mgo.UpdateAgentID(req.UserID, req.Agent)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, apiRsp{Code: Succ,  Msg: "成功", Data: nil })
}