package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx"
)

var db *pgx.ConnPool

func init() {
	poolCfg := pgx.ConnPoolConfig{MaxConnections: 9, AcquireTimeout: time.Second * 9}
	var e error
	poolCfg.ConnConfig = pgx.ConnConfig{Host: config.DbConf.Host, Port: uint16(config.DbConf.Port), Database: config.DbConf.DbName, User: config.DbConf.User, Password: config.DbConf.Password, TLSConfig: &tls.Config{InsecureSkipVerify: true}}
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
