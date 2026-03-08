package units

import (
	"fmt"
)

type Bytes int64

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
