package util

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"
	"unsafe"

	"github.com/golang/protobuf/proto"
)

func SliceByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToSliceByte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// 函　数：生成随机数
// 概　要：
// 参　数：
//      min: 最小值
//      max: 最大值
// 返回值：
//      int64: 生成的随机数
func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

//order为true:两个slice对应值都一样,为false:只要求a中有的元素,b中都存在
func IntSliceEqualBCE(a, b []int, order bool) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}
	if !order { //只要求值相等
		sort.Ints(a)
		sort.Ints(b)
	}
	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

const MIN = 0.000001

//f1>f2时返回false,精度时Min
func Float64Equal(f1, f2 float64) bool {
	return math.Dim(f1, f2) < MIN
}

// pb转json字符串
func PB2JSON(from_pb proto.Message, f bool) string {
	var json_str []byte
	var err error
	if f {
		json_str, err = json.MarshalIndent(from_pb, "", "	")
	} else {
		json_str, err = json.Marshal(from_pb)
	}
	if err != nil {
		fmt.Printf("PB2JSON err:%s", err)
	}
	return string(json_str)
}

func LoadJSON(fp string, v interface{}) {
	fd, err := os.Open(fp)
	if err != nil {
		fmt.Printf("loadJSON err:%s", err)
	}
	defer fd.Close()
	err = json.NewDecoder(fd).Decode(v)
	if err != nil {
		fmt.Printf("loadJSON err:%s", err)
	}
	fmt.Printf("load jsonfile:%s successfully", fp)
}

func WriteJSON(fp string, v interface{}) {
	fd, err := os.Create(fp)
	if err != nil {
		fmt.Printf("WriteJSON err:%s", err)
	}
	defer fd.Close()
	err = json.NewEncoder(fd).Encode(v)
	if err != nil {
		fmt.Printf("WriteJSON err:%s", err)
	}
	fmt.Printf("WriteJSON:%s successfully", fp)
}

func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
