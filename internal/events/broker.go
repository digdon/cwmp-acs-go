package events

import (
	"sync"

	"github.com/google/uuid"
)

type Subscriber struct {
	Events chan Event
	Filter func(Event) bool
}

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]Subscriber
}

var GlobalBroker = &Broker{
	subscribers: make(map[string]Subscriber),
}

// Subscribe registers a new subscriber and returns a unique ID and a read-only
// channel on which events will be delivered. Call Unsubscribe when done.
func (b *Broker) Subscribe() (string, <-chan Event) {
	return b.SubscribeWithFilter(nil)
}

// SubscribeWithFilter registers a new subscriber with an optional filter function.
// Only events for which the filter returns true will be delivered to this subscriber.
func (b *Broker) SubscribeWithFilter(filter func(Event) bool) (string, <-chan Event) {
	id := uuid.NewString()
	ch := make(chan Event, 32) // buffered so publishing never blocks the CWMP handler
	b.mu.Lock()
	b.subscribers[id] = Subscriber{
		Events: ch,
		Filter: filter,
	}
	b.mu.Unlock()
	return id, ch
}

// Unsubscribe removes the subscriber and closes its channel.
func (b *Broker) Unsubscribe(id string) {
	b.mu.Lock()
	if sub, ok := b.subscribers[id]; ok {
		close(sub.Events)
		delete(b.subscribers, id)
	}
	b.mu.Unlock()
}

// Publish sends an event to all current subscribers. Slow subscribers that
// have a full buffer will have the event dropped rather than blocking the caller.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subscribers {
		if sub.Filter != nil && !sub.Filter(e) {
			// Filter has been specified and event does not match — skip this subscriber
			continue
		}
		select {
		case sub.Events <- e:
		default:
			// subscriber too slow — drop the event
		}
	}
}
