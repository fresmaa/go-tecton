package main

import (
	"fmt"
	"os"

	"github.com/fresmaa/go-tecton/cmd/tecton/cli"
)

func main() {
	// Initialize and execute the root CLI command
	if err := cli.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
