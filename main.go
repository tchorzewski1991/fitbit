package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stderr, "[fit-ledger] ", log.Lmicroseconds)

	if err := run(logger); err != nil {
		logger.Fatalln(err)
	}
}

func run(logger *log.Logger) error {
	logger.Println("service start")
	defer logger.Println("service end")
	return nil
}
