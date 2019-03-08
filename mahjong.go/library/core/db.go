package core

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	// _ "github.com/go-sql-driver/mysql"
	// "github.com/jmoiron/sqlx"
)

var (
	DBWriter string
	// DBReader string
	// DBLogger string

// DB mysql db pool
// DB *sqlx.DB
)

type dbconfig struct {
	Host    string
	Port    int
	User    string
	Passwd  string
	Db      string
	Charset string
}

type dbconfigMap struct {
	Node map[string]dbconfig
}

func LoadDBConfig(cfg_file string) {
	content, err := ioutil.ReadFile(cfg_file)
	if err != nil {
		panic(err)
	}
	var conf_struct dbconfigMap
	if _, err := toml.Decode(string(content), &conf_struct); err != nil {
		panic(err)
	}
	// fmt.Printf("%V\n", conf_struct)
	conf := conf_struct.Node
	DBWriter = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", conf["appwriter"].User, conf["appwriter"].Passwd, conf["appwriter"].Host, conf["appwriter"].Port, conf["appwriter"].Db, conf["appwriter"].Charset)
	// DBReader = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", conf["appreader"].User, conf["appreader"].Passwd, conf["appreader"].Host, conf["appreader"].Port, conf["appreader"].Db, conf["appreader"].Charset)
	// DBLogger = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", conf["logger"].User, conf["logger"].Passwd, conf["logger"].Host, conf["logger"].Port, conf["logger"].Db, conf["logger"].Charset)
}
