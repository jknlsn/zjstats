package collector

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/jknlsn/zjstat/internal/metrics"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type darwinCollector struct {
	prevCPUTimes []cpu.TimesStat
	prevCollect  time.Time
	preferredUser string
}

func newDarwinCollector() *darwinCollector {
	pref := preferredUser()
	return &darwinCollector{
		preferredUser: pref,
	}
}

func (c *darwinCollector) Collect() (*metrics.Snapshot, error) {
	now := time.Now()

	cpuPct, err := c.collectCPU()
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}

	memPct, err := c.collectMemory()
	if err != nil {
		return nil, fmt.Errorf("memory: %w", err)
	}

	gpuPct := collectGPU()

	disks, err := c.collectDisks()
	if err != nil {
		return nil, fmt.Errorf("disks: %w", err)
	}

	ctx := c.collectContext()

	c.prevCollect = now

	return &metrics.Snapshot{
		CPU:     cpuPct,
		GPU:     gpuPct,
		Memory:  memPct,
		Disks:   disks,
		Context: ctx,
	}, nil
}

func (c *darwinCollector) collectCPU() (float64, error) {
	times, err := cpu.Times(false) // per-cpu=false, total only
	if err != nil {
		return 0, err
	}
	if len(times) == 0 {
		return 0, fmt.Errorf("no cpu times returned")
	}

	if c.prevCPUTimes == nil || len(c.prevCPUTimes) == 0 {
		c.prevCPUTimes = times
		return 0, nil // first sample, no delta yet
	}

	prev := c.prevCPUTimes[0]
	curr := times[0]

	idleDelta := curr.Idle - prev.Idle
	totalDelta := (curr.User + curr.System + curr.Nice + curr.Iowait + curr.Irq + curr.Softirq + curr.Steal + curr.Guest + curr.GuestNice + curr.Idle) -
		(prev.User + prev.System + prev.Nice + prev.Iowait + prev.Irq + prev.Softirq + prev.Steal + prev.Guest + prev.GuestNice + prev.Idle)

	c.prevCPUTimes = times

	if totalDelta <= 0 {
		return 0, nil
	}

	used := 100.0 * (1.0 - idleDelta/totalDelta)
	if used < 0 {
		used = 0
	}
	if used > 100 {
		used = 100
	}
	return used, nil
}

func (c *darwinCollector) collectMemory() (float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return vm.UsedPercent, nil
}

func (c *darwinCollector) collectDisks() ([]metrics.Disk, error) {
	var out []metrics.Disk

	for _, mount := range []string{"/", "/Volumes/OWC"} {
		usage, err := disk.Usage(mount)
		if err != nil {
			// OWC may not be mounted; skip silently.
			continue
		}
		out = append(out, metrics.Disk{
			Mount:       mount,
			UsedPercent: usage.UsedPercent,
		})
	}

	return out, nil
}

func (c *darwinCollector) collectContext() metrics.Context {
	u, _ := user.Current()
	if u == nil {
		u = &user.User{Username: "?"}
	}
	host, _ := os.Hostname()
	return metrics.Context{
		Hostname:      strings.SplitN(host, ".", 2)[0],
		CurrentUser:   u.Username,
		PreferredUser: c.preferredUser,
		SSHTTY:        os.Getenv("SSH_TTY") != "",
	}
}

func preferredUser() string {
	// Derive from the install path: /Users/<name>/.config/...
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	ex, _ = filepath.EvalSymlinks(ex)
	if strings.HasPrefix(ex, "/Users/") {
		parts := strings.SplitN(strings.TrimPrefix(ex, "/Users/"), "/", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	// Fallback
	u, _ := user.Current()
	if u != nil {
		return u.Username
	}
	return ""
}
