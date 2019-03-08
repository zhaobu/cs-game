* 麻将项目的推送服务
* 游戏服务器将需要push的内容，以json格式，写入到redis中
* 本服务器从redis中拿到push内容，再推送给用户