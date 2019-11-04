package cache

func Pub(channel string, message []byte) {
	redisCli.Publish( channel, message)
}
