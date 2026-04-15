package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/daemon"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "path to config file (JSON)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("portwatch %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("failed to create daemon: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("portwatch %s starting (interval: %s, ports: %d-%d)",
		version, cfg.Interval, cfg.PortRange.From, cfg.PortRange.To)

	go func() {
		if err := d.Run(); err != nil {
			log.Fatalf("daemon error: %v", err)
		}
	}()

	<-sigCh
	log.Println("shutting down")
	d.Stop()
}
