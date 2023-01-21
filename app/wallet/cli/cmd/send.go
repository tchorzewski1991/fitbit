package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"github.com/tchorzewski1991/fitbit/core/blockchain/database"
)

var (
	url   string
	to    string
	nonce uint64
	value uint64
	tip   uint64
	data  []byte
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends a new transaction",
	Run:   sendRun,
}

func init() {
	sendCmd.Flags().StringVarP(
		&url,
		"url", "u",
		"http://localhost:3000",
		"The url of the public node.",
	)
	sendCmd.Flags().StringVarP(
		&to,
		"to", "t",
		"",
		"The receiver of the transaction.",
	)
	sendCmd.Flags().Uint64VarP(
		&nonce,
		"nonce", "n",
		0,
		"Transaction id to send.",
	)
	sendCmd.Flags().Uint64VarP(
		&value,
		"value", "v",
		0,
		"Transaction value to send.",
	)
	sendCmd.Flags().Uint64VarP(
		&tip,
		"tip", "c",
		0,
		"Transaction tip to add.",
	)
	sendCmd.Flags().BytesHexVarP(
		&data,
		"data", "d",
		nil,
		"Transaction data to send.",
	)
	rootCmd.AddCommand(sendCmd)
}

func sendRun(_ *cobra.Command, _ []string) {
	accountLocation, err := generateAccountLocation()
	if err != nil {
		fmt.Println(fmt.Errorf("generate account location err: %w", err))
		os.Exit(1)
	}

	priv, err := crypto.LoadECDSA(accountLocation)
	if err != nil {
		fmt.Println(fmt.Errorf("load priv key err: %w", err))
		os.Exit(1)
	}

	fromAddress, err := database.PubToAccountID(priv.PublicKey)
	if err != nil {
		fmt.Println(fmt.Errorf("set sender address err: %w", err))
		os.Exit(1)
	}

	toAddress, err := database.ToAccountID(to)
	if err != nil {
		fmt.Println(fmt.Errorf("set reciver address err: %w", err))
		os.Exit(1)
	}

	if nonce < 0 {
		nonce = defaultNonce
	}

	const chainID = 1
	tx := database.Tx{
		ChainID: chainID,
		Nonce:   nonce,
		From:    fromAddress,
		To:      toAddress,
		Value:   value,
		Tip:     tip,
	}

	signedTx, err := tx.Sign(priv)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	payload, err := json.Marshal(signedTx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := http.Post(fmt.Sprintf("%s/v1/tx/submit", url), "application/json", bytes.NewBuffer(payload))
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		err = fmt.Errorf("post request err: %w", err)
		fmt.Println(err)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read response body err: %w", err)
		fmt.Println(err)
		os.Exit(1)
	}

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("request failed | status: %d | body: %s", resp.StatusCode, string(body))
		fmt.Println(err)
		os.Exit(1)
	}

	msg := fmt.Sprintf("request success | status: %d | body: %s", resp.StatusCode, string(body))
	fmt.Println(msg)
}
