package format

import (
	"fmt"
	"strings"

	"github.com/jknlsn/zjstat/internal/config"
	"github.com/jknlsn/zjstat/internal/metrics"
)

// Snapshot renders a metrics snapshot using the provided configuration.
func Snapshot(s *metrics.Snapshot, cfg *config.Config) string {
	var b strings.Builder

	for _, m := range cfg.Metrics {
		valStr, val, ok := resolveMetric(s, m)
		if !ok {
			if m.HideIfMissing {
				continue
			}
			valStr = "--%"
		}

		b.WriteString(bg(cfg.Theme.Background))
		b.WriteString(fg(labelColor(val, m, cfg.Theme)))
		b.WriteString(m.Label)
		b.WriteString(": ")
		b.WriteString(bg(cfg.Theme.Background))
		b.WriteString(fg(cfg.Theme.Text))
		b.WriteString(valStr)
		b.WriteByte(' ')
	}

	return strings.TrimRight(b.String(), " ")
}

func resolveMetric(s *metrics.Snapshot, m config.MetricConfig) (display string, value float64, ok bool) {
	switch m.Name {
	case "cpu":
		return fmt.Sprintf(m.Format, s.CPU), s.CPU, true
	case "gpu":
		if s.GPU < 0 {
			return "", 0, false
		}
		return fmt.Sprintf(m.Format, float64(s.GPU)), float64(s.GPU), true
	case "memory":
		return fmt.Sprintf(m.Format, s.Memory), s.Memory, true
	case "disk":
		for _, d := range s.Disks {
			if d.Mount == m.Mount {
				return fmt.Sprintf(m.Format, d.UsedPercent), d.UsedPercent, true
			}
		}
		return "", 0, false
	default:
		return "", 0, false
	}
}

func labelColor(val float64, m config.MetricConfig, t config.Theme) string {
	if m.AlertAt > 0 && val >= m.AlertAt {
		return t.Alert
	}
	if m.WarnAt > 0 && val >= m.WarnAt {
		return t.Warn
	}
	return t.Label
}

func bg(c string) string { return "#[bg=" + c + "]" }
func fg(c string) string { return "#[fg=" + c + "]" }
