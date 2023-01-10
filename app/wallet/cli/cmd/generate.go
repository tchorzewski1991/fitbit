package cmd

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
)

const (
	defaultAccountName = "private.ecdsa"
	defaultAccountPath = "data/accounts"
	keyExtension       = ".ecdsa"
)

var (
	accountName string
	accountPath string
)

func init() {
	generateCmd.Flags().StringVarP(
		&accountName,
		"account-name", "a",
		defaultAccountName,
		"The name of the account.",
	)
	generateCmd.Flags().StringVarP(
		&accountPath,
		"account-path", "p",
		defaultAccountPath,
		"The path to the account private key.",
	)
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates new private key for the account",
	Run:   generateRun,
}

func generateRun(_ *cobra.Command, _ []string) {
	accountLocation, err := generateAccountLocation()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	priv, err := crypto.GenerateKey()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = crypto.SaveECDSA(accountLocation, priv)
	if err != nil {
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
