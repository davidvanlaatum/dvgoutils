package colour

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestWindowsConsolePreparation(t *testing.T) {
	t.Parallel()
	const required = windows.ENABLE_PROCESSED_OUTPUT | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	for _, tc := range []struct {
		name           string
		mode           uint32
		getErr, setErr error
		want, wantSet  bool
	}{
		{name: "enable both flags", want: true, wantSet: true},
		{name: "preserve existing flags", mode: windows.ENABLE_WRAP_AT_EOL_OUTPUT | windows.DISABLE_NEWLINE_AUTO_RETURN | windows.ENABLE_LVB_GRID_WORLDWIDE, want: true, wantSet: true},
		{name: "add VT processing", mode: windows.ENABLE_PROCESSED_OUTPUT, want: true, wantSet: true},
		{name: "add processed output", mode: windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING, want: true, wantSet: true},
		{name: "already enabled", mode: required | windows.ENABLE_WRAP_AT_EOL_OUTPUT, want: true},
		{name: "not a console", getErr: windows.ERROR_INVALID_HANDLE},
		{name: "cannot read mode", getErr: windows.ERROR_ACCESS_DENIED},
		{name: "unsupported VT processing", setErr: windows.ERROR_INVALID_PARAMETER, wantSet: true},
		{name: "cannot set mode", setErr: windows.ERROR_ACCESS_DENIED, wantSet: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			const handle = windows.Handle(123)
			var gotSet bool
			prepare := func(fd uintptr) bool {
				return prepareConsole(windows.Handle(fd), func(got windows.Handle, mode *uint32) error {
					r.Equal(handle, got)
					*mode = tc.mode
					return tc.getErr
				}, func(got windows.Handle, mode uint32) error {
					r.Equal(handle, got)
					r.Equal(tc.mode|required, mode)
					gotSet = true
					return tc.setErr
				})
			}
			lookup := func(string) (string, bool) { return "", false }
			r.Equal(tc.want, autoColour(&descriptorWriter{}, lookup, prepare))
			r.Equal(tc.wantSet, gotSet)
		})
	}
}
