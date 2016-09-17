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

	http.HandleFunc("/api/text/view", pasteburn.MakeTextViewHandler(ctx, s))
	http.HandleFunc("/api/text/create", pasteburn.MakeTextAddHandler(ctx, s))
	http.HandleFunc("/api/multi/view", pasteburn.MakeMultiTextViewHandler(ctx, s))
	http.HandleFunc("/api/multi/create", pasteburn.MakeMultiTextAddHandler(ctx, s))
	http.HandleFunc("/api/image/view", pasteburn.MakeImageViewHandler(ctx, s))
	http.HandleFunc("/api/image/create", pasteburn.MakeImageAddHandler(ctx, s))
	http.ListenAndServe("127.0.0.1:8080", nil)
}
