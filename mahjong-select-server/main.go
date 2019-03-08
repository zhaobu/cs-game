package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"mahjong-select-server/config"
	"mahjong-select-server/core"
	fbsInfo "mahjong-select-server/fbs/info"
	"mahjong-select-server/ierror"
	"mahjong-select-server/servers"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	flatbuffers "github.com/google/flatbuffers/go"
)

var (
	// 环境，用来读取不同的配置文件
	env = flag.String("env", "local", "env")
	// 监听端口
	port = flag.Int("port", 8081, "port")
)
var leagueUrl string

func init() {
	// 解析url参数
	flag.Parse()

	leagueUrl = "http://0.0.0.0:31213/getUserRace"
}

// Result 选服结果
type Result struct {
	Code       int    `json:"code"`
	Errmsg     string `json:"errmsg"`
	RoomId     int64  `json:"room_id"`
	RoomRemote string `json:"room_remote"`
	RaceId     int64  `json:"race_id"`
	RaceRemote string "json:`race_remote`"
}

func main() {
	defer util.RecoverPanic()

	// 初始化日志配置
	core.LoadLoggerConfig(core.GetConfigFile("log.toml", *env, "conf"))
	defer core.Logger.Flush()
	// 初始化Redis配置
	core.LoadRedisConfig(core.GetConfigFile("redis.toml", *env, "conf"))

	// 初始化服务器信息
	// 加载配置的游戏服务器列表
	servers.InitGameServers()
	// 加载预发布规则
	servers.InitPreviewRule()
	// 加载灰度发布规则
	servers.InitGrayRules()
	// 加载当前活跃的服务器
	servers.InitActiveServers()

	// 测试
	http.HandleFunc("/test", hello)
	// 跨域
	http.HandleFunc("/jsonp", jsonp)
	// 选服服务
	http.HandleFunc("/client/selectServer", selectServer)
	http.HandleFunc("/client/selectServerForLocal", selectServerForLocal)
	http.HandleFunc("/selectIntact", selectIntact)
	http.HandleFunc("/client/getMyServerForLocal", getMyServerForLocal)
	http.HandleFunc("/leagueRoomServer", leagueRoomServer)
	http.HandleFunc("/selectServerRemote", selectServerRemote)
	core.Logger.Info("start listen:%v", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), nil)
	if err != nil {
		core.Logger.Errorf("ListenAndServe:%v", err)
		return
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	sleepTime, _ := strconv.Atoi(r.Form.Get("sleep"))
	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
	w.Write([]byte("Hello world"))
}

func jsonp(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	w.Write([]byte(fmt.Sprintf("%v('ok')", r.Form.Get("callback"))))
}

