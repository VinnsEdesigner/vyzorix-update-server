package main

import (
	"log"

	"github.com/VinnsEdesigner/vyzorix/apps/api/tests"
)

// main runs all tests.
func main() {
	if err := tests.RunAllTests(); err != nil {
		log.Fatalf("Tests failed: %v", err)
	}
}
