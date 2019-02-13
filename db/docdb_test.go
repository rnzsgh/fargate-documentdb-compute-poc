package db

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("TestInit", func(t *testing.T) {
		if err := Ping(context.TODO()); err != nil {
			t.Errorf("Unable to ping db: %v", err)
		}
	})
}
