package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx"
)

const (
	host     = "localhost"
	port     = 5432
	database = "s3db_01"
	user     = "postgres"
	pass     = ""
)

var db *pgx.ConnPool

func init() {
	poolCfg := pgx.ConnPoolConfig{MaxConnections: 9, AcquireTimeout: time.Second * 9}
	var e error
	poolCfg.ConnConfig = pgx.ConnConfig{Host: host, Port: port, Database: database, User: user, Password: pass}
	db, e = pgx.NewConnPool(poolCfg)
	if e != nil {
		log.Println(e.Error())
		os.Exit(1)
	}
}

func store(name, raddr, app, bucket string) error {
	b, e := ioutil.ReadFile("sql/queries/insert.sql")
	if e != nil {
		return e
	}
	_, e = db.Exec(string(b), name, raddr, app, bucket)
	if e != nil {
		return e
	}
	return nil
}
