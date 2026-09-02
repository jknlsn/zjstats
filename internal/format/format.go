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
		b.WriteString(labelSep(valStr))
		b.WriteString(bg(cfg.Theme.Background))
		b.WriteString(fg(cfg.Theme.Text))
		b.WriteString(valStr)
		b.WriteByte(' ')
	}

	return strings.TrimRight(b.String(), " ")
}

// labelSep joins a label to its rendered value. Percent formats pad to
// two digits ("%2.0f%%"), so a 3-digit value ("100%") renders one char
// wider; dropping the pad space there keeps every metric a constant
// width and the bar never shifts when a value crosses 100.
func labelSep(valStr string) string {
	if len(valStr) >= 4 {
		return ":"
	}
	return ": "
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

// SnapshotZjd renders with zjd's widget markup instead of
// zjstatus's: tokens carry concrete ANSI indices (themed by the terminal
// palette at render time) and are complete styles. No background tokens —
// zjd's bar renders on the plain terminal background, so a bg chip
// (ANSI 8 is not surface0 in most palettes) reads as a dim box. $-theme
// references resolve through ansiIndex; unknown names drop the colour and
// keep any modifiers.
func SnapshotZjd(s *metrics.Snapshot, cfg *config.Config) string {
	var b strings.Builder
	for _, m := range cfg.Metrics {
		valStr, val, ok := resolveMetric(s, m)
		if !ok {
			if m.HideIfMissing {
				continue
			}
			valStr = "--%"
		}

		b.WriteString(token(labelColor(val, m, cfg.Theme)))
		b.WriteString(m.Label)
		b.WriteString(labelSep(valStr))
		b.WriteString(token(cfg.Theme.Text))
		b.WriteString(valStr)
		b.WriteByte(' ')
	}
	return strings.TrimRight(b.String(), " ")
}

// token builds one zjd style token: `#[fg=F]` with `,bold`/`,dim`
// modifiers carried from the theme entry; empty when nothing applies.
func token(spec string) string {
	name, mods := splitSpec(spec)
	var attrs []string
	if idx := ansiIndex(name); idx != "" {
		attrs = append(attrs, "fg="+idx)
	}
	for _, mod := range mods {
		if mod == "bold" || mod == "dim" {
			attrs = append(attrs, mod)
		}
	}
	if len(attrs) == 0 {
		return ""
	}
	return "#[" + strings.Join(attrs, ",") + "]"
}

// splitSpec splits a theme entry like "$blue,bold" into its colour name
// and remaining modifier words.
func splitSpec(spec string) (string, []string) {
	parts := strings.Split(strings.TrimSpace(spec), ",")
	return strings.TrimPrefix(parts[0], "$"), parts[1:]
}

// ansiIndex maps a zjstatus theme colour name to the ANSI index zjd
// renders (terminal-palette themed). Unknown names return "".
func ansiIndex(name string) string {
	switch name {
	case "red", "maroon", "peach", "flamingo":
		return "1"
	case "green":
		return "2"
	case "yellow":
		return "3"
	case "blue":
		return "4"
	case "mauve", "magenta", "pink", "lavender":
		return "5"
	case "cyan", "teal", "sky", "sapphire":
		return "6"
	case "text", "subtext0", "subtext1", "white", "rosewater":
		return "7"
	case "surface0", "surface1", "surface2", "overlay0", "overlay1", "overlay2":
		return "8"
	case "base", "mantle", "crust", "black":
		return "0"
	default:
		return ""
	}
}
