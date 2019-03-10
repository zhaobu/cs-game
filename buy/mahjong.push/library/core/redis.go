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
	RedisClient1 *redis.Pool
	redisNode1   string
)

const (
	// DBMember redis db for member
	RedisDBPUSH = 4
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
	// 建立连接池
	RedisClient1 = &redis.Pool{
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

// RedisDo 执行redis命令
func RedisDo(client *redis.Pool, db int, command string, args ...interface{}) (reply interface{}, err error) {
	redisConn := client.Get()
	defer redisConn.Close()
	redisConn.Send("select", db)
	return redisConn.Do(command, args...)
}

// RedisDoInt 执行redis命令，且返回int值
func RedisDoInt(client *redis.Pool, db int, command string, args ...interface{}) (int, error) {
	return redis.Int(RedisDo(client, db, command, args...))
}

// RedisDoInt64 执行redis命令，且返回int64值
func RedisDoInt64(client *redis.Pool, db int, command string, args ...interface{}) (int64, error) {
	return redis.Int64(RedisDo(client, db, command, args...))
}

// RedisDoBool 执行redis命令，且返回bool
func RedisDoBool(client *redis.Pool, db int, command string, args ...interface{}) (bool, error) {
	return redis.Bool(RedisDo(client, db, command, args...))
}

// RedisDoString 执行redis命令，且返回string
func RedisDoString(client *redis.Pool, db int, command string, args ...interface{}) (string, error) {
	return redis.String(RedisDo(client, db, command, args...))
}
