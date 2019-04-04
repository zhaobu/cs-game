package main

import (
	"io"
	"os"

	"flag"
	"fmt"

	"go.uber.org/zap"
)

var (
	// wxID = flag.String("wxid", "wx_1", "wx id")
	// addr = flag.String("addr", "localhost:9876", "tcp listen address")
	addr     = flag.String("addr", "192.168.0.10:9876", "tcp listen address")
	fileName = flag.String("fileName", "./desk.json", "deskInfo readfileName")

	log  *zap.SugaredLogger //printf风格
	tlog *zap.Logger        //structured 风格
)

func main() {
	flag.Parse()

	player := &Player{}
	player.Connect(*addr)
}

func writebuf(path string, wbuf string) {
	//打开文件，新建文件
	f, err := os.Create(path)
	if err != nil {
		fmt.Println("err = ", err)
		return
	}

	//使用完毕，需要关闭文件
	defer f.Close()
	_, err = f.WriteString(wbuf) //n表示写入的字节数
	if err != nil {
		fmt.Println("err = ", err)
	}
}

func readFile(path string) []byte {
	//打开文件
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("err = ", err)
		return nil
	}

	//关闭文件
	defer f.Close()

	buf := make([]byte, 1024*2) //2k大小

	//n代表从文件读取内容的长度
	n, err1 := f.Read(buf)
	if err1 != nil && err1 != io.EOF { //文件出错，同时没有到结尾
		fmt.Println("err1 = ", err1)
		return nil
	}
	fmt.Printf("read from%s:%s", path, string(buf[:n]))
	return buf[:n]
}
