package svr

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aminpaks/go-streams/pkg/async"
	"github.com/aminpaks/go-streams/pkg/h"
	"github.com/aminpaks/go-streams/pkg/mw"
	"github.com/aminpaks/go-streams/pkg/re"
	"github.com/aminpaks/go-streams/pkg/users"
)

func New(depsCtx context.Context, port string) {
	syncGroup := async.NewSyncGroup()
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	router := chi.NewRouter()

	log.Print("Prepping to start up server...\n________________________________")
	router.Use(mw.RequestLog)
	router.Use(mw.Logger)
	router.Use(mw.Recoverer)

	router.Route("/users", func(r chi.Router) {
		err := users.NewUserController(depsCtx, shutdownCtx, syncGroup, r)
		if err != nil {
			log.Fatalf("failed to initialize user controller: %v", err)
		}
	})

	router.NotFound(h.New(func(rw http.ResponseWriter, r *http.Request) h.Renderer {
		return re.Json(http.StatusNotFound, re.JsonErrors(re.ToJsonErrors("Not found")...))
	}))

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Print("Attempting to start server...")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Print("Waiting for shutdown signal...")
	<-shutdownSignal
	log.Print("Attempting to shutdown server...")
	shutdown()

	// Just a timeout cancelation context to shutdown the server after 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	log.Print("before shutdown -- ", time.Now().Format(time.Stamp))
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}
	// Waits up to 10 seconds for all the concurrent workers to completed their work
	if timedout := syncGroup.WaitTimeout(time.Second * 10); timedout {
		log.Print("-- shutdown timedout!")
	}
	log.Print("after shutdown -- ", time.Now().Format(time.Stamp))
}
