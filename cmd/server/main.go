package main

import (
	"flag"
	"net/http"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/graysonchao/pasteburn"
)

func main() {
	var (
		dbPath = flag.String("dbpath", "./pasteburn.db", "Database path")
	)
	flag.Parse()

	s, err := pasteburn.NewBoltBackedService(*dbPath)
	if err != nil {
		panic(err)
	}

	log.Info("Starting server...")

	ctx := context.Background()

	http.HandleFunc("/api/view", pasteburn.MakeViewHandler(ctx, s))
	http.HandleFunc("/api/create", pasteburn.MakeAddHandler(ctx, s))
	http.ListenAndServe("127.0.0.1:8080", nil)
}
