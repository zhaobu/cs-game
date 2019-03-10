package core

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
)

// 定义常量
var (
	RedisClient *redis.Pool
	redisNode   string
)

// DB_INDEX 数据所在的redis db编号
const (
	DB_INDEX = 4
)

type redisConfig struct {
	Host string
	Port int
	Auth string
}

type redisConfigMap struct {
	Node map[string]redisConfig
}

// LoadRedisConfig 加载redis配置
func LoadRedisConfig(cfgFile string) {
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		panic(err)
	}
	var confStruct redisConfigMap
	if _, err := toml.Decode(string(content), &confStruct); err != nil {
		panic(err)
	}
	conf := confStruct.Node

	redisNode = fmt.Sprintf("%s:%d", conf["redisNode"].Host, conf["redisNode"].Port)
	// 建立连接池
	RedisClient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     100,
		MaxActive:   2000,
		Wait:        true,
		IdleTimeout: 8 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisNode)
			if err != nil {
				return nil, err
			}
			if len(conf["redisNode"].Auth) > 0 {
				if _, err := c.Do("AUTH", conf["redisNode"].Auth); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, nil
		},
	}
}

// RedisDo 封装执行redis命令
func RedisDo(command string, args ...interface{}) (reply interface{}, err error) {
	redisConn := RedisClient.Get()
	defer redisConn.Close()
	redisConn.Send("select", DB_INDEX)
	return redisConn.Do(command, args...)
}

// RedisDoString 从redis中取出string
func RedisDoString(command string, args ...interface{}) (string, error) {
	return redis.String(RedisDo(command, args...))
}

// RedisDoStringMap 从redis中取出string
func RedisDoStringMap(command string, args ...interface{}) (map[string]string, error) {
	return redis.StringMap(RedisDo(command, args...))
}

// RedisDoBytes 从redis中取出string
func RedisDoBytes(command string, args ...interface{}) ([]byte, error) {
	return redis.Bytes(RedisDo(command, args...))
}

// RedisDoBool 返回redis结果的bool值
func RedisDoBool(command string, args ...interface{}) (bool, error) {
	return redis.Bool(RedisDo(command, args...))
}