// 选服服务
func selectServer(w http.ResponseWriter, r *http.Request) {
	core.Logger.Debug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	core.Logger.Debug("收到客户端的选服请求:%v", r.RemoteAddr)
	// 读取命令行参数
	r.ParseForm()
	token := r.Form.Get("token")
	// 解析fbs参数
	buf, _ := ioutil.ReadAll(r.Body)
	core.Logger.Debug("接收到传过来的body参数:%v", buf)
	request := fbsInfo.GetRootAsSelectServerRequest(buf, 0)
	userId := int(request.UserId())
	selectType := string(request.RoomNum())
	core.Logger.Debug("[selectServer]解析参数,userId:%v, selectType:%v, token:%v", userId, selectType, token)

	// 验证参数完整性
	var roomId, raceId int64
	var remote, version string
	var err *ierror.Error
	// 检查参数完整性
	version, err = verify(userId, selectType, token)

	// 选服
	if err == nil {
		roomId, remote, raceId, err = doSelect(userId, selectType, version)
	}
	if err != nil {
		core.Logger.Error("[selectServer]选服出错,userId:%v,selectType:%v,version:%v, code:%v,err:%v", userId, selectType, version, err.GetCode(), err.Error())
	} else {
		core.Logger.Info("[selectServer]选服成功,userId:%v,selectType:%v,version:%v,roomId:%v,remote:%v", userId, selectType, version, roomId, remote)
	}
	responseData(w, err, remote, roomId, raceId)
	core.Logger.Debug("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

// 专为内网提供的选服服务(暂时只有h5使用)
func selectServerForLocal(w http.ResponseWriter, r *http.Request) {
	core.Logger.Debug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	core.Logger.Debug("收到内网的选服请求:%v", r.RemoteAddr)

	// 解析参数
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("user_id"))
	version := r.Form.Get("version")
	selectType := r.Form.Get("select_type")
	core.Logger.Debug("[selectServerForLocal]解析参数,userId:%v, selectType:%v, version:%v", userId, selectType, version)

	// 执行选服操作
	result := doSelectServer(userId, selectType, version, false, false)

	if result.Code != 0 {
		core.Logger.Error("[selectServerForLocal]选服出错,userId:%v,selectType:%v,version:%v, code:%v,err:%v", userId, selectType, version, result.Code, result.Errmsg)
	} else {
		core.Logger.Info("[selectServerForLocal]选服成功,userId:%v,selectType:%v,version:%v,roomId:%v,remote:%v", userId, selectType, version, result.RoomId, result.RoomRemote)
	}

	var returnString string
	if result.Code != 0 && result.RaceId == 0 {
		returnString = strconv.Itoa(result.Code)
	} else if result.RoomRemote != "" {
		returnString = result.RoomRemote + "|" + strconv.FormatInt(result.RoomId, 10) + "|" + strconv.FormatInt(result.RaceId, 10)
	}
	w.Write([]byte(returnString))

}

// 为网关提供的选服接口
func selectIntact(w http.ResponseWriter, r *http.Request) {
	core.Logger.Debug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	core.Logger.Debug("[selectIntact]:%v", r.RemoteAddr)

	// 解析参数
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("user_id"))
	version := r.Form.Get("version")
	selectType := r.Form.Get("select_type")
	core.Logger.Debug("[selectIntact]解析参数,userId:%v, selectType:%v, version:%v", userId, selectType, version)

	result := doSelectServer(userId, selectType, version, false, true)
	core.Logger.Info("[selectIntact]doSelectServer result:%#v", result)

	bs, _ := json.Marshal(result)
	w.Write(bs)
}

// 获取用户当前的房间数据
func getMyServerForLocal(w http.ResponseWriter, r *http.Request) {
	core.Logger.Debug("读取用户当前房间:%v", r.RemoteAddr)
	// 解析参数
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("user_id"))

	// 读取用户当前房间
	roomId, remote, err := getUserServer(userId)
	if err != nil {
		core.Logger.Error("[selectServerForLocal]选服出错,userId:%v, code:%v,err:%v", userId, err.GetCode(), err.Error())
	} else {
		core.Logger.Info("[selectServerForLocal]选服成功,userId:%v,roomId:%v,remote:%v", userId, roomId, remote)
	}

	var raceId int64
	if roomId == 0 || remote == "" {
		raceId, remote = getUserRace(userId)
	}
	responseDataForLocal(w, err, remote, roomId, raceId)
}

// 获取新的服务器
func leagueRoomServer(w http.ResponseWriter, r *http.Request) {
	server := getNewServer(0, "LEAGUE", "latest")
	core.Logger.Info("[newServer]remote:%v", server)
	w.Write([]byte(server))
}

// 获取新的服务器
func selectServerRemote(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	version := r.Form.Get("version")
	selectType := r.Form.Get("select_type")
	server := getNewServer(0, selectType, version)
	core.Logger.Info("[selectServerRemote]remote:%v", server)
	w.Write([]byte(server))
}

