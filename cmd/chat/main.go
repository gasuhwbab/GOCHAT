package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gasuhwbab/chat_server/internal/config"
	"github.com/gasuhwbab/chat_server/internal/server"
)

func main() {
	var (
		host         = flag.String("host", "localhost", "Host listening on")
		port         = flag.Int("port", 8080, "Port listening on")
		idleTimeout  = flag.Duration("idle-timeout", 120*time.Second, "Idle read timeout per client")
		writeTimeout = flag.Duration("write-timeout", 10*time.Second, "Write timeout per message")
		msgMaxBytes  = flag.Int("msg-max-bytes", 4096, "Max message size in bytes")
		maxClient    = flag.Int("max-clients", 0, "Max number of clients (0-unlimited)")
		debug        = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	cfg := config.Config{
		Host:         *host,
		Port:         *port,
		IdleTimeout:  *idleTimeout,
		WriteTimeout: *writeTimeout,
		MsgMaxBytes:  *msgMaxBytes,
		MaxClients:   *maxClient,
		LogDebug:     *debug,
	}

	serv := server.NewServer(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := serv.Run(ctx); err != nil {
		log.Fatal("Server stopped", err.Error())
	}
}
