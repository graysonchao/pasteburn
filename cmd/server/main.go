package main

import (
	"flag"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/graysonchao/pasteburn"
	"github.com/zenazn/goji"
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

	goji.Use(pasteburn.CorsMiddleware)

	goji.Get("/api/text/view", pasteburn.MakeTextViewHandler(ctx, s))
	goji.Post("/api/text/create", pasteburn.MakeTextAddHandler(ctx, s))
	goji.Get("/api/multi/view", pasteburn.MakeMultiTextViewHandler(ctx, s))
	goji.Post("/api/multi/create", pasteburn.MakeMultiTextAddHandler(ctx, s))
	goji.Get("/api/image/view", pasteburn.MakeImageViewHandler(ctx, s))
	goji.Post("/api/image/create", pasteburn.MakeImageAddHandler(ctx, s))
	goji.Serve()
}
