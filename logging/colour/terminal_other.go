//go:build !windows

package colour

import "golang.org/x/term"

func prepareTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
