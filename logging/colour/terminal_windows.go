package colour

import "golang.org/x/sys/windows"

func prepareTerminal(fd uintptr) bool {
	return prepareConsole(windows.Handle(fd), windows.GetConsoleMode, windows.SetConsoleMode)
}

func prepareConsole(handle windows.Handle, getMode func(windows.Handle, *uint32) error, setMode func(windows.Handle, uint32) error) bool {
	var mode uint32
	if err := getMode(handle, &mode); err != nil {
		return false
	}
	// Both flags are required for ANSI output. Preserve all other console modes.
	// https://learn.microsoft.com/en-us/windows/console/setconsolemode
	want := mode | windows.ENABLE_PROCESSED_OUTPUT | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if mode == want {
		return true
	}
	return setMode(handle, want) == nil
}
