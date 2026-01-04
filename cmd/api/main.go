package main

import (
	"os"

	"github.com/patrickjm/api-cli/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
