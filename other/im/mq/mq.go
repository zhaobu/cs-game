package mq

import (
	"cy/other/im/mq/redis"
	"fmt"
)

type MQKind int

const (
	KindRedis MQKind = iota + 1
)

type MQ interface {
	Conn(network, address, lKey string) error
	Push([]byte) error
	Pop() ([]byte, error)
}

func NewMQ(k MQKind) MQ {
	if k == KindRedis {
		return &mqredis.Redis{}
	}
	panic(fmt.Sprintf("bad kind %v", k))
}
