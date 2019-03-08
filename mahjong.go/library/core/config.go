package core

import (
	"io/ioutil"

	"fmt"

	"github.com/BurntSushi/toml"
)

type AppCfg struct {
	Env                     string   `toml:"env"`                        // 所属环境
	Version                 string   `toml:"version"`                    // 服务器版本, latest表示最新
	EnableDefineTiles       int      `toml:"enable_define_tiles"`        // 是否支持自定义牌, 0|1
	PushUrl                 string   `toml:"push_url"`                   // 消息推送地址
	EnableJoinRandomRoom    int      `toml:"enable_join_random_room"`    // 是否允许用户加入随机房间, 0|1, 如果不允许, 会为每一个用户创建一个随机房间, 随后填充机器人
	EnableRandomRoomSetting int      `toml:"enable_random_room_setting"` // 是否允许指定随机房间的配置
	EnableAutoHosting       int      `toml:"enable_auto_hosting"`        // 是否允许自动托管, 0|1
	ClubRemote              string   `toml:"club_remote"`                // 俱乐部的地址
	PlaybackSaveType        string   `toml:"playback_save_type"`         // 回放存储方式 redis|oss
	OssEndpoint             string   `toml:"oss_endpoint"`               // oss 配置
	OssBucket               string   `toml:"oss_bucket"`
	OssAccessKeyId          string   `toml:"oss_access_key_id"`
	OssAccessKeySecret      string   `toml:"oss_access_key_secret"`
	UserAvatarUrl           string   `toml:"user_avatar_url"`             //用户头像地址
	LeagueRemote            string   `toml:"league_remote"`               // 联赛服务器地址
	LeagueGetUserRaceURL    string   `toml:"leage_get_user_race_url"`     // 获取用户当前raceid的地址
	GatewayRemotes          []string `toml:"gateway_remotes"`             // 网关地址列表
	RankWinStreakRewardsURL string   `toml:"rank_win_streak_rewards_url"` // 排位赛连胜奖励推送地址
	SelectorTokenVerifyURL  string   `toml:"selector_token_verify_url"`   // 选牌服务token验证地址
}

var (
	AppConfig *AppCfg
)

func init() {
	AppConfig = &AppCfg{}
}

// GetConfigFile 读取配置文件路径
func GetConfigFile(filename, env, confDir string) string {
	return fmt.Sprintf("%s/env/%s/%s", confDir, env, filename)
}

// GetSharedConfigFile 读取共享配置文件路径
func GetSharedConfigFile(filename, confDir string) string {
	return fmt.Sprintf("%s/shared/%s", confDir, filename)
}

// LoadAppConfig 载入配置
func LoadAppConfig(file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if _, err := toml.Decode(string(content), &AppConfig); err != nil {
		panic(err)
	}
}

// IsLocal 判断是否本地环境
func IsLocal() bool {
	return AppConfig.Env == "local"
}

// IsProduct 判断是否生产环境
func IsProduct() bool {
	return AppConfig.Env == "product"
}

// IsForReview 判断是否送审环境
// 暂时放在qa上
func IsForReview() bool {
	return AppConfig.Env == "qa"
}

// GetOssCfg 获取OSS配置
func GetOssCfg() (string, string, string, string) {
	return AppConfig.OssEndpoint, AppConfig.OssBucket, AppConfig.OssAccessKeyId, AppConfig.OssAccessKeySecret
}
