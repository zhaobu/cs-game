package main

import (
	"cy/game/db/mgo"
	"cy/game/hall/web"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego/httplib"
	"github.com/gin-gonic/gin"
)

type apiRsp struct {
	Code web.RspCode `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// @Summary 查询用户信息
// @Description
// @Accept  json
// @Produce json
// @Param   userid     path    int     true        "用户ID"
// @Success 0 {object} pbcommon.UserInfo  "用户信息"
// @Failure 2 {object} web.RspCode "用户ID无效"
// @Failure 3 {object} web.RspCode "更新失败"
// @Router /userinfo/{userid} [get]
func userinfo(c *gin.Context) {
	uidStr := c.Param("id")
	if uidStr == "" {
		return
	}

	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: web.ArgInvalid, Msg: err.Error()})
		return
	}

	info, err := mgo.QueryUserInfo(uid)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: web.NotFound, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiRsp{Code: web.Succ, Data: info})
}

// @Summary 更新用户财富
// @Description
// @Accept  json
// @Produce json
// @Param   request       body    web.UpdateWealthReq     true        "type:1金币 2砖石 event:事件类型 暂时未定义"
// @Success 0 {object} pbcommon.UserInfo  "用户信息"
// @Failure 1 {object} web.RspCode "更新失败"
// @Router /updatewealth [post]
func updateWealth(c *gin.Context) {
	req := web.UpdateWealthReq{}
	c.ShouldBindJSON(&req)

	info, err := mgo.UpdateWealth(req.UserID, req.Typ, req.Change)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: web.Failed, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiRsp{Code: web.Succ, Data: info})
	// TODO 通知游戏服务
	// TODO 记录事件
}

// @Summary 绑定代理
// @Description
// @Accept  json
// @Produce json
// @Param   request      body    web.BindAgentReq     true        "agent: 代理ID"
// @Success 0 {object} pbcommon.UserInfo  "用户信息"
// @Failure 1 {object} web.RspCode "绑定失败"
// @Router /bindagent [post]
func bindAgent(c *gin.Context) {
	req := web.BindAgentReq{}
	c.ShouldBindJSON(&req)

	info, err := mgo.UpdateAgentID(req.UserID, req.Agent)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: web.Failed, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiRsp{Code: web.Succ, Data: info})
	// TODO 通知游戏服务
	// TODO 记录事件
}

// @Summary 查询游戏列表
// @Description
// @Accept  json
// @Produce json
// @Success 0 {string} string  "成功"
// @Router /gamelist [get]
func gameList(c *gin.Context) {
	glist, err := queryGameList()
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: web.Failed, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiRsp{Code: web.Succ, Data: glist})
}

func queryGameList() (gamelist []string, err error) {
	// "http://192.168.0.90:8500/v1/kv/cy_game/game"
	url := fmt.Sprintf("http://%s/v1/kv%s/game", *consulAddr, *basePath)

	req := httplib.Get(url)
	req.Param("recurse", "true")
	req.Param("keys", "")
	body, err := req.String()
	if err != nil {
		return nil, err
	}

	var jsonB []string
	err = json.Unmarshal([]byte(body), &jsonB)
	if err != nil {
		return nil, err
	}

	for _, v := range jsonB {
		ss := strings.Split(v, "/")
		if len(ss) == 3 {
			gamelist = append(gamelist, ss[2])
		}
	}
	return
}
