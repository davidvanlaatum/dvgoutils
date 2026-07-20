package units

import (
	"fmt"
	"log/slog"
	"time"
)

type Bits int64

type BitsPerSecond float64

const (
	Kb Bits = 1000
	Mb      = Kb * 1000
	Gb      = Mb * 1000
	Tb      = Gb * 1000
	Pb      = Tb * 1000
	Eb      = Pb * 1000
)

func (b Bits) String() string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d b", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cb", float64(b)/float64(div), "KMGTPE"[exp])
}

func (b Bits) LogValue() slog.Value {
	return slog.StringValue(b.String())
}

func (b Bits) PerSecond(duration time.Duration) BitsPerSecond {
	if duration <= 0 {
		panic("duration must be positive")
	}

	return BitsPerSecond(float64(b) / duration.Seconds())
}

func (b BitsPerSecond) String() string {
	const unit = 1000
	const prefixes = "KMGTPE"
	if b < unit {
		return fmt.Sprintf("%.1f b/s", b)
	}
	div, exp := float64(unit), 0
	for n := float64(b) / unit; n >= unit && exp < len(prefixes)-1; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cb/s", float64(b)/div, prefixes[exp])
}

func (b BitsPerSecond) LogValue() slog.Value {
	return slog.StringValue(b.String())
}
