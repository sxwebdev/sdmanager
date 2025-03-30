package main

import (
	"fmt"
	"os"

	"github.com/sxwebdev/sdmanager"
)

var (
	version    = "dev"
	commitHash = "none"
	buildDate  = "none"
)

func PrintVersion() {
	fmt.Printf("Systemd Manager %s\n", version)
	fmt.Printf("Commit hash: %s\n", commitHash)
	fmt.Printf("Build date: %s\n", buildDate)
}

func main() {
	err := sdmanager.RunSystemdManager()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
