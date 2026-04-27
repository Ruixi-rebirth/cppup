package main

import (
	"os"

	"github.com/Ruixi-rebirth/cppup/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
