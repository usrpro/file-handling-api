package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func init() {

}

func imageHandler(wr http.ResponseWriter, r *http.Request) {
	wr.Header().Set("Access-Control-Allow-Origin", acao)
	switch r.RequestURI {
	case "/single-image":
		sharedImageHandler(wr, r)
	case "/batch-images":
		sharedBatchImageHandler(wr, r)
	case "/delete-file":
		deleteFileHandler(wr, r)
	}
}

func defTables() {
	b, e := ioutil.ReadFile(strings.Join([]string{"sql", "definition.sql"}, "/"))
	if e != nil {
		log.Println(e.Error())
		os.Exit(1)
	}
	if _, e = db.Exec(string(b)); e != nil {
		log.Println(e.Error())
	}
}

func main() {
	defTables()
	//goconfig.Read(&config)
	//log15.Debug("Parsed configuration", "config", config)
	iSig := make(chan os.Signal, 1)
	signal.Notify(iSig, os.Interrupt)
	s3Client = s3Init()
	s := http.Server{Addr: config.Server.Listen, Handler: http.HandlerFunc(imageHandler)}
	go func() {
		s.ListenAndServe()
	}()
	<-iSig
	ctx, cc := context.WithTimeout(context.TODO(), time.Second*30)
	defer cc()
	log.Println("Shutdown called.")
	if e := s.Shutdown(ctx); e != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
