package ierror

// ErrorConfig 错误码定义
var errorConfig = map[int]string{
	0:    "操作成功",
	-1:   "操作失败",
	-100: "参数缺失,action:%v, param:%v",
	-101: "参数不合法, action:%v, param:%v, given value:%v",
	-102: "报警对象未支持, target:%v",
	-103: "读取配置文件失败, file:%v",
}
