package metrics

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readCPUTimes() (cpuTimes, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuTimes{}, fmt.Errorf("read /proc/stat: %w", err)
	}

	return parseCPUTimes(data)
}

func parseCPUTimes(data []byte) (cpuTimes, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "cpu ") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			return cpuTimes{}, fmt.Errorf(
				"invalid aggregate cpu line: %q",
				line,
			)
		}

		values := make([]uint64, 8)

		for i := range values {
			value, err := strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return cpuTimes{}, fmt.Errorf(
					"parse cpu field %d: %w",
					i+1,
					err,
				)
			}

			values[i] = value
		}

		user := values[0]
		nice := values[1]
		system := values[2]
		idle := values[3]
		iowait := values[4]
		irq := values[5]
		softirq := values[6]
		steal := values[7]

		idleAll := idle + iowait
		work := user + nice + system + irq + softirq + steal

		return cpuTimes{
			total: idleAll + work,
			idle:  idleAll,
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return cpuTimes{}, fmt.Errorf("scan /proc/stat: %w", err)
	}

	return cpuTimes{}, errors.New("aggregate cpu line not found")
}

func parseUptimeInfo(data []byte) (UptimeInfo, error) {

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return UptimeInfo{}, errors.New("uptime value not found")
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return UptimeInfo{}, fmt.Errorf("parse uptime: %w", err)
	}

	return UptimeInfo{Seconds: uint64(seconds)}, nil
}

func parseMemInfo(data []byte) (MemoryInfo, error) {

	info := MemoryInfo{}
	lines := strings.Split(string(data), "\n")

	var totalFound, availableFound bool
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return MemoryInfo{}, fmt.Errorf("invalid MemTotal line: %q", line)
			}

			tmp, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return MemoryInfo{}, fmt.Errorf("parse MemTotal: %w", err)
			}
			totalFound = true
			info.TotalBytes = tmp
			info.TotalBytes = info.TotalBytes * 1024

		}

		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return MemoryInfo{}, fmt.Errorf("invalid MemAvailable line: %q", line)
			}

			tmp, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return MemoryInfo{}, fmt.Errorf("parse MemAvailable: %w", err)
			}
			availableFound = true
			info.AvailableBytes = tmp
			info.AvailableBytes = info.AvailableBytes * 1024

		}

	}
	if !totalFound && !availableFound {
		return MemoryInfo{}, errors.New("MemTotal and MemAvailable not found")
	}
	if !availableFound {
		return MemoryInfo{}, errors.New("MemAvailable not found")
	}
	if !totalFound {
		return MemoryInfo{}, errors.New("MemTotal not found")
	}

	if info.AvailableBytes > info.TotalBytes {
		return MemoryInfo{}, fmt.Errorf("MemAvailable exceeds MemTotal: available=%d total=%d", info.AvailableBytes, info.TotalBytes)
	}

	info.UsedBytes = info.TotalBytes - info.AvailableBytes

	return info, nil

}
