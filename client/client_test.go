package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yadukrishnan2004/antrelay-sdk/client"
)

func TestNew(t *testing.T) {
	t.Run("returns client with valid config", func(t *testing.T) {
		c, err := client.New(client.Config{
			ServerURL: "http://localhost:8080",
			Queue:     "orders-queue",
		})

		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("returns error when ServerURL is empty", func(t *testing.T) {
		c, err := client.New(client.Config{
			Queue: "orders-queue",
		})

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "ServerURL")
	})

	t.Run("returns error when Queue is empty", func(t *testing.T) {
		c, err := client.New(client.Config{
			ServerURL: "http://localhost:8080",
		})

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "Queue")
	})

	t.Run("returns error when both fields are empty", func(t *testing.T) {
		c, err := client.New(client.Config{})

		assert.Error(t, err)
		assert.Nil(t, c)
	})
}