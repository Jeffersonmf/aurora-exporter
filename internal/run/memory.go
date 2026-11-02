package run

import (
	"context"
	"sync"
)

const (
	defaultCapacity  = 1000
	defaultListLimit = 50
)

// MemoryStore keeps the most recent runs in memory, evicting the
// oldest once capacity is reached. History does not survive a restart;
// a durable Store (e.g. a table in the target database) can replace it
// without touching any caller.
type MemoryStore struct {
	mu       sync.RWMutex
	capacity int
	byID     map[string]Run
	order    []string // insertion order, oldest first
}

// NewMemoryStore returns a store holding up to capacity runs (zero
// uses a default of 1000).
func NewMemoryStore(capacity int) *MemoryStore {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	return &MemoryStore{capacity: capacity, byID: map[string]Run{}}
}

// Save inserts or updates r.
func (s *MemoryStore) Save(_ context.Context, r Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byID[r.ID]; !exists {
		s.order = append(s.order, r.ID)
		if len(s.order) > s.capacity {
			delete(s.byID, s.order[0])
			s.order = s.order[1:]
		}
	}
	s.byID[r.ID] = r
	return nil
}

// Get returns the run with id.
func (s *MemoryStore) Get(_ context.Context, id string) (Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.byID[id]
	if !ok {
		return Run{}, ErrNotFound
	}
	return r, nil
}

// List returns runs matching f, newest first.
func (s *MemoryStore) List(_ context.Context, f Filter) ([]Run, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Run, 0, min(limit, len(s.order)))
	for i := len(s.order) - 1; i >= 0 && len(out) < limit; i-- {
		r := s.byID[s.order[i]]
		if f.Pipeline == "" || r.Pipeline == f.Pipeline {
			out = append(out, r)
		}
	}
	return out, nil
}