func doSelect(userId int, selectType, version string) (roomId int64, remote string, raceId int64, err *ierror.Error) {
	// 读取用户当前房间
	roomId, remote, err = getUserServer(userId)

	if roomId == int64(0) {
		raceId, remote = getUserRace(userId)
	}

	if roomId == 0 || remote == "" {
		// 如果用户当前处于某个房间内，需要直接返回当前房间信息
		// 如果用户未处于房间内，需要根据selectType来做分支处理
		// selectType=房间number的时候，表示加入房间
		// selectType=CREATE_ROOM、RANDOM_ROOM、KING_ROOM、CREATE_TERMINAL_ROOM表示新建房间
		// selectType=h5表示唤起app时的重连，统一返回-1019

		switch selectType {
		case "h5":
			if (roomId == 0 && err == nil) || err != nil {
				err = ierror.NewError(-1019)
			}
		case "RECONNECT_ROOM":
			if roomId == 0 && err == nil {
				err = ierror.NewError(-1002)
			}
		case "CREATE_ROOM":
			fallthrough
		case "RANDOM_ROOM":
			fallthrough
		case "KING_ROOM":
			fallthrough
		case "COIN_ROOM":
			fallthrough
		case "CREATE_TERMINAL_ROOM":
			fallthrough
		case "LEAGUE_SERVER":
			fallthrough
		case "RANK":
			// 创建房间的选服
			remote = getNewServer(userId, selectType, version)
			if remote == "" {
				// 选择新服失败
				err = ierror.NewError(-1030)
			} else {
				// 选服成功，清空原来的错误
				err = nil
			}
		default: // 加入房间， 需要判断加入的房间是否存在
			roomId, remote, err = getJoinServer(userId, selectType)
		}
	}

	// 如果选到了服，则将ip地址转成成slb
	if remote != "" {
		remote = servers.IPMap.Conversion(remote)
	}

	core.Logger.Debug("[doSelect]userId:%v,version:%v", userId, version)
	return
}

