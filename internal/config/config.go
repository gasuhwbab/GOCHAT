package config

import (
	"time"
)

type Config struct {
	Host         string
	Port         int
	IdleTimeout  time.Duration
	WriteTimeout time.Duration
	MsgMaxBytes  int
	MaxClients   int
	LogDebug     bool
}
