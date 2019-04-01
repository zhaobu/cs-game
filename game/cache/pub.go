package cache

func Pub(channel string, message []byte) {
	c := redisPool.Get()
	defer c.Close()

	c.Do("PUBLISH", channel, message)
}
