package main

import (
	"log"
	"os"
	"strings"
	"time"
)

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		log.Printf("invalid %s=%q, using fallback %s", key, raw, fallback)
		return fallback
	}

	return d
}
