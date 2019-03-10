# 发布工具
===

维护人 | 版本号 | 描述 | 日期
--- | --- | --- | ---
王虎 | 0.1 | 初版 | 2017-10-13

## 1. 概述
* 项目托管在码云(gitee.com)，克隆地址:

~~~
git@gitee.com:maxpanda/mahjong-process-task.go.git
~~~
* 本工具需要在每个游戏服部署
* 部署方法：克隆下来项目之后，进入项目根目录，执行命令 `./deploy` 查看部署命令
* 本项目只可以内网访问
* API中的$host参数，是每个游戏服务器的`内网IP`


## 2. API
### 2.1 检查版本是否安装
* 地址: http://$host:12121/checkInstall
* 请求方式: GET
* 参数列表:

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
version | string | 是 | 游戏版本号

* 示例:

~~~
http://10.26.239.72:12121/checkInstall?version=2.1.7
~~~

* response格式: JSON
* response示例:

~~~
{"code":0,"message":"版本已安装, server:10.26.239.72, version:2.1.7, 安装时间:2017-10-13 15:08:09"}
{"code":-7,"message":"版本未安装, server:10.26.239.72, version:2.1.100"}
~~~

### 2.2 在游戏服安装版本
* 地址: http://$host:12121/install
* 请求方式: GET
* 参数列表:

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
tag | string | 是 | 代码库branch/tag
version | string | 是 | 安装版本

* 示例:

~~~
http://10.26.239.72:12121/install?tag=master&version=2.1.a
~~~

* response格式: JSON
* response示例:

~~~
{"code":0,"message":"安装申请发起成功，预计需要30s左右可完成"}
~~~
* 安装是异步的，所以此接口，只是发起一个安装申请，安装大约需要30秒，可以通过checkInstall接口，判断是否安装完成

### 2.3 启动游戏服
* 地址: http://$host:12121/start
* 请求方式: GET
* 参数列表:

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
port | int | 是 | 服务监听端口
version | string | 是 | 安装版本

* 示例:

~~~
http://10.26.239.72:12121/start?port=8993&version=2.1.a
~~~

* response格式: JSON
* response示例:

~~~
{"code":0,"message":"操作成功"}
~~~

### 2.4 关闭游戏服
* 地址: http://$host:12121/stop
* 请求方式: GET
* 参数列表:

参数名 | 字段类型 | 是否必须 | 描述
--- | --- | --- | ---
port | int | 是 | 服务监听端口

* 示例:

~~~
curl http://10.26.239.72:12121/stop?port=8993
~~~

* response格式: JSON
* response示例:

~~~
{"code":0,"message":"操作成功"}
~~~

## 3. 错误码对照表
* 服务端接口必返回一个code，code=0表示请求成功，小于0表示失败

code值 | 说明
--- | ---
-1 | 操作失败
-2 | 参数缺失
-3 | 参数错误
-4 | 端口错误，操作服务器，未支持此端口
-5 | 服务器未安装对应版本，需要先安装
-6 | 执行命令失败
-7 | 版本未安装
