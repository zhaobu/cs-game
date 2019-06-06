package api

import (
	"cy/game/db/mgo"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

//设置用户红名接口
// UpdateWealthReq 更新财富
type SetUserRedName struct {
	UserID    uint64 `json:"UserId" form:"UserId"`
	IsRedName int32  `json:"IsRedName" form:"IsRedName"`
	Sign      string `json:"sign"  form:"sign"`
}

//设置用户红名接口
func SetUserRedNameReq(c *gin.Context){
	req := SetUserRedName{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%d%s", req.UserID,req.IsRedName, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!"})
		return
	}
	_,err := mgo.SetUserRedName(req.UserID,req.IsRedName == 1)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, apiRsp{Code: Succ, Msg: "成功", Data: nil})
}