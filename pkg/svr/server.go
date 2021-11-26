package svr

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aminpaks/go-streams/pkg/h"
	"github.com/aminpaks/go-streams/pkg/mw"
	"github.com/aminpaks/go-streams/pkg/re"
	"github.com/aminpaks/go-streams/pkg/users"
)

func New(port string) {
	router := chi.NewRouter()

	router.Use(mw.RequestLog)
	router.Use(mw.Logger)
	router.Use(mw.Recoverer)

	router.Route("/users", func(r chi.Router) {
		err := users.NewUserController(r)
		if err != nil {
			log.Fatalf("failed to initialize user controller: %v", err)
		}
	})

	router.NotFound(h.NewH(func(rw http.ResponseWriter, r *http.Request) h.Renderer {
		return re.Json(http.StatusNotFound, re.BuildJsonErrors(errors.New("not found")))
	}))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
