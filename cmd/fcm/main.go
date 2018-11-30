package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stream/internal/app/pusher/fcm"
	"stream/internal/pkg/privlib/config"
	"stream/internal/pkg/privlib/logger"

	"github.com/gorilla/mux"
)

var (
	log = logger.GetInstance()
	cfg = config.GetInstance()
)

func main() {
	svr := &http.Server{
		Addr:         cfg.GetString("fcm.http.addr"),
		ReadTimeout:  time.Second * cfg.GetDuration("fcm.http.timeout.read"),
		WriteTimeout: time.Second * cfg.GetDuration("fcm.http.timeout.write"),
		Handler:      route(fcm.RegisterClient, fcm.UnregisterClient),
	}

	go onSignal(func() {
		shutdownTimeout := cfg.GetDuration("fcm.http.timeout.shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()
		// TODO: stop worker
		fcm.UnsubscribeAll()

		err := svr.Shutdown(ctx)
		if err != nil {
			log.WithError(err).Errorf("Failed to shutdown gracefully after %ds", shutdownTimeout)
		} else {
			log.Debugf("succeeded to shutdown gracefully")
		}
	})

	log.Infof("server listning (http://%s) ...", svr.Addr)
	err := svr.ListenAndServe()
	if err != http.ErrServerClosed {
		log.WithError(err).Error("Http Server stopped unexpected")
		//q.Shutdown() // TODO: stop worker
	} else {
		log.WithError(err).Info("Http Server stopped")
	}
}

func route(register, unregister http.HandlerFunc) *mux.Router {
	r := mux.NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, r.Method+" method not allowed", http.StatusMethodNotAllowed)
	})
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad url", http.StatusNotFound)
	})

	api := r.
		PathPrefix("/api/v1.0/client").
		Subrouter()
	api.Methods(http.MethodPost).HandlerFunc(register)
	api.Methods(http.MethodDelete).HandlerFunc(unregister)

	r.
		PathPrefix("/").
		Handler(http.FileServer(http.Dir("internal/app/pusher/fcm/web"))).
		Methods(http.MethodGet)

	return r
}

func onSignal(shutdown func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	var signalsReceived uint

	for {
		select {
		case s := <-sig:
			log.Warnf("Signal received: %v", s)
			signalsReceived++

			if signalsReceived < 2 {
				log.Warnf("Waiting for running tasks to finish before shutting down")
				go shutdown()

				os.Exit(0)
			} else {
				os.Exit(0)
			}
		}
	}
}
