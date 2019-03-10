package game

import (
	"fmt"
	"sync"
)

// CoinRoomQueue 金币场房间队列
type CoinRoomQueue struct {
	Rooms    *sync.Map
	CoinType int // 金币场类型
	GameType int // 游戏类型
}

// CoinRoomQueueFactory 金币场队列容器
type CoinRoomQueueFactory struct {
	Queues *sync.Map
}

// NewCoinRoomQueue 新建金币场房间队列
func NewCoinRoomQueue(coinType, gType int) *CoinRoomQueue {
	return &CoinRoomQueue{
		Rooms:    &sync.Map{},
		CoinType: coinType,
		GameType: gType,
	}
}

// NewCoinRoomQueueFactory 新建金币场队列容器
func NewCoinRoomQueueFactory() *CoinRoomQueueFactory {
	return &CoinRoomQueueFactory{
		Queues: &sync.Map{},
	}
}

// Put 写入一个队列到容器
func (factory *CoinRoomQueueFactory) Put(queue *CoinRoomQueue) {
	factory.Queues.Store(getQueueKey(queue.CoinType, queue.GameType), queue)
}

// Get 读取一个队列到容器
// 如果容器不存在，则更新一个容器
func (factory *CoinRoomQueueFactory) Get(coinType, gType int) *CoinRoomQueue {
	value, exists := factory.Queues.Load(getQueueKey(coinType, gType))
	if !exists {
		queue := NewCoinRoomQueue(coinType, gType)
		factory.Put(queue)
		return queue
	}
	return value.(*CoinRoomQueue)
}

func getQueueKey(coinType, gType int) string {
	return fmt.Sprintf("%d_%d", coinType, gType)
}
