package pubsub

import "sync/atomic"

type (
	// ID represents a generic identifier.
	ID uint64
	// TopicID represents a topic identifier.
	TopicID ID
	// SubscriberID represents a subscriber identifier.
	SubscriberID ID
)

type atomicCounter struct{ last ID }

// Next returns next available topic identifier.
func (ac *atomicCounter) Next() ID {
	return ID(atomic.AddUint64((*uint64)(&ac.last), 1))
}
