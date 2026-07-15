package postgres

import (
	"context"
	"fmt"

	"github.com/Fregres/VPSconnect/internal/metrics"
)

func (s *Storage) SaveMetric(ctx context.Context, serverID int64, status metrics.Status) error {
	query := `INSERT INTO metric_samples (server_id,
	collected_at, cpu_usage_percent,memory_total_bytes,
	memory_used_bytes,disk_total_bytes,disk_used_bytes,
	uptime_seconds)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	_, err := s.pool.Exec(ctx, query, 1, status.CollectedAt, status.CPU.UsagePercent, int64(status.Memory.TotalBytes), int64(status.Memory.UsedBytes), status.Disk.TotalBytes, status.Disk.UsedBytes, status.Uptime.Seconds)
	if err != nil {
		return fmt.Errorf("save metric sample: %w", err)
	}
	return nil
}
