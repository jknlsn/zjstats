package format

import (
	"strings"
	"testing"

	"github.com/jknlsn/zjstat/internal/config"
	"github.com/jknlsn/zjstat/internal/metrics"
)

func TestSnapshotZjd(t *testing.T) {
	cfg := &config.Config{
		Theme: config.Theme{
			Background: "$surface0",
			Label:      "$blue,bold",
			Text:       "$text",
			Alert:      "$red,bold",
			Warn:       "$yellow,bold",
		},
		Metrics: []config.MetricConfig{
			{Name: "cpu", Label: "cpu", Format: "%2.0f%%"},
			{Name: "gpu", Label: "gpu", Format: "%2.0f%%", HideIfMissing: true},
			{Name: "memory", Label: "mem", Format: "%2.0f%%", WarnAt: 80},
		},
	}
	snap := &metrics.Snapshot{CPU: 25, GPU: -1, Memory: 73}

	got := SnapshotZjd(snap, cfg)
	want := "#[fg=4,bold]cpu: #[fg=7]25% #[fg=4,bold]mem: #[fg=7]73%"
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}

	// Warn threshold flips the label colour; unknown theme names drop the
	// colour but keep modifiers; the background theme is ignored (the bar
	// renders on the plain terminal background).
	snap.Memory = 95
	cfg.Theme.Background = "$surface0"
	cfg.Theme.Label = "$notacolor,bold"
	got = SnapshotZjd(snap, cfg)
	want = "#[bold]cpu: #[fg=7]25% #[fg=3,bold]mem: #[fg=7]95%"
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// Three-digit values drop the pad space after the colon so every metric
// keeps a constant rendered width and the bar never shifts.
func TestSnapshotZjdConstantWidth(t *testing.T) {
	cfg := &config.Config{
		Theme:   config.Theme{Label: "$blue,bold", Text: "$text"},
		Metrics: []config.MetricConfig{{Name: "cpu", Label: "cpu", Format: "%2.0f%%"}},
	}
	snap := &metrics.Snapshot{CPU: 45}
	narrow := SnapshotZjd(snap, cfg)
	snap.CPU = 100
	wide := SnapshotZjd(snap, cfg)

	if len(narrow) != len(wide) {
		t.Fatalf("width changed: %d (%q) vs %d (%q)", len(narrow), narrow, len(wide), wide)
	}
	// End to end: label colon followed directly by the value token — no
	// pad space at 100%%.
	if !strings.Contains(wide, "cpu:#") {
		t.Fatalf("expected pad space dropped at 100%%, got %q", wide)
	}
	for _, v := range []string{"100%", "999%"} {
		if got := labelSep(v); got != ":" {
			t.Fatalf("labelSep(%q) = %q, want %q", v, got, ":")
		}
	}
	for _, v := range []string{"45%", " 5%", "--%"} {
		if got := labelSep(v); got != ": " {
			t.Fatalf("labelSep(%q) = %q, want %q", v, got, ": ")
		}
	}
}
