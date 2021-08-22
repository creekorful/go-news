package main

import (
	"context"
	"flag"
	"github.com/creekorful/go-news/internal/database"
	"github.com/creekorful/go-news/internal/scheduler"
	"github.com/creekorful/go-news/internal/server"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	// the program version, exported using LDFLAGS
	version = "dev"

	databaseFlag = flag.String("database", "db.sqlite", "path to the SQLite database")
)

func main() {
	flag.Parse()

	log.Printf("starting go-news %s", version)

	db, err := database.Connect(*databaseFlag)
	if err != nil {
		log.Fatalf("error while connecting to the db: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// start the scheduler
	sch := scheduler.NewScheduler(db)
	go func() {
		err := sch.Schedule(ctx, 1*time.Minute)
		if err != nil {
			log.Printf("scheduler error: %s", err)
		}
	}()

	// start the web server
	srv := server.NewServer(db)
	go func() {
		if err := srv.Serve(ctx, "0.0.0.0:8080"); err != nil {
			log.Printf("server error: %s", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	cancel()

	// stop the web server
	_ = srv.Shutdown(ctx)
}
