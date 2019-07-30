package main

import (
	"github.com/fulldump/goconfig"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type serverConfig struct {
	Listen string
}

type dbConfig struct {
	User, Password string
	Host, DbName   string
	Port           int
	TLS            bool
}

type s3config struct {
	S3Protocol string
	S3Host     string
	S3key      string
	S3sk       string
	S3Bucket   string
}

type configType struct {
	Server serverConfig
	DbConf dbConfig
	S3Conf s3config
}

var config = configType{
	Server: serverConfig{
		Listen: ":9090",
	},
	DbConf: dbConfig{
		Host:     "/run/postgresql",
		Port:     5432,
		DbName:   "s3db_01",
		User:     "postgres",
		Password: "",
	},
	S3Conf: s3config{
		S3Protocol: "https:",
		S3Host:     "",
		S3key:      "",
		S3sk:       "",
		S3Bucket:   "",
	},
}

func init() {
	goconfig.Read(&config)
	log15.Debug("Parsed configuration", "config", config)
}
