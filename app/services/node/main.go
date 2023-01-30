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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tchorzewski1991/fitbit/app/services/node/handlers"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
	"github.com/tchorzewski1991/fitbit/core/blockchain/genesis"
	"github.com/tchorzewski1991/fitbit/core/blockchain/network"
	"github.com/tchorzewski1991/fitbit/core/blockchain/state"
	"github.com/tchorzewski1991/fitbit/core/blockchain/storage/disk"
	"github.com/tchorzewski1991/fitbit/core/blockchain/worker"
	"github.com/tchorzewski1991/fitbit/core/logger"
	"github.com/tchorzewski1991/fitbit/core/nameservice"
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
	defer func() {
		_ = log.Sync()
	}()

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
			AccountsPath string   `conf:"default:data/accounts"`
			DataPath     string   `conf:"default:data/miner"`
			Beneficiary  string   `conf:"default:test"`
			OriginPeers  []string `conf:"default:0.0.0.0:4000"`
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

	// Prepare collection of known peers.
	knownPeers := network.NewPeerSet()
	for _, peer := range cfg.State.OriginPeers {
		knownPeers.Add(network.NewPeer(peer))
	}
	knownPeers.Add(network.NewPeer(cfg.Node.PrivateHost))

	ns, err := nameservice.New(cfg.State.AccountsPath)
	if err != nil {
		return fmt.Errorf("loading nameservice err: %w", err)
	}

	// Prepare location of beneficiary private key
	beneficiaryPrivLocation := fmt.Sprintf("%s/%s.ecdsa", cfg.State.AccountsPath, cfg.State.Beneficiary)

	// Load beneficiary private key
	priv, err := crypto.LoadECDSA(beneficiaryPrivLocation)
	if err != nil {
		return fmt.Errorf("loading beneficiary private key err: %w", err)
	}

	// Prepare beneficiary address out of public - private key pair
	beneficiaryID, err := database.PubToAccountID(priv.PublicKey)
	if err != nil {
		return fmt.Errorf("loading beneficiary account ID err: %w", err)
	}

	gen, err := genesis.Load()
	if err != nil {
		return fmt.Errorf("loading genesis file err: %w", err)
	}

	storage, err := disk.New(cfg.State.DataPath)
	if err != nil {
		return fmt.Errorf("loading disk storage err: %w", err)
	}

	eventHandler := func(s string, args ...any) {
		log.Infow(fmt.Sprintf(s, args...))
	}

	// Load core API through state abstraction.
	s, err := state.New(state.Config{
		BeneficiaryID: beneficiaryID,
		Host:          cfg.Node.PrivateHost,
		Genesis:       gen,
		Storage:       storage,
		EventHandler:  eventHandler,
		KnownPeers:    knownPeers,
	})
	if err != nil {
		return fmt.Errorf("loading state err: %w", err)
	}
	defer func() {
		_ = s.Shutdown()
	}()

	worker.Run(s, eventHandler)

	// ========================================================================
	// Starting nodes

	nodeErrors := make(chan error, 1)
	nodeShutdown := make(chan os.Signal, 1)
	signal.Notify(nodeShutdown, syscall.SIGTERM, syscall.SIGINT)

	// ========================================================================
	// Starting public node

	publicMux := handlers.PublicMux(handlers.Config{
		Log:         log,
		State:       s,
		NameService: ns,
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

	// ========================================================================
	// Starting private node

	privateMux := handlers.PrivateMux(handlers.Config{
		Log:         log,
		State:       s,
		NameService: ns,
	})

	privateNode := http.Server{
		Addr:         cfg.Node.PrivateHost,
		Handler:      privateMux,
		ReadTimeout:  cfg.Node.ReadTimeout,
		WriteTimeout: cfg.Node.WriteTimeout,
		IdleTimeout:  cfg.Node.IdleTimeout,
	}

	go func() {
		log.Infow("starting private node", "host", privateNode.Addr)
		defer log.Infow("private node stopped")

		if err = privateNode.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			nodeErrors <- err
		}
	}()

	// ========================================================================
	// Handle shutdown signal gracefully

	select {
	case sig := <-nodeShutdown:
		log.Infow("starting shutdown", "signal", sig)
		defer log.Infow("shutdown complete", "signal", sig)

		ctx, cancelPub := context.WithTimeout(context.Background(), cfg.Node.ShutdownTimeout)
		defer cancelPub()

		log.Infow("shutdown public node")
		if err = publicNode.Shutdown(ctx); err != nil {
			_ = publicNode.Close()
			return fmt.Errorf("cannot shutdown public node gracefully: %w", err)
		}

		ctx, cancelPriv := context.WithTimeout(context.Background(), cfg.Node.ShutdownTimeout)
		defer cancelPriv()

		log.Infow("shutdown private node")
		if err = privateNode.Shutdown(ctx); err != nil {
			_ = privateNode.Close()
			return fmt.Errorf("cannot shutdown private node gracefully: %w", err)
		}
	case err = <-nodeErrors:
		return fmt.Errorf("node err: %w", err)
	}

	return nil
}
