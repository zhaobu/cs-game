package mjlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Table struct {
	tbl map[int]int
}

func (this *Table) init() {
	this.tbl = map[int]int{}
}

func (this *Table) check(key int) bool {
	_, ok := this.tbl[key]
	return ok
}

func (this *Table) add(key int) {
	this.tbl[key] = 1
}

func (this *Table) dump(name string) {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		fmt.Printf("生成文件%s失败", name)
		return
	}
	buf := bufio.NewWriter(file)
	for key, _ := range this.tbl {
		n := int(key)
		fmt.Fprintf(buf, "%d\n", n)
	}
	buf.Flush()
}

func (this *Table) load(name string) {
	file, err := os.Open(name)
	defer file.Close()
	if err != nil {
		fmt.Printf("读取文件%s失败\n", name)
		os.Exit(0)
		return
	}

	reader := bufio.NewReader(file)
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		str := string(buf)
		key, _ := strconv.Atoi(str)
		this.tbl[key] = 1
	}
}
