package worker_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/worker"
)

func TestBackoff(t *testing.T) {
	t.Run("starts at initial duration", func(t *testing.T) {
		b := worker.NewBackoff(1*time.Second, 30*time.Second)

		assert.Equal(t, 1*time.Second, b.Next())
	})

	t.Run("doubles on each call", func(t *testing.T) {
		b := worker.NewBackoff(1*time.Second, 30*time.Second)

		b.Next()                                         // 1s
		assert.Equal(t, 2*time.Second, b.Next())         // 2s
		assert.Equal(t, 4*time.Second, b.Next())         // 4s
	})

	t.Run("never exceeds maximum", func(t *testing.T) {
		b := worker.NewBackoff(1*time.Second, 5*time.Second)

		b.Next() // 1s
		b.Next() // 2s
		b.Next() // 4s
		b.Next() // would be 8s but capped at 5s

		assert.Equal(t, 5*time.Second, b.Next())
	})

	t.Run("resets back to initial after success", func(t *testing.T) {
		b := worker.NewBackoff(1*time.Second, 30*time.Second)

		b.Next() // 1s
		b.Next() // 2s
		b.Next() // 4s
		b.Reset()

		assert.Equal(t, 1*time.Second, b.Next())
	})
}