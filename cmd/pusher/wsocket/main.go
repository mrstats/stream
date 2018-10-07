package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"stream/internal/app/pusher/wsocket"
	"stream/internal/pkg/privlib/config"
	"stream/internal/pkg/privlib/logger"
)

var (
	log = logger.GetInstance()
	cfg = config.GetInstance()
)

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

func main() {
	hub := wsocket.NewHub()
	go hub.Run()

	ws := wsocket.NewWorker(hub.GetBroadcastChan())
	go ws.Run()

	hub.SetFallowChan(ws.GetFollowChan())
	r := mux.NewRouter()
	s := r.
		Methods(http.MethodGet).
		Subrouter()

	index, err := ioutil.ReadFile(string("./internal/app/pusher/wsocket/web/index.html"))
	if err != nil {
		log.WithError(err).Error("index.html can not be read")
	}
	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})

	s.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsocket.ServeWS(hub, w, r)
	}).Queries("token", "{token}")

	svr := &http.Server{
		Addr:         cfg.GetString("pusher.http.addr"),
		ReadTimeout:  time.Second * cfg.GetDuration("pusher.http.timeout.read"),
		WriteTimeout: time.Second * cfg.GetDuration("pusher.http.timeout.write"),
		Handler:      r,
	}

	go onSignal(func() {
		shutdownTimeout := cfg.GetDuration("pusher.http.timeout.shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()
		// TODO: stop worker
		err := svr.Shutdown(ctx)
		if err != nil {
			log.WithError(err).Errorf("Failed to shutdown gracefully after %ds", shutdownTimeout)
		} else {
			log.Debugf("succeeded to shutdown gracefully")
			svr = nil
		}
	})

	log.Infof("open http://%s (note: some browsers doesn't allow insecure websocket connections to localhost)", svr.Addr)
	err = svr.ListenAndServe()
	if err != http.ErrServerClosed {
		log.WithError(err).Error("Http Server stopped unexpected")
		//q.Shutdown() // TODO: stop worker
	} else {
		log.WithError(err).Info("Http Server stopped")
	}
}
