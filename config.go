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
	Host, Name     string
	Port           int
	TLS            bool
}

type s3config struct {
	Protocol string
	Host     string
	Key      string
	Secret   string
	Bucket   string
	TLS      bool
}

type configType struct {
	Server   serverConfig
	Database dbConfig
	S3       s3config
}

var config = configType{
	Server: serverConfig{
		Listen: ":9090",
	},
	Database: dbConfig{
		Host: "/run/postgresql",
		Name: "s3db_01",
		User: "postgres",
	},
	S3: s3config{
		Host:   "o.auroraobjects.eu",
		Key:    "",
		Secret: "",
		Bucket: "file-api-test",
		TLS:    true,
	},
}

func init() {
	goconfig.Read(&config)
	log15.Debug("Parsed configuration", "config", config)
}
