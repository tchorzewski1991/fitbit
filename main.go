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
	// ========================================================================
	// Starting app

	log.Infow("service start")
	defer log.Infow("service end")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config output err: %w", err)
	}
	log.Infow("config parsed", "config", out)

	return nil
}
