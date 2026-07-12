package metrics

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

func CollectStatus() (Status, error) {
	mem, err := collectMemInfo()
	if err != nil {
		return Status{}, fmt.Errorf("collect memory: %w", err)
	}

	uptime, err := collectUptimeInfo()
	if err != nil {
		return Status{}, fmt.Errorf("collect uptime: %w", err)
	}

	disk, err := collectDiskInfo("/")
	if err != nil {
		return Status{}, fmt.Errorf("collect disk: %w", err)
	}

	cpu, err := collectCPUInfo()

	if err != nil {
		return Status{}, fmt.Errorf("collect cpu %w", err)
	}

	return Status{
		Memory:      mem,
		Uptime:      uptime,
		CollectedAt: time.Now().UTC(),
		Disk:        disk,
		CPU:         cpu,
	}, nil
}

func collectCPUInfo() (CPUInfo, error) {
	first, err := readCPUTimes()
	if err != nil {
		return CPUInfo{}, fmt.Errorf("read first CPU sample: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	second, err := readCPUTimes()
	if err != nil {
		return CPUInfo{}, fmt.Errorf("read second CPU sample: %w", err)
	}

	if second.total < first.total {
		return CPUInfo{}, errors.New("CPU total time decreased")
	}

	if second.idle < first.idle {
		return CPUInfo{}, errors.New("CPU idle time decreased")
	}

	totalDelta := second.total - first.total
	idleDelta := second.idle - first.idle

	if totalDelta == 0 {
		return CPUInfo{}, errors.New("CPU total delta is zero")
	}

	if idleDelta > totalDelta {
		return CPUInfo{}, errors.New("CPU idle delta exceeds total delta")
	}

	workDelta := totalDelta - idleDelta

	usagePercent := float64(workDelta) /
		float64(totalDelta) *
		100

	return CPUInfo{
		UsagePercent: usagePercent,
	}, nil
}

func collectMemInfo() (MemoryInfo, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}, err
	}
	return parseMemInfo(data)
}

func collectUptimeInfo() (UptimeInfo, error) {
	data, err := os.ReadFile("/proc/uptime")

	if err != nil {
		return UptimeInfo{}, err
	}

	return parseUptimeInfo(data)
}

func collectDiskInfo(path string) (DiskInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskInfo{}, err
	}

	blocksize := uint64(stat.Bsize)
	return DiskInfo{TotalBytes: blocksize * stat.Blocks, AvailableBytes: blocksize * stat.Bavail, UsedBytes: blocksize * (stat.Blocks - stat.Bfree)}, nil
}
