package units

import (
	"fmt"
	"log/slog"
	"math"
	"time"
)

type Bytes int64

type BytesPerSecond float64

const (
	KiB Bytes = 1 << (10 * (iota + 1))
	MiB
	GiB
	TiB
	PiB
	EiB
)

func (b Bytes) String() string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (b Bytes) LogValue() slog.Value {
	return slog.StringValue(b.String())
}

func (b Bytes) PerSecond(duration time.Duration) BytesPerSecond {
	if duration <= 0 {
		return BytesPerSecond(math.NaN())
	}

	return BytesPerSecond(float64(b) / duration.Seconds())
}

func (b BytesPerSecond) String() string {
	const unit = 1024
	const prefixes = "KMGTPE"
	if math.IsNaN(float64(b)) || b < unit {
		return fmt.Sprintf("%.1f B/s", b)
	}
	div, exp := float64(unit), 0
	for n := float64(b) / unit; n >= unit && exp < len(prefixes)-1; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB/s", float64(b)/div, prefixes[exp])
}

func (b BytesPerSecond) LogValue() slog.Value {
	return slog.StringValue(b.String())
}
