package metrics

import "time"

type MemoryInfo struct {
	TotalBytes     uint64 `json:"total_bytes"`
	AvailableBytes uint64 `json:"available_bytes"`
	UsedBytes      uint64 `json:"used_bytes"`
}

type UptimeInfo struct {
	Seconds uint64 `json:"seconds"`
}

type DiskInfo struct {
	TotalBytes     uint64 `json:"total_bytes"`
	AvailableBytes uint64 `json:"available_bytes"`
	UsedBytes      uint64 `json:"used_bytes"`
}

type Status struct {
	Memory      MemoryInfo `json:"memory"`
	Uptime      UptimeInfo `json:"uptime"`
	CollectedAt time.Time  `json:"collected_at"`
	Disk        DiskInfo   `json:"disk"`
	CPU         CPUInfo    `json:"cpu"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
}

type cpuTimes struct {
	total uint64
	idle  uint64
}
