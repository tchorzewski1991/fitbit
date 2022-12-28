package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
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

	return nil
}
