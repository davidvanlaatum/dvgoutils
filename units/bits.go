package units

import (
	"fmt"
)

type Bits int64

const (
	Kb Bits = 1000 * (iota + 1)
	Mb
	Gb
	Tb
	Pb
	Eb
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
