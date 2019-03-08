package config

import (
	"github.com/fwhappy/util"
)

// ServerNameMap 服务器ip=>name对照表
var ServerNameMap = map[string]string{
	"127.0.0.1":     "local",
	"10.26.239.72":  "qa",
	"10.27.96.45":   "g1",
	"10.28.35.195":  "g2",
	"10.27.96.49":   "g3",
	"10.132.21.143": "g4",
	"10.80.109.66":  "g5",
	"10.80.109.18":  "g6",
}

// ServerPorts 服务器对应的端口范围，做一层保护
var ServerPorts = map[string][]int{
	"local": []int{9000},
	"qa":    []int{8990, 8991, 8992, 8993, 8994, 8995, 8996, 8997, 8998, 8999, 9000},
	"g1":    []int{8990, 8991, 8992, 8993, 8994, 8995, 8996, 8997, 8998, 8999},
	"g2":    []int{8980, 8981, 8982, 8983, 8984, 8985, 8986, 8987, 8988, 8989},
	"g3":    []int{8970, 8971, 8972, 8973, 8974, 8975, 8976, 8977, 8978, 8979},
	"g4":    []int{8960, 8961, 8962, 8963, 8964, 8965, 8966, 8967, 8968, 8969},
	"g5":    []int{8950, 8951, 8952, 8953, 8954, 8955, 8956, 8957, 8958, 8959},
	"g6":    []int{8940, 8941, 8942, 8943, 8944, 8945, 8946, 8947, 8948, 8949},
}

// GetServerNameByIP 根据服务器ip或者服务器名称
func GetServerNameByIP(ip string) string {
	return ServerNameMap[ip]
}

// VerifyServerPort 验证服务器端口是否正确
func VerifyServerPort(server string, port int) bool {
	serverName := ServerNameMap[server]
	return util.IntInSlice(port, ServerPorts[serverName])
}
