package api

import (
	"fmt"
	"game/db/mgo"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 请求参数
type PointCardBuyParam struct {
	UserId     uint64 `json:"userid" form:"userid"`         //用户id
	OrderId    string `json:"orderid" form:"orderid"`       //订单号
	DiamondNum uint32 `json:"diamondnum" form:"diamondnum"` //兑换砖石数
	Sign       string `json:"sign"  form:"sign"`            //签名
}

//设置用户红名接口
func PointCardBuyReq(c *gin.Context) {
	req := PointCardBuyParam{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%s%d%s", req.UserId, req.OrderId, req.DiamondNum, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!", Data: nil})
		return
	}
	mgo.AddUserPointcard(req.UserId, req.OrderId, time.Now().Unix(), req.DiamondNum)
	c.JSON(http.StatusOK, apiRsp{Code: Succ, Msg: "成功", Data: nil})
}
