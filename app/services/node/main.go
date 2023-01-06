package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/tchorzewski1991/fitbit/core/blockchain/genesis"
	"github.com/tchorzewski1991/fitbit/core/logger"
	"go.uber.org/zap"
)

const (
	service = "fitbit"
	prefix  = "NODE"
)

var build = "develop"

func main() {
	log, err := logger.New(service, build)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync() // nolint:errcheck

	if err = run(log); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// ========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Api struct {
			Host            string        `conf:"default:0.0.0.0:3000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			IdleTimeout     time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
	}{
		Version: conf.Version{
			Build: build,
		},
	}

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config err: %w", err)
	}

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config output err: %w", err)
	}
	log.Infow("config parsed", "config", out)

	// ========================================================================
	// Setup blockchain components

	_, err = genesis.Load()
	if err != nil {
		return fmt.Errorf("loading genesis file err: %w", err)
	}

	// ========================================================================
	// Starting node API

	api := http.Server{
		Addr:         cfg.Api.Host,
		Handler:      nil,
		ReadTimeout:  cfg.Api.ReadTimeout,
		WriteTimeout: cfg.Api.WriteTimeout,
		IdleTimeout:  cfg.Api.IdleTimeout,
	}

	errorCh := make(chan error, 1)
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Infow("starting node")
		defer log.Infow("node stopped")

		if err = api.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorCh <- err
		}
	}()

	select {
	case sig := <-shutdownCh:
		log.Infow("starting shutdown", "signal", sig)
		defer log.Infow("shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Api.ShutdownTimeout)
		defer cancel()

		if err = api.Shutdown(ctx); err != nil {
			_ = api.Close()
			return fmt.Errorf("cannot shutdown node gracefully: %w", err)
		}
	case err = <-errorCh:
		return fmt.Errorf("node err: %w", err)
	}

	return nil
}