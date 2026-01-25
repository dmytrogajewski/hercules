package rbtree

import (
	"hash/fnv"
	"sync"
)

// ShardedAllocator manages multiple Allocators to allow parallel access.
type ShardedAllocator struct {
	shards []*Allocator
}

// NewShardedAllocator creates a new ShardedAllocator with n shards.
func NewShardedAllocator(n int, hibernationThreshold int) *ShardedAllocator {
	if n <= 0 {
		n = 1
	}
	shards := make([]*Allocator, n)
	for i := 0; i < n; i++ {
		shards[i] = NewAllocator()
		shards[i].HibernationThreshold = hibernationThreshold / n // Distribute threshold?
		// Or keep threshold per shard? Threshold is usually for "when to hibernate".
		// If we split data, we should probably scale threshold or keep it same if it's "min size to bother compressing".
		// Let's keep it proportionate or just copy. If it's 0, it's 0.
		if hibernationThreshold > 0 {
			shards[i].HibernationThreshold = hibernationThreshold / n
			if shards[i].HibernationThreshold == 0 {
				shards[i].HibernationThreshold = 1000 // minimal reasonable default if division results in 0
			}
		}
	}
	return &ShardedAllocator{shards: shards}
}

// GetShard returns the allocator shard for the given key.
func (s *ShardedAllocator) GetShard(key string) *Allocator {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := int(h.Sum32()) % len(s.shards)
	if idx < 0 {
		idx = -idx
	}
	return s.shards[idx]
}

// Shards returns all underlying allocators.
func (s *ShardedAllocator) Shards() []*Allocator {
	return s.shards
}

// Hibernate hibernates all shards in parallel.
func (s *ShardedAllocator) Hibernate() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.shards))
	for _, shard := range s.shards {
		go func(a *Allocator) {
			defer wg.Done()
			a.Hibernate()
		}(shard)
	}
	wg.Wait()
}

// Boot boots all shards in parallel.
func (s *ShardedAllocator) Boot() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.shards))
	for _, shard := range s.shards {
		go func(a *Allocator) {
			defer wg.Done()
			a.Boot()
		}(shard)
	}
	wg.Wait()
}

// Serialize serializes all shards.
// Note: This needs a strategy for filenames. Currently Allocator.Serialize takes a path.
// We probably need to manage paths outside or add index to filename.
// But Allocator.Serialize is what we call.
// This method might not be needed if caller handles iteration, but convenient.
// We'll leave it to caller or implement later if needed by Burndown.
