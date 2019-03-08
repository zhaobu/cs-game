package config

import "os"

// GetDeployFile 获取deploy文件路径
func GetDeployFile() string {
	return os.Getenv("GOPATH") + "/src/mahjong.go/deploy"
}

// GetMainFile 获取版本对应的主程序文件绝对路径
func GetMainFile(version string) string {
	return os.Getenv("GOBIN") + "/mahjong.go." + version
}

// GetRobotFile 获取版本对应的主程序文件绝对路径
func GetRobotFile(version string) string {
	return os.Getenv("GOBIN") + "/mahjong.go.robot" + version
}
