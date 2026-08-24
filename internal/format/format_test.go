package format

import (
	"testing"

	"github.com/jknlsn/zjstat/internal/config"
	"github.com/jknlsn/zjstat/internal/metrics"
)

func TestSnapshotZjHerder(t *testing.T) {
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

	got := SnapshotZjHerder(snap, cfg)
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
	got = SnapshotZjHerder(snap, cfg)
	want = "#[bold]cpu: #[fg=7]25% #[fg=3,bold]mem: #[fg=7]95%"
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}}
