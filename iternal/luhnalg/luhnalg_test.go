package luhnalg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckNumber(t *testing.T) {
	t.Run("test true #1", func(t *testing.T) {
		got := CheckNumber(4561261212345467)
		require.True(t, got)
	})

	t.Run("test true #2", func(t *testing.T) {
		got := CheckNumber(3465502494)
		require.True(t, got)
	})

	t.Run("test true #3", func(t *testing.T) {
		got := CheckNumber(7000166989766106378)
		require.True(t, got)
	})

	t.Run("test false #1", func(t *testing.T) {
		got := CheckNumber(4561261212345464)
		require.False(t, got)
	})

	t.Run("test false #2", func(t *testing.T) {
		got := CheckNumber(346550249)
		require.False(t, got)
	})

	t.Run("test false #3", func(t *testing.T) {
		got := CheckNumber(700016698976610637)
		require.False(t, got)
	})
}
