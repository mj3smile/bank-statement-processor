package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/mj3smile/bank-statement-processor/internal/infra/log"
)

type Bus interface {
	Publish(event Event)
	Subscribe() <-chan Event
	Close()
}

type bus struct {
	appCtx      context.Context
	mu          sync.RWMutex
	subscribers []chan Event
	bufferSize  int
	closed      bool
}

func NewBus(appCtx context.Context) Bus {
	return NewBusWithBuffer(appCtx, 100)
}

func NewBusWithBuffer(appCtx context.Context, bufferSize int) Bus {
	return &bus{
		appCtx:      appCtx,
		subscribers: make([]chan Event, 0),
		bufferSize:  bufferSize,
		closed:      false,
	}
}

func (b *bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		log.Warn(b.appCtx, "attempted to publish to closed event bus")
		return
	}

	for _, ch := range b.subscribers {
		ch <- event
	}
}

func (b *bus) Subscribe() <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		log.Warn(b.appCtx, "attempted to subscribe to closed event bus")
		ch := make(chan Event)
		close(ch)
		return ch
	}

	ch := make(chan Event, b.bufferSize)
	b.subscribers = append(b.subscribers, ch)

	log.Info(b.appCtx, fmt.Sprintf("new subscriber added (total: %d)", len(b.subscribers)))
	return ch
}

func (b *bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	for i, ch := range b.subscribers {
		close(ch)
		log.Info(b.appCtx, fmt.Sprintf("closed subscriber channel %d", i))
	}

	b.subscribers = nil
	log.Info(b.appCtx, "event bus closed")
}

func (b *bus) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
