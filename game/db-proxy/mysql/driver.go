package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type MysqlDb struct {
	db  *sql.DB
	dsn string
}

func NewMyDb(dsn string) (mdb *MysqlDb, err error) {
	mdb = &MysqlDb{}
	mdb.db, err = sql.Open(`mysql`, dsn)
	if err != nil {
		return
	}
	mdb.db.SetMaxIdleConns(20)
	mdb.db.SetMaxOpenConns(20)
	err = mdb.db.Ping()
	return
}

func (mdb *MysqlDb) Close() {
	mdb.db.Close()
}

func (mdb *MysqlDb) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	result, err = mdb.db.Exec(query, args...)
	return
}

func (mdb *MysqlDb) Query(query string, args ...interface{}) (result []map[string]string, err error) {
	var rows *sql.Rows
	rows, err = mdb.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	colName, err := rows.Columns()
	if err != nil {
		return
	}
	if len(colName) == 0 {
		err = fmt.Errorf("columns size == 0 query[%s]\n", query)
		return
	}

	vols := make([]sql.NullString, len(colName))
	scans := make([]interface{}, len(colName))

	for idx := range vols {
		scans[idx] = &vols[idx]
	}

	for rows.Next() {
		errScan := rows.Scan(scans...)
		if errScan != nil {
			logrus.Warn(errScan.Error())
			continue
		}

		oneRow := make(map[string]string)
		for idx, v := range vols {
			cn := colName[idx]
			if v.Valid {
				oneRow[cn] = v.String
			} else {
				oneRow[cn] = ""
			}

		}

		result = append(result, oneRow)
	}

	err = rows.Err()

	return
}
