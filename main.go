package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	token string
}

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

func (s *Server) auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if token != s.token {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func readCPUTimes() (cpuTimes, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuTimes{}, fmt.Errorf("read /proc/stat: %w", err)
	}

	return parseCPUTimes(data)
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

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	statusHandler := http.HandlerFunc(s.handleStatus)
	protectedStatusHandler := s.auth(statusHandler)

	mux.Handle("GET /api/v1/status", protectedStatusHandler)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	return mux
}

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
	return Status{
		Memory:      mem,
		Uptime:      uptime,
		CollectedAt: time.Now().UTC(),
		Disk:        disk,
		CPU:         cpu,
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("json encode error", err)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := CollectStatus()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("encode status response %v", err)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, "OK")
}

func main() {

	token, exists := os.LookupEnv("VPSCONNECT_TOKEN")

	if !exists || strings.TrimSpace(token) == "" {
		log.Fatal("VPSCONNECT_TOKEN is required")
	}

	srv := &Server{token: token}
	const address = "127.0.0.1:6767"
	server := &http.Server{Addr: address, Handler: srv.routes()}

	log.Printf("server started on %s", address)
	errChan := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, ShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ShutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
		log.Println("Graceful shutdown complete")
	case err := <-errChan:
		log.Fatalf("ListenAndServe error:%v", err)
	}

}
