// Package collector provides the HTTP collector server and write buffer.
package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/CodecWhiz/streamlens/internal/cmcd"
	"github.com/CodecWhiz/streamlens/internal/storage"
)

// Buffer collects CMCD events and flushes them to ClickHouse in batches.
type Buffer struct {
	mu       sync.Mutex
	events   []cmcd.Event
	client   *storage.Client
	maxSize  int
	interval time.Duration
	done     chan struct{}
	wg       sync.WaitGroup
}

// NewBuffer creates a new write buffer.
// It flushes when maxSize events are collected or every interval, whichever comes first.
func NewBuffer(client *storage.Client, maxSize int, interval time.Duration) *Buffer {
	b := &Buffer{
		client:   client,
		maxSize:  maxSize,
		interval: interval,
		done:     make(chan struct{}),
	}
	b.wg.Add(1)
	go b.flushLoop()
	return b
}

// Add adds an event to the buffer. Thread-safe.
func (b *Buffer) Add(event cmcd.Event) {
	b.mu.Lock()
	b.events = append(b.events, event)
	shouldFlush := len(b.events) >= b.maxSize
	b.mu.Unlock()

	if shouldFlush {
		b.Flush()
	}
}

// Flush sends all buffered events to ClickHouse.
func (b *Buffer) Flush() {
	b.mu.Lock()
	if len(b.events) == 0 {
		b.mu.Unlock()
		return
	}
	batch := b.events
	b.events = nil
	b.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.client.InsertEvents(ctx, batch); err != nil {
		log.Printf("ERROR: flush %d events: %v", len(batch), err)
		// Re-add failed events
		b.mu.Lock()
		b.events = append(batch, b.events...)
		b.mu.Unlock()
		return
	}
	log.Printf("Flushed %d events to ClickHouse", len(batch))
}

// Close stops the flush loop and flushes remaining events.
func (b *Buffer) Close() {
	close(b.done)
	b.wg.Wait()
	b.Flush()
}

func (b *Buffer) flushLoop() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.Flush()
		case <-b.done:
			return
		}
	}
}
