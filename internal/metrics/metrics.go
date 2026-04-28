package metrics

// Snapshot holds a point-in-time view of all collected metrics.
type Snapshot struct {
	CPU    float64 `json:"cpu"`    // percentage 0-100
	GPU    int     `json:"gpu"`    // percentage 0-100, -1 if unavailable
	Memory float64 `json:"memory"` // percentage 0-100

	Disks []Disk `json:"disks"` // per-mount usage

	Context Context `json:"context"` // ssh/user info
}

// Disk represents usage for a single mount point.
type Disk struct {
	Mount       string  `json:"mount"`
	UsedPercent float64 `json:"used_percent"`
}

// Context captures the runtime environment for conditional display.
type Context struct {
	Hostname      string `json:"hostname"`
	CurrentUser   string `json:"current_user"`
	PreferredUser string `json:"preferred_user"`
	SSHTTY        bool   `json:"ssh_tty"`
}
