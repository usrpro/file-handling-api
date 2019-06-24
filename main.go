package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

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

func main() {
	iSig := make(chan os.Signal, 1)
	signal.Notify(iSig, os.Interrupt)
	s := http.Server{Addr: "0.0.0.0:9090", Handler: http.HandlerFunc(imageHandler)}
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
