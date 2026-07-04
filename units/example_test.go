package units_test

import (
	"fmt"
	"log/slog"

	"github.com/davidvanlaatum/dvgoutils/units"
)

func ExampleBytes() {
	size := 42 * units.MiB

	fmt.Println(size)

	// Output:
	// 42.0 MiB
}

func ExampleBits() {
	rate := 100 * units.Mb

	fmt.Println(rate)

	// Output:
	// 100.0 Mb
}

func ExampleBytes_LogValue() {
	value := slog.AnyValue(1536 * units.Bytes(1))

	fmt.Println(value.Resolve())

	// Output:
	// 1.5 KiB
}
