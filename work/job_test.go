package work

import (
	"testing"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func TestCreateJobEntry(t *testing.T) {
	t.Run("createJobEntry", func(t *testing.T) {
		now := time.Now()
		id := primitive.NewObjectID()
		if err := createJobEntry(&Job{Id: &id, Start: &now, Stop: &now}); err != nil {
			t.Errorf("Problem creating job entry: %v", err)
		} else {
			if j, err := findJobById(&id); err != nil {
				t.Errorf("Cannot load job entry: %v", err)
			} else if j == nil {
				t.Error("Null job entry returned")
			}
		}
	})
}
