package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"watchd/internal/config"
	"watchd/internal/daemon"
)

func main() {
	configPath := flag.String("config", "/etc/watchd/config.yaml", "Path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.SetOutput(os.Stdout)
	log.Println("watchd started")

	monitor := daemon.New(cfg)
	if err := monitor.Run(ctx); err != nil {
		log.Fatalf("watchd stopped with error: %v", err)
	}

	log.Println("watchd stopped")
}
