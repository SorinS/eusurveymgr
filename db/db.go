package db

import (
	"database/sql"
	"eusurveymgr/log"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func ConnectToMySQL(host string, port int, user, password, dbName string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		log.Errorf("MYSQL -- Error pinging DB: %v", err)
		return nil, err
	}
	log.Infof("MYSQL -- Connected to MySQL Server")
	return db, nil
}