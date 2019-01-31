package db

import (
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("TestInit", func(t *testing.T) {
		if err := ping(); err != nil {
			t.Errorf("Unable to ping db: %v", err)
		}
	})
}