// 选服逻辑
// enableSlbExchange 是否需要做slb转换
// enableRaceIdAndRoomId 在有roomId的情况下，是否继续获取raceId
func doSelectServer(userId int, selectType, version string, enableSlbExchange, enableRaceIdAndRoomId bool) *Result {
	core.Logger.Debug(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	core.Logger.Debug("[doSelectServer]userId:%v, selectType:%v, version:%v, enableSlbExchange:%v, enableRaceIdAndRoomId:%v",
		userId, selectType, version, enableSlbExchange, enableRaceIdAndRoomId)

	var roomId, raceId int64
	var roomRemote, raceRemote string
	var err *ierror.Error

	// 读取用户当前房间
	roomId, roomRemote, err = getUserServer(userId)
	// 读取当前比赛
	if roomId == 0 || enableRaceIdAndRoomId {
		raceId, raceRemote = getUserRace(userId)
	}

	// 根据用户当前的房间，

	if roomId == 0 || roomRemote == "" {
		// 如果用户当前处于某个房间内，需要直接返回当前房间信息
		// 如果用户未处于房间内，需要根据selectType来做分支处理
		// selectType=房间number的时候，表示加入房间
		// selectType=CREATE_ROOM、RANDOM_ROOM、KING_ROOM、CREATE_TERMINAL_ROOM表示新建房间
		// selectType=h5表示唤起app时的重连，统一返回-1019

		switch selectType {
		case "h5":
			if (roomId == 0 && err == nil) || err != nil {
				err = ierror.NewError(-1019)
			}
		case "RECONNECT_ROOM":
			if roomId == 0 && err == nil {
				err = ierror.NewError(-1002)
			}
		case "CREATE_ROOM":
			fallthrough
		case "RANDOM_ROOM":
			fallthrough
		case "KING_ROOM":
			fallthrough
		case "COIN_ROOM":
			fallthrough
		case "CREATE_TERMINAL_ROOM":
			fallthrough
		case "LEAGUE_SERVER":
			fallthrough
		case "RANK":
			// 创建房间的选服
			roomRemote = getNewServer(userId, selectType, version)
			if roomRemote == "" {
				// 选择新服失败
				err = ierror.NewError(-1030)
			} else {
				// 选服成功，清空原来的错误
				err = nil
			}
		default: // 加入房间， 需要判断加入的房间是否存在
			roomId, roomRemote, err = getJoinServer(userId, selectType)
			core.Logger.Debug("==========err:%v", err)
		}
	}

	core.Logger.Debug("==========err:%v", err)

	result := &Result{
		RoomId:     roomId,
		RoomRemote: roomRemote,
		RaceId:     raceId,
		RaceRemote: raceRemote,
	}
	if err != nil {
		result.Code = err.GetCode()
		result.Errmsg = err.Error()
	}

	core.Logger.Debug("==========result:%v", result)

	if enableSlbExchange {
		if len(result.RoomRemote) > 0 {
			result.RoomRemote = servers.IPMap.Conversion(result.RoomRemote)
		}
		if len(result.RaceRemote) > 0 {
			result.RaceRemote = servers.IPMap.Conversion(result.RaceRemote)
		}
	}

	return result
}

// 校验参数
func verify(userId int, selectType, token string) (string, *ierror.Error) {
	var version string
	if userId == 0 {
		return version, ierror.NewError(-1001)
	}
	version, err := verifyToken(userId, token)
	if err != nil {
		return version, err
	}
	return version, nil
}

// 检查token
func verifyToken(userId int, token string) (string, *ierror.Error) {
	var version string
	if token == "" {
		core.Logger.Error("[verifyToken]token missed, userId:%v", userId)
		return version, ierror.NewError(-1007)
	}
	// 解析token
	tokenInfo, err := util.CheckToken(token, config.TOKEN_SECRET_KEY)
	if err != nil {
		core.Logger.Error("[verifyToken]token parsed error, userId:%v,err:%v", userId, err.Error())
		return version, ierror.NewError(-1007)
	}
	// token解析出的userId与传入的是否一致
	if int(tokenInfo[0].(float64)) != userId {
		core.Logger.Error("[verifyToken]token parse userId error, userId:%v,parsedUserId:%v", userId, int(tokenInfo[0].(float64)))
		return version, ierror.NewError(-1007)
	}
	version = tokenInfo[3].(string)

	// 验证token与cache中存储的�������否一致
	cacheToken, _ := core.RedisDoString(core.RedisClient0, "get", fmt.Sprintf(config.CACHE_KEY_USER_TOKEN, userId))
	if cacheToken != token {
		core.Logger.Error("[verifyToken]token != cachedToken, userId:%v,token:%v,cacheToken:%v", userId, token, cacheToken)
		return version, ierror.NewError(-1007)
	}
	return version, nil
}

func responseData(w http.ResponseWriter, err *ierror.Error, remote string, roomId int64, raceId int64) {
	var ip string
	var port int
	if remote != "" {
		s := strings.Split(remote, ":")
		ip = s[0]
		port, _ = strconv.Atoi(s[1])
	}

	builder := flatbuffers.NewBuilder(0)
	var commonResult flatbuffers.UOffsetT
	ipString := builder.CreateString(ip)

	// 解析错误
	errCode := 0
	errMsg := ""
	if err != nil {
		errCode = err.GetCode()
		errMsg = err.Error()
	}
	errMsgString := builder.CreateString(errMsg)

	fbsInfo.CommonResultStart(builder)
	fbsInfo.CommonResultAddCode(builder, int32(errCode))
	fbsInfo.CommonResultAddMsg(builder, errMsgString)
	commonResult = fbsInfo.CommonResultEnd(builder)

	fbsInfo.SelectServerResponseStart(builder)
	fbsInfo.SelectServerResponseAddResult(builder, commonResult)
	fbsInfo.SelectServerResponseAddIp(builder, ipString)
	fbsInfo.SelectServerResponseAddPort(builder, int32(port))
	fbsInfo.SelectServerResponseAddRoomId(builder, uint32(roomId))
	fbsInfo.SelectServerResponseAddRaceId(builder, raceId)
	orc := fbsInfo.SelectServerResponseEnd(builder)
	builder.Finish(orc)
	w.Write(builder.FinishedBytes())
}

func responseDataForLocal(w http.ResponseWriter, err *ierror.Error, remote string, roomId int64, raceId int64) {
	var returnString string
	if err != nil {
		returnString = strconv.Itoa(err.GetCode())
	} else if remote != "" {
		returnString = remote + "|" + strconv.FormatInt(roomId, 10) + "|" + strconv.FormatInt(raceId, 10)
	}
	w.Write([]byte(returnString))
}

// 获取用户当前的房间
func getUserServer(userId int) (int64, string, *ierror.Error) {
	// 读取用户当前房间id
	roomId, _ := core.RedisDoInt64(core.RedisClient2, "get", fmt.Sprintf(config.CACHE_KEY_USER_ROOMID, userId))

	if roomId > 0 {
		remote, err := getServerByRoomId(userId, roomId)
		if err != nil {
			return 0, "", err
		}
		return roomId, remote, nil
	}

	// 如果用户不在房间内，则取用户当前是否在联赛内
	return 0, "", nil
}

// 获取待加入的房间
func getJoinServer(userId int, roomNum string) (int64, string, *ierror.Error) {
	roomId, _ := core.RedisDoInt64(core.RedisClient1, "get", fmt.Sprintf(config.CACHE_KEY_ROOM_NUMBER_ID, roomNum))
	if roomId > 0 {
		remote, err := getServerByRoomId(userId, roomId)
		if err != nil {
			return 0, "", err
		}
		return roomId, remote, nil
	}

	return 0, "", ierror.NewError(-1002)
}

func getServerByRoomId(userId int, roomId int64) (string, *ierror.Error) {
	// 寻找房间id对应的remote信息
	remote, _ := core.RedisDoString(core.RedisClient1, "get", fmt.Sprintf(config.CACHE_KEY_ROOM_REMOTE, roomId))
	if remote == "" {
		core.Logger.Error("根据roomId查找remote失败,userId:%v,roomId:%v", userId, roomId)
		return "", ierror.NewError(-1015)
	}
	// 判断游戏服务器，是否健康
	isActive := servers.IsActive(remote)
	core.Logger.Debug("检测远程服务器是否活跃,userId:%v,roomId:%v,remote:%v,isActive:%v", userId, roomId, remote, isActive)
	if !isActive {
		core.Logger.Error("用户对应的游戏服务器已经挂了,userId:%v,roomId:%v,remote:%v", userId, roomId, remote)
		return "", ierror.NewError(-1015)
	}
	// 仅仅上面两个判断还是不够的，还需要判断游戏服上是否真的有这个房间
	roomExists, _ := core.RedisDoBool(core.RedisClient3, "sismember", fmt.Sprintf(config.CACHE_KEY_HALL_ROOM_IDS, remote), roomId)
	if !roomExists {
		core.Logger.Error("远程服务器上已经不存在这个房间了,userId:%v,roomId:%v,remote:%v", userId, roomId, remote)
		return "", ierror.NewError(-1015)
	}
	return remote, nil
}

// 选择一个服务器
// 第一判断预发规则
// 第二判断灰度规则
// 第三去正常服务中选
func getNewServer(userId int, selectType string, version string) string {
	// 当前活跃的游戏服列表
	activeServers := servers.GetActiveServers()
	core.Logger.Debug("[getNewServer]当前活跃的服务器列表,userId:%v,activeServers:%v", userId, activeServers)
	// 有效服务器列表
	validServers := map[string]bool{}

	// 找出所有活跃的服务器
	for _, s := range servers.GetGameServers() {
		// 跳过游戏类型不匹配的游戏服
		requireServerTypes := chooseServerTypeBySelectType(selectType)
		if !util.InStringSlice(s.ServerType, requireServerTypes) {
			core.Logger.Debug("[getNewServer]找出有效服务器，跳过不匹配的游戏类型,userId:%v,remote:%v,was:%v,require:%v", userId,
				s.Remote, s.ServerType, requireServerTypes)
			continue
		}
		// 跳过清人的游戏服
		if s.Enable == "0" {
			core.Logger.Debug("[getNewServer]找出有效服务器，跳过清人的服务器,userId:%v,remote:%v", userId, s.Remote)
			continue
		}
		// 跳过不活跃的游戏服
		if !util.InStringSlice(s.Remote, activeServers) {
			core.Logger.Debug("[getNewServer]找出有效服务器，跳过不活跃的游戏服,userId:%v,remote:%v", userId, s.Remote)
			continue
		}
		validServers[s.Remote] = true
	}
	core.Logger.Debug("[getNewServer]当前有效的服务器列表,userId:%v,selectType:%v,validServers:%#v", userId, selectType, validServers)

	if len(validServers) == 0 {
		core.Logger.Error("无可用的活跃服务器,userId:%v,selectType:%v", userId, selectType)
		return ""
	}

	// 匹配预发规则
	remote := servers.PRule.CheckPreview(userId, validServers)
	if remote != "" {
		core.Logger.Debug("用户进入预发服务器,userId:%v,remote:%v", userId, remote)
		return remote
	}

	// 匹配灰度规则
	remote = servers.CheckGray(userId, validServers, version)
	if remote != "" {
		core.Logger.Debug("用户进入灰度服务器,userId:%v,remote:%v,version:%v", userId, remote, version)
		return remote
	}

	// 从剩下的服务器中随机一个
	if len(validServers) > 0 {
		core.Logger.Debug("随机一个可用服务器,可用列表,userId:%v,version:%v,validServers:%v", userId, remote, validServers)
		s := make([]string, 0, len(validServers))
		for remote := range validServers {
			s = append(s, remote)
		}
		sort.Strings(s)
		remote = s[util.RandIntn(len(s))]

		core.Logger.Debug("用户随机一个可用服务器，userId:%v,remote:%v,version:%v", userId, remote, version)
		return remote
	}
	return ""
}

// 获取用户是否在房间内
func getUserRace(userId int) (raceId int64, remote string) {
	url := leagueUrl + "?userId=" + strconv.Itoa(userId)
	core.Logger.Debug("[getUserRace]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[getUserRace]het.Get, url:%v, error:%v", url, err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[getUserRace]het.Get read body, url:%v, error:%v", url, err.Error())
	}
	core.Logger.Debug("[getUserRace]het.Get read body,body:%v", string(body))

	js, _ := simplejson.NewJson(body)
	code, _ := js.Get("code").Int()
	if code < 0 {
		message, _ := js.Get("message").String()
		core.Logger.Error("[getUserRace]userId:%v, code:%v, msg:%v", userId, code, message)
	} else {
		raceId, _ = js.Get("raceId").Int64()
		if raceId > 0 {
			remote = getNewServer(0, "LEAGUE_SERVER", "latest")
		}
		core.Logger.Info("[getUserRace]userId:%v, raceId:%v, remote:%v", userId, raceId, remote)
	}
	return
}

// 根据客户端传入的选服类型，来返回匹配的服务器类型
func chooseServerTypeBySelectType(selectType string) []string {
	switch selectType {
	case "CREATE_ROOM":
		return []string{"0", "1"}
	case "RANDOM_ROOM":
		return []string{"0", "2"}
	case "KING_ROOM":
		return []string{"0", "3"}
	case "COIN_ROOM":
		return []string{"0", "5"}
	case "CREATE_TERMINAL_ROOM":
		return []string{"0", "4"}
	case "LEAGUE":
		return []string{"0", "6"}
	case "LEAGUE_SERVER":
		return []string{"7"}
	case "RANK":
		return []string{"0", "8"}
	default:
		return []string{"0"}
	}
}
