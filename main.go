package main

import (
	"os"

	"github.com/kacheo/devlog/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
