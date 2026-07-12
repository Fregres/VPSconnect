package metrics

import (
	"testing"
)

func TestParseMemInfo(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte("MemTotal: 1000 kB\n" + "MemAvailable: 400 kB\n")

		got, err := parseMemInfo(data)
		if err != nil {
			t.Fatalf("parseMemInfo() error = %v", err)
		}

		want := MemoryInfo{TotalBytes: 1_024_000, AvailableBytes: 409_600, UsedBytes: 614_400}

		if got != want {
			t.Errorf("parseMemInfo %+v, want %+v ", got, want)
		}
	})

	t.Run("Available > Total", func(t *testing.T) {
		data := []byte("MemTotal: 400 kB\n" + "MemAvailable: 1000 kB\n")

		_, err := parseMemInfo(data)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

	})

	t.Run("missing MemTotal", func(t *testing.T) {
		data := []byte("MemAvailable: 400 kB\n")
		_, err := parseMemInfo(data)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("missing MemAvailable", func(t *testing.T) {
		data := []byte("MemTotal: 1000 kB\n")
		_, err := parseMemInfo(data)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("wrong format", func(t *testing.T) {
		data := []byte("MemTotal: zero kB\n" + "MemAvailable: 1000 kB\n")
		_, err := parseMemInfo(data)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("wrong format", func(t *testing.T) {
		data := []byte("MemTotal: 1000 kB\n" + "MemAvailable: bobr kB\n")
		_, err := parseMemInfo(data)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

}

func TestParseUptimeInfo(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte("10505.63 24872.45")
		got, err := parseUptimeInfo(data)
		if err != nil {
			t.Fatalf("parseUptimeInfo() error = %v", err)
		}
		want := UptimeInfo{Seconds: 10505}

		if got != want {
			t.Fatalf("parseUptimeInfo %+v, want %+v", got, want)
		}
	})

	t.Run("missing data", func(t *testing.T) {
		data := []byte("")
		_, err := parseUptimeInfo(data)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
	t.Run("wrong data", func(t *testing.T) {
		data := []byte("wrong data test")
		_, err := parseUptimeInfo(data)

		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestParseCPUTimes(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte("cpu  710603 2563 206231 42062767 4761 13 5267 0 0 0")
		got, err := parseCPUTimes(data)
		if err != nil {
			t.Fatalf("parseCPUTimes() error %v", err)
		}

		want := cpuTimes{
			idle:  42_067_528,
			total: 42_992_205,
		}

		if got != want {

			t.Fatalf("parseCPUTimes() %+v, want %+v", got, want)
		}
	})

	t.Run("bla-bla-bla", func(t *testing.T) {
		data := []byte("cpu0  1 1 1 1 1 1 1 0 0 0 \n" + "cpu  710603 2563 206231 42062767 4761 13 5267 0 0 0\n")
		got, err := parseCPUTimes(data)
		if err != nil {
			t.Fatalf("parseCPUTimes() error %v", err)
		}

		want := cpuTimes{
			idle:  42_067_528,
			total: 42_992_205,
		}

		if got != want {
			t.Fatalf("parseCPUTimes() %+v, want %+v", got, want)
		}
	})
	t.Run("wrong data format", func(t *testing.T) {
		data := []byte("cpu hello 1 2 3 4 5 6 7")
		_, err := parseCPUTimes(data)
		if err == nil {
			t.Fatalf("error expected, got nil")
		}
	})

	t.Run("no cpu line", func(t *testing.T) {
		data := []byte("cpu0 1 2 3 4 5 6 7 8")
		_, err := parseCPUTimes(data)
		if err == nil {
			t.Fatalf("error expected, got nil")
		}
	})
	t.Run("not enought arguments", func(t *testing.T) {
		data := []byte("cpu 1 2 3 4 5")
		_, err := parseCPUTimes(data)
		if err == nil {
			t.Fatalf("error expected, got nil")
		}
	})
}
