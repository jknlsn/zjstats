package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/jknlsn/zjstat/internal/collector"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("zjstatd: %v", err)
	}
}

func run() error {
	coll := collector.New()

	// Seed first snapshot immediately.
	snap, err := coll.Collect()
	if err != nil {
		return fmt.Errorf("initial collect: %w", err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}
	sockPath := filepath.Join(cacheDir, "zjstatd.sock")
	_ = os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("listen %s: %w", sockPath, err)
	}
	defer ln.Close()

	var mu sync.Mutex
	latest := snap

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		s := latest
		mu.Unlock()
		if s == nil {
			http.Error(w, `{"error":"no data"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s)
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	// Update loop.
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("zjstatd listening on %s", sockPath)

	for {
		select {
		case <-ticker.C:
			snap, err := coll.Collect()
			if err != nil {
				log.Printf("collect error: %v", err)
				continue
			}
			mu.Lock()
			latest = snap
			mu.Unlock()

		case sig := <-sigCh:
			log.Printf("received %s, shutting down", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
			return nil
		}
	}
}
