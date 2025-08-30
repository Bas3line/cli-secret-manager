package main

import (
	"log"

	"sm-cli/pkg/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("cli exited with error: %v", err)
	}
}
