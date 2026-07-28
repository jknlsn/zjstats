package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jknlsn/zjstat/internal/config"
	"github.com/jknlsn/zjstat/internal/format"
	"github.com/jknlsn/zjstat/internal/metrics"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "zjstat: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	sockPath := filepath.Join(cacheDir, "zjstatd.sock")

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 200 * time.Millisecond,
	}

	resp, err := client.Get("http://unix/metrics")
	if err != nil {
		return fmt.Errorf("daemon not responding (is zjstatd running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("daemon returned %s: %s", resp.Status, string(body))
	}

	var snap metrics.Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return fmt.Errorf("decode metrics: %w", err)
	}

	_, err = fmt.Fprint(os.Stdout, format.Snapshot(&snap, cfg))
	return err
}
