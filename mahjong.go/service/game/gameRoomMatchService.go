package game

import "sync"

// MatchRoomQueue 随机房间队列
type MatchRoomQueue struct {
	Rooms map[int64]string

	// 读写锁
	Mux *sync.RWMutex
}

// NewMatchRoomQueue 新建随机房间队列
func NewMatchRoomQueue() *MatchRoomQueue {
	queue := &MatchRoomQueue{}
	queue.Rooms = map[int64]string{}
	queue.Mux = &sync.RWMutex{}

	return queue
}

// Add 添加一个房间到队列
func (queue *MatchRoomQueue) Add(roomId int64, number string) {
	queue.Mux.Lock()
	defer queue.Mux.Unlock()

	queue.Rooms[roomId] = number
}

// Del 从队列中删除一个房间
// 如果房间存在, 返回true, 否则返回false
func (queue *MatchRoomQueue) Del(roomId int64) bool {
	queue.Mux.Lock()
	defer queue.Mux.Unlock()

	_, exists := queue.Rooms[roomId]
	if exists {
		delete(queue.Rooms, roomId)
	}

	return exists
}

// MatchRoomQueueFactory 随机队列仓库
type MatchRoomQueueFactory struct {
	QueueList map[int]*MatchRoomQueue

	Mux *sync.RWMutex
}

// NewMatchRoomQueueFactory 生成一个随机队列仓库
func NewMatchRoomQueueFactory() *MatchRoomQueueFactory {
	factory := &MatchRoomQueueFactory{}
	factory.QueueList = make(map[int]*MatchRoomQueue)
	factory.Mux = &sync.RWMutex{}

	return factory
}

// Put 向队列仓库中新加一条推列
func (factory *MatchRoomQueueFactory) Put(gameType int, queue *MatchRoomQueue) {
	factory.Mux.Lock()
	defer factory.Mux.Unlock()

	factory.QueueList[gameType] = queue
}

// Get 从队列仓库中获取一条队列
// 如果不存在，则新建一条队列
func (factory *MatchRoomQueueFactory) Get(gameType int) *MatchRoomQueue {
	factory.Mux.Lock()
	defer factory.Mux.Unlock()

	queue, exists := factory.QueueList[gameType]
	if !exists {
		queue = NewMatchRoomQueue()
		factory.QueueList[gameType] = queue
	}

	return queue
}
