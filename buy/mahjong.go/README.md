技术选型：
1、mysql
	阿里云rds服务器
	分成写库、从库、日志库

2、Redis
	阿里云redis，多实例配置，不做分布式

3、配置文件管理
	格式：toml
	不同环境不同配置文件夹，代码同一份

4、日志输出
	beelog

5、错误、异常处理


6、服务器监控
	a.基础监控，有阿里云提供
	b.活动服务器监控:
		进程每秒更新redis中的活动时间戳，php程序监控程序此时间戳，超过5s未更新，则被认为服务器异常
	c.业务监控
		暂无

7、orm
	beeorm
	
8、vendor
	glide

9、项目发布
	ssh连上远程服务器之后，执行 godep 命令
	后期需要统一管理
	
10、数据传输
	json + flatbuffers
	
11、日志管理
filebeat -> logstash -> elasticsearch -> kibana -> nginx


* 项目启动方式
项目必须在mahjong.go下执行，否则会加载配置文件失败

eg:
cd /Users/xh/gojob/src/mahjong.go
../../bin/mahjong.go
