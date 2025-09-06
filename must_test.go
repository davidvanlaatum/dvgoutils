package dvgoutils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMust(t *testing.T) {
	r := require.New(t)
	r.Equal(1, Must(1, nil))
}

func TestMust_Panic(t *testing.T) {
	r := require.New(t)
	r.Panics(func() {
		Must(1, errors.New("error"))
	})
}
