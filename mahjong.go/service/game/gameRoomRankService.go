package game

import (
	"sync"
)

// RankRoomQueue 排位赛房间队列
type RankRoomQueue struct {
	Rooms     *sync.Map
	RankLevel int // 排位赛等级
}

// RankRoomQueueFactory 金币场队列容器
type RankRoomQueueFactory struct {
	Queues *sync.Map
}

// NewRankRoomQueue 新建金币场房间队列
func NewRankRoomQueue(level int) *RankRoomQueue {
	return &RankRoomQueue{
		Rooms:     &sync.Map{},
		RankLevel: level,
	}
}

// NewRankRoomQueueFactory 新建金币场队列容器
func NewRankRoomQueueFactory() *RankRoomQueueFactory {
	return &RankRoomQueueFactory{
		Queues: &sync.Map{},
	}
}

// Put 写入一个队列到容器
func (factory *RankRoomQueueFactory) Put(queue *RankRoomQueue) {
	factory.Queues.Store(queue.RankLevel, queue)
}

// Get 读取一个队列到容器
// 如果容器不存在，则更新一个容器
func (factory *RankRoomQueueFactory) Get(level int) *RankRoomQueue {
	value, exists := factory.Queues.Load(level)
	if !exists {
		queue := NewRankRoomQueue(level)
		factory.Put(queue)
		return queue
	}
	return value.(*RankRoomQueue)
}
