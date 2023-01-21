package cmd

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates new private key for the account",
	Run:   generateRun,
}

func init() {
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
