package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/fwhappy/mahjong/protocal"
	fbs "mahjong.club/fbs/Common"
)

// 接受命令行输入
func onInput() {
	for {
		var input string
		input, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		if len(input) == 0 {
			continue
		} else if input == "quit" || input == "exit" {
			break
		} else if input == "close" {
			conn.Close()
		} else if input == "help" || input == "usage" {
			showUsage()
			continue
		}

		// 包类型
		var packageType uint8
		params := strings.Split(input, ".")
		p1, _ := strconv.Atoi(params[0])
		packageType = uint8(p1)
		command := getParamsInt(1, params)

		if id == 0 {
			if packageType == protocal.PACKAGE_TYPE_DATA && command == fbs.CommandClubI2CRoomListRequest {
				// donothing
			} else if packageType != protocal.PACKAGE_TYPE_HANDSHAKE {
				showClientError("请先登录")
				continue
			}
		}

		switch packageType {
		case protocal.PACKAGE_TYPE_HANDSHAKE: // 握手
			c2sHandShake(params[1:]...)
		case protocal.PACKAGE_TYPE_HANDSHAKE_ACK: // 握手完成
		case protocal.PACKAGE_TYPE_HEARTBEAT: // 心跳
		case protocal.PACKAGE_TYPE_KICK: // 退出
			// break
			// os.Exit(0)
			return
		case protocal.PACKAGE_TYPE_DATA: // 数据包
			onInputData(command, params[2:])
		default:
			showClientError("未支持的包类型:%v", packageType)
		}
	}
}

func onInputData(command int, args []string) {
	switch command {
	case fbs.CommandClubJoinRequest:
		c2sClubJoinRequest(getParamsInt(0, args))
	case fbs.CommandClubQuitRequest:
		c2sClubQuitRequest(getParamsInt(0, args))
	case fbs.CommandClubClubMessageNotify:
		c2sClubClubMessageNotify(getParamsInt(0, args), getParamsInt(1, args), getParams(2, args))
	case fbs.CommandClubClubMessageListNotify:
		c2sClubClubMessageListNotify(getParamsInt(0, args), getParamsInt(1, args), getParamsInt(2, args))
	case fbs.CommandClubI2CRoomListRequest:
		c2sClubI2CRoomListRequest(getParamsInt(0, args))
	default:
		showClientError("[onInputData]未支持的协议id:%v", command)
	}
}

func getParams(position int, args []string) string {
	if position > len(args) {
		return ""
	}
	return args[position]
}

func getParamsInt(position int, args []string) int {
	param := getParams(position, args)
	if param == "" {
		return 0
	}
	value, _ := strconv.Atoi(param)
	return value
}
