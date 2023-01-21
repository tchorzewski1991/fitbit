package cmd

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Generates public address of the account",
	Run:   addressRun,
}

func init() {
	rootCmd.AddCommand(addressCmd)
}

func addressRun(_ *cobra.Command, _ []string) {
	accountLocation, err := generateAccountLocation()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	priv, err := crypto.LoadECDSA(accountLocation)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	address, err := database.PubToAccountID(priv.PublicKey)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(address)
}
