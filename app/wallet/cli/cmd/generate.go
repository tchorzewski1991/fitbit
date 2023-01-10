package cmd

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

const (
	defaultAccountName = "private.ecdsa"
	defaultAccountPath = "data/accounts"
	keyExtension       = ".ecdsa"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates new private key for the account",
	Run:   generateRun,
}

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
