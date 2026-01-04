package main

import (
	"os"

	"github.com/patrickmoriarty/api-cli/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
