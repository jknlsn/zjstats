# zjstat

A fast, daemon-based system metrics collector for [Zellij](https://zellij.dev/) status bars (inspired by [tmstat-rs](https://github.com/playbahn/tmstat-rs)).

Instead of spawning multiple shell processes every few seconds, `zjstatd` runs continuously in the background, samples metrics once per second, and serves raw JSON over a Unix domain socket. The `zjstat` client reads a user-supplied configuration file, queries the daemon, and renders a zjstatus-compatible format string.

## Architecture

```
┌─────────────┐     Unix socket      ┌─────────────┐
│  zjstatd    │ ◄─────────────────── │   zjstat    │
│  (daemon)   │   HTTP /metrics      │   (client)  │
│             │   raw JSON           │  formats    │
│  samples    │                      │  config     │
│  CPU/GPU/   │                      │  driven     │
│  mem/disk   │                      │             │
└─────────────┘                      └─────────────┘
```

**Separation of concerns**

- **zjstatd** knows *nothing* about colours, layout, or zjstatus. It only collects and serves data.
- **zjstat** knows *nothing* about `host_statistics`, IOKit, or `statfs`. It only reads config, fetches JSON, and prints text.
- This makes the daemon tiny, testable, and reusable by other tools.

**Why a daemon?**

- **Accurate CPU %** requires delta computation between two samples. A daemon maintains continuous state instead of blocking for `iostat`.
- **Sub-millisecond latency** for status updates.
- **Extensible**: add network throughput, battery, thermal throttling, per-process stats without changing the zellij layout.
- **Lightweight**: typically <10 MB RSS and <0.1 % CPU on Apple Silicon.

## Install

```bash
cd /path/to/zjstat
go build ./cmd/zjstatd
go build ./cmd/zjstat

# Or install to $GOPATH/bin
go install ./cmd/zjstatd
go install ./cmd/zjstat
```

Go builds are ad-hoc signed, and **pro-one's** launchd Launch Constraint
rejects ad-hoc-signed binaries. Simplest policy: sign once and copy that
same binary to every host — a properly signed build runs everywhere, only
pro-one strictly requires it. Re-signing replaces the running binary, so
rebuild → sign → swap (never over a live inode) → restart the agent there:

```bash
codesign -s "Apple Development: Jake Nelson (9475H48R63)" ~/.config/zellij/bin/zjstatd
launchctl kickstart -k gui/$(id -u)/dev.zjstatd
```

## Usage

### 1. Start the daemon

```bash
zjstatd
```

It creates a Unix socket at `~/Library/Caches/zjstatd.sock` and logs to stderr.

### 2. Query manually

```bash
zjstat
# → cpu: 25% gpu: 63% mem: 73% ssd: 89% @local
```

For [den](../den) bar widgets, render with its markup instead —
concrete ANSI indices (themed by the terminal palette) instead of zjstatus
`$name` references, one complete-style token per segment:

```bash
zjstat --zjd
# → #[fg=4,bold]cpu: #[fg=7]25% #[fg=4,bold]gpu: …
```

No background tokens: den's bar renders on the plain terminal
background (an ANSI-8 bg reads as a dim box, not surface0).

Wire it as the host's `[bar.widget.stats]` in `~/.conf../den/config.toml`
(see den's `docs/config.md` § Cross-machine widgets). The same config —
metrics, labels, thresholds, `hide_if_missing` — drives both renderers.

### 3. Configuration

`zjstat` looks for `~/.config/zjstat/config.toml`. If it doesn't exist, it uses built-in defaults that match the original hard-coded layout.

Copy the included [`config.toml`](config.toml) to customise colours, labels, metrics, and context rules:

```bash
mkdir -p ~/.config/zjstat
cp config.toml ~/.config/zjstat/config.toml
```

Example snippet:

```toml
[theme]
background = "$surface0"
label = "$blue,bold"
text = "$text"

[[metrics]]
name = "cpu"
label = "cpu"
format = "%2.0f%%"

[[metrics]]
name = "gpu"
label = "gpu"
format = "%2.0f%%"
hide_if_missing = true

[[metrics]]
name = "disk"
label = "ext"
format = "%2.0f%%"
mount = "/Volumes/OWC"
hide_if_missing = true
```

### 4. Wire into zellij

Update your layout (`~/.config/zellij/layouts/default.kdl`):

```kdl
format_right "{command_status} #[bg=$surface0,fg=$peach,bold]{session} #[bg=$surface0,fg=$yellow,bold]@#(hostname -s) #[bg=$surface0,fg=$text,bold]{datetime}"

command_status_command "/Users/YOURNAME/.config/zellij/bin/zjstat"
command_status_format  "{stdout}"
command_status_interval "1"
command_status_rendermode "dynamic"
```

**Note:** If using [zjstatus](https://github.com/dj95/zjstatus), you **must** use `rendermode "dynamic"`. The output from `zjstat` contains inline zjstatus formatting directives (e.g. `#[fg=$blue,bold]`), and `dynamic` mode tells zjstatus to parse and render those directives correctly. `static` mode treats the output as plain text, which causes the status bar section to disappear.

Then remove all the old `command_cpu`, `command_gpu`, `command_mem`, etc. blocks.

### 5. Auto-start with launchd (optional)

Create `~/Library/LaunchAgents/dev.zjstatd.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>dev.zjstatd</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOURNAME/.config/zellij/bin/zjstatd</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/zjstatd.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/zjstatd.err</string>
</dict>
</plist>
```

Load it:

```bash
launchctl load ~/Library/LaunchAgents/dev.zjstatd.plist
```

## Metrics exposed

The daemon speaks HTTP over a Unix socket. Only one endpoint:

- `GET /metrics` — returns a JSON `Snapshot`:
  ```json
  {
    "cpu": 12.5,
    "gpu": 44,
    "memory": 71.0,
    "disks": [
      {"mount": "/", "used_percent": 89.0},
      {"mount": "/Volumes/OWC", "used_percent": 23.0}
    ],
    "context": {
      "hostname": "myhost",
      "current_user": "alice",
      "preferred_user": "alice",
      "ssh_tty": false
    }
  }
  ```

| Metric | Source | macOS API |
|--------|--------|-----------|
| CPU % | Delta of total / idle ticks | `host_statistics` via gopsutil |
| GPU % | Apple Silicon utilization | IOKit `AGXAccelerator` |
| Memory % | Used / total physical | `host_statistics64` via gopsutil |
| Disk % | Per-mount usage | `statfs` via gopsutil |
| Context | SSH / user / hostname | `SSH_TTY`, `os/user`, `os.Hostname` (exposed in JSON, not rendered) |

## Project layout

```
zjstat/
├── cmd/
│   ├── zjstatd/          # daemon — collects & serves raw JSON
│   └── zjstat/           # client — reads config, queries daemon, formats
├── internal/
│   ├── config/           # TOML config parser + defaults
│   ├── format/           # zjstatus string formatter (config-driven)
│   ├── metrics/          # Snapshot structs
│   └── collector/        # macOS metric collection (gopsutil + IOKit)
├── config.toml           # sample configuration
├── dev.zjstatd.plist     # launchd template (copy to ~/Library/LaunchAgents/)
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Example setup

The [`example/`](example/) directory contains a complete Zellij config for a quickstart:

| File | What it is |
|------|------------|
| `config.kdl` | Zellij keybindings, theme, and settings |
| `layout.kdl` | Tab template with zjstatus and zjstat wired in |
| `context.sh` | Context-aware SSH/local indicators (replaces the old `metric.sh`; cpu/gpu/mem/disk now come from `zjstatd`) |

### Before using

1. Replace **every** `/Users/YOUR_USERNAME/` path with your actual home path.
2. Install `zjstatus.wasm` from the [releases page](https://github.com/dj95/zjstatus/releases) to `~/.config/zellij/plugins/`.
3. Build and install `zjstat` and `zjstatd` as described above.
4. Copy `context.sh` to `~/.config/zellij/shell/` and make it executable.
5. Optionally set `ZJSTAT_PREFERRED_USER` in your shell rc if your primary user differs from `$(id -un)`.

## Future ideas

- [ ] Network throughput (delta of interface counters)
- [ ] Battery percentage / cycle count
- [ ] Thermal pressure state
- [ ] Top-N CPU processes
- [ ] More context placeholders (e.g. `{session}`, `{datetime}`)
