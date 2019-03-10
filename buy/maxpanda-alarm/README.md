# maxpanda报警服务
===

维护人 | 版本号 | 描述 | 日期
--- | --- | --- | ---
王虎 | 0.1 | 初版 | 2017-09-29

## 1. 概述

* 支持报警方式
	* 邮件报警
	* 短信报警

* host:
	* 测试环境:http://10.26.239.72:13579/
	* 生产环境:http://10.27.5.231:13579/
* 请求方式：支持POST和GET
* 参数格式：*=json_encode($params)
* api示例
	* data = json_encode(['key':'value',...])
	* http://10.26.239.72:13579/alarm?*=data
	* http://10.26.239.72:13579/alarm/mail?*=data
	* http://10.26.239.72:13579/alarm/mobile?*=data
* API中target字段说明
	* target表示报警对象的配置，默认是`default`，如需定制，请联系报警服务的维护者
	* 每个target在报警服务器对应一个配置文件
	* 配置发送者、接收者的相关信息
* 短信报警
	* 短信报警，接入的是阿里云的短信服务
	* 每中短信报警，都需要先去后台申请一个模板id
* 短信报警的body，支持html标签

## 2. API

### 2.1 全方式报警
* 地址: $host/alarm
* 参数列表:

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
target | string | 是 | 报警对象
subject | string | 是 | 邮件标题
body | string | 是 | 邮件内容
subject | string | 是 | 邮件标题
body | string | 是 | 邮件内容

* 参数示例:

~~~
{"body":"测试报警body","sms_code":"SMS_76115014","sms_params":"{\"code\":1234}","subject":"测试报警subject","target":"default"}
~~~


### 2.2 邮件报警
* 地址: $host/alarm/mail
* 参数列表

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
target | string | 是 | 报警对象
subject | string | 是 | 邮件标题
body | string | 是 | 邮件内容

参数示例:

~~~
{"body":"测试报警body","subject":"测试报警subject","target":"default"}
~~~

### 2.3 短信报警
* 地址: $host/alarm/mobile
* 参数地址

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
target | string | 是 | 报警对象
sms_code | string | 是 | sms模板id，在阿里云后台申请
sms_params | json string | 是 | sms模板参数

参数示例:

~~~
{"sms_code":"SMS_76115014","sms_params":"{\"code\":1234}","target":"default"}
~~~

