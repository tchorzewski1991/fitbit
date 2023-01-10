package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	accountName string
	accountPath string
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "FitbitCLI",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateAccountLocation() (string, error) {

	if strings.TrimSpace(accountName) == "" {
		return "", errors.New("account-name cannot be empty")
	}

	if strings.TrimSpace(accountPath) == "" {
		return "", errors.New("account-path cannot be empty")
	}

	if !strings.HasSuffix(accountName, keyExtension) {
		accountName += keyExtension
	}

	err := os.MkdirAll(accountPath, 0750)
	if err != nil {
		return "", err
	}

	return path.Join(accountPath, accountName), nil
}
