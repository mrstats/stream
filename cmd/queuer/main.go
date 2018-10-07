package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stream/internal/app/queuer"
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
	q := queuer.NewQueuer()

	svr := &http.Server{
		Addr:         cfg.GetString("queuer.http.addr"),
		ReadTimeout:  time.Second * cfg.GetDuration("queuer.http.timeout.read"),
		WriteTimeout: time.Second * cfg.GetDuration("queuer.http.timeout.write"),
		Handler:      q.Route(),
	}

	go onSignal(func() {
		shutdownTimeout := cfg.GetDuration("queuer.http.timeout.shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		err := svr.Shutdown(ctx)
		if err != nil {
			log.WithError(err).Errorf("Failed to shutdown Queuer gracefully after %ds", shutdownTimeout)
		} else {
			log.Debugf("succeeded to shutdown Queuer gracefully")
			svr = nil
		}

		q.Shutdown()
	})

	log.WithTime(time.Now()).Infof("Queuer Server listen and queue at %s", svr.Addr)
	err := svr.ListenAndServe()
	if err != http.ErrServerClosed {
		log.WithError(err).Error("Http Server stopped unexpected")
		q.Shutdown()
	} else {
		log.WithError(err).Info("Http Server stopped")
	}
}
