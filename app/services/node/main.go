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
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers"
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
		Node struct {
			PublicHost      string        `conf:"default:0.0.0.0:3000"`
			PrivateHost     string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			IdleTimeout     time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		State struct {
			AccountsPath string `conf:"default:data/accounts"`
			DataPath     string `conf:"default:data/miner"`
			Beneficiary  string `conf:"default:test"`
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

	// Prepare beneficiary private key location
	beneficiaryPrivLocation := fmt.Sprintf("%s/%s.ecdsa", cfg.State.AccountsPath, cfg.State.Beneficiary)

	// Load beneficiary private key
	priv, err := crypto.LoadECDSA(beneficiaryPrivLocation)
	if err != nil {
		return fmt.Errorf("loading beneficiary private key err: %w", err)
	}

	// Build beneficiary address out of private - public key pair
	beneficiaryID, err := database.PubToAccountID(priv.PublicKey)
	if err != nil {
		return fmt.Errorf("loading beneficiary account ID err: %w", err)
	}

	gen, err := genesis.Load()
	if err != nil {
		return fmt.Errorf("loading genesis file err: %w", err)
	}

	// ========================================================================
	// Starting public node

	nodeErrors := make(chan error, 1)
	nodeShutdown := make(chan os.Signal, 1)
	signal.Notify(nodeShutdown, syscall.SIGTERM, syscall.SIGINT)

	publicMux := handlers.PublicMux(handlers.Config{
		Log: log,
	})

	publicNode := http.Server{
		Addr:         cfg.Node.PublicHost,
		Handler:      publicMux,
		ReadTimeout:  cfg.Node.ReadTimeout,
		WriteTimeout: cfg.Node.WriteTimeout,
		IdleTimeout:  cfg.Node.IdleTimeout,
	}

	go func() {
		log.Infow("starting public node", "host", publicNode.Addr)
		defer log.Infow("public node stopped")

		if err = publicNode.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			nodeErrors <- err
		}
	}()

	select {
	case sig := <-nodeShutdown:
		log.Infow("starting shutdown", "signal", sig)
		defer log.Infow("shutdown complete", "signal", sig)

		ctx, cancelPub := context.WithTimeout(context.Background(), cfg.Node.ShutdownTimeout)
		defer cancelPub()

		log.Infow("shutting down public node")
		if err = publicNode.Shutdown(ctx); err != nil {
			_ = publicNode.Close()
			return fmt.Errorf("cannot shutdown public node gracefully: %w", err)
		}
	case err = <-nodeErrors:
		return fmt.Errorf("node err: %w", err)
	}

	return nil
}
