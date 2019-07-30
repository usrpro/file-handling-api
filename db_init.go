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
	poolCfg.ConnConfig = pgx.ConnConfig{
		Host:     config.Database.Host,
		Port:     uint16(config.Database.Port),
		Database: config.Database.Name,
		User:     config.Database.User,
		Password: config.Database.Password,
	}
	if config.Database.TLS {
		poolCfg.ConnConfig.TLSConfig = &tls.Config{
			ServerName: config.Database.Host,
		}
	}
	var e error
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
