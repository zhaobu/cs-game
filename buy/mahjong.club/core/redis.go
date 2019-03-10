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
	RedisClient0 *redis.Pool

	redisNode1 string
)

const (
	RedisDBCommon = 0 // 通用
)

type redisConfig struct {
	Host string
	Port int
	Auth string
}

type redisConfigMap struct {
	Node map[string]redisConfig
}

func LoadRedisConfig(cfg_file string) {
	content, err := ioutil.ReadFile(cfg_file)
	if err != nil {
		panic(err)
	}
	var conf_struct redisConfigMap
	if _, err := toml.Decode(string(content), &conf_struct); err != nil {
		panic(err)
	}
	// fmt.Printf("%V\n", conf_struct)
	conf := conf_struct.Node

	redisNode1 = fmt.Sprintf("%s:%d", conf["redisNode1"].Host, conf["redisNode1"].Port)
	// redisNode2 = fmt.Sprintf("%s:%d", conf["redisNode2"].Host, conf["redisNode2"].Port)
	// redisNode2Slave = fmt.Sprintf("%s:%d", conf["redisNode2Slave"].Host, conf["redisNode2Slave"].Port)
	// 建立连接池
	RedisClient0 = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     100,
		MaxActive:   2000,
		Wait:        true,
		IdleTimeout: 8 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisNode1)
			if err != nil {
				return nil, err
			}
			if len(conf["redisNode1"].Auth) > 0 {
				if _, err := c.Do("AUTH", conf["redisNode1"].Auth); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
	}
}

// 执行redis命令
func RedisDo(client *redis.Pool, command string, args ...interface{}) (reply interface{}, err error) {
	redisConn := client.Get()
	defer redisConn.Close()
	return redisConn.Do(command, args...)
}

func RedisDoInt(client *redis.Pool, command string, args ...interface{}) (int, error) {
	return redis.Int(RedisDo(client, command, args...))
}
func RedisDoInts(client *redis.Pool, command string, args ...interface{}) ([]int, error) {
	return redis.Ints(RedisDo(client, command, args...))
}

func RedisDoInt64(client *redis.Pool, command string, args ...interface{}) (int64, error) {
	return redis.Int64(RedisDo(client, command, args...))
}

func RedisDoBool(client *redis.Pool, command string, args ...interface{}) (bool, error) {
	return redis.Bool(RedisDo(client, command, args...))
}

func RedisDoString(client *redis.Pool, command string, args ...interface{}) (string, error) {
	return redis.String(RedisDo(client, command, args...))
}

func RedisDoBytes(client *redis.Pool, command string, args ...interface{}) ([]byte, error) {
	return redis.Bytes(RedisDo(client, command, args...))
}

// RedisDoStringMap 从redis中取出string
func RedisDoStringMap(client *redis.Pool, command string, args ...interface{}) (map[string]string, error) {
	return redis.StringMap(RedisDo(client, command, args...))
}

func RedisDoFloat64(client *redis.Pool, command string, args ...interface{}) (float64, error) {
	return redis.Float64(RedisDo(client, command, args...))
}
