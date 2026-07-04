package dvgoutils_test

import (
	"fmt"
	"strconv"

	"github.com/davidvanlaatum/dvgoutils"
)

func ExampleFilterSlice() {
	numbers := []int{1, 2, 3, 4, 5, 6}

	even := dvgoutils.FilterSlice(numbers, func(v int) bool {
		return v%2 == 0
	})

	fmt.Println(even)

	// Output:
	// [2 4 6]
}

func ExampleMapSlice() {
	numbers := []int{1, 2, 3}

	labels := dvgoutils.MapSlice(numbers, func(v int) string {
		return fmt.Sprintf("item-%d", v)
	})

	fmt.Println(labels)

	// Output:
	// [item-1 item-2 item-3]
}

func ExampleMust() {
	value := dvgoutils.Must(strconv.Atoi("42"))

	fmt.Println(value)

	// Output:
	// 42
}

func ExamplePtr() {
	value := dvgoutils.Ptr("ready")

	fmt.Println(*value)

	// Output:
	// ready
}
