package run_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/run"
)

var ctx = context.Background()

func TestMemoryStore_SaveGetUpdate(t *testing.T) {
	t.Parallel()
	s := run.NewMemoryStore(0)

	_ = s.Save(ctx, run.Run{ID: "a", Status: run.StatusRunning})
	_ = s.Save(ctx, run.Run{ID: "a", Status: run.StatusSucceeded})

	got, err := s.Get(ctx, "a")
	if err != nil || got.Status != run.StatusSucceeded {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	if all, _ := s.List(ctx, run.Filter{}); len(all) != 1 {
		t.Fatalf("an update must not duplicate the run, got %d", len(all))
	}
	if _, err := s.Get(ctx, "missing"); !errors.Is(err, run.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMemoryStore_ListNewestFirstWithFilterAndLimit(t *testing.T) {
	t.Parallel()
	s := run.NewMemoryStore(0)
	for i := range 6 {
		p := "even"
		if i%2 == 1 {
			p = "odd"
		}
		_ = s.Save(ctx, run.Run{ID: fmt.Sprint(i), Pipeline: p})
	}

	got, _ := s.List(ctx, run.Filter{Pipeline: "odd", Limit: 2})
	if len(got) != 2 || got[0].ID != "5" || got[1].ID != "3" {
		t.Fatalf("got %+v, want runs 5 and 3", got)
	}
}

func TestMemoryStore_EvictsOldest(t *testing.T) {
	t.Parallel()
	s := run.NewMemoryStore(2)
	for _, id := range []string{"1", "2", "3"} {
		_ = s.Save(ctx, run.Run{ID: id})
	}
	if _, err := s.Get(ctx, "1"); !errors.Is(err, run.ErrNotFound) {
		t.Fatal("oldest run was not evicted")
	}
	if all, _ := s.List(ctx, run.Filter{}); len(all) != 2 {
		t.Fatalf("len = %d, want 2", len(all))
	}
}
