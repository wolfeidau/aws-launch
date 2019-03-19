package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {

	got := New(nil)
	require.NotNil(t, got)
}
