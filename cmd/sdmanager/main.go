package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	defer cancel()

	err := sdmanager.RunSystemdManager(sdmanager.WithContext(ctx))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
