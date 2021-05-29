package pubsub

import (
	"context"
	"errors"
)

// Backend is an abstract pub/sub topic storage.
type Backend interface {
	// NewTopic creates a new topic and returns its ID.
	NewTopic() TopicID

	// NewSubscriber registers a new subscriber for the topic and returns
	// subscriber's ID.
	NewSubscriber(TopicID) (SubscriberID, error)

	// Publish sends the message to all subscribers of the topic.
	Publish(TopicID, interface{}) error

	// Receive waits for data.
	Receive(context.Context, TopicID, SubscriberID) (interface{}, error)

	// Close closes the topic preventing any messages from being published.
	// All subscribers will receive ErrTopicClosed and should disconnect.
	Close(TopicID) error

	// Disconnect disconnects the subscriber.
	Disconnect(TopicID, SubscriberID) error
}

var (
	// ErrTopicNotFound is returned when operating on topics (e.g. sending to)
	// that don't exist.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrSubNotFound is returned when requesting data for subscribers that
	// don't exist.
	ErrSubNotFound = errors.New("subscriber not found")

	// ErrTopicClosed is returned when a topic is closed. There will be no
	// any further messages published on this topic. Subscribers should
	// disconnect and let the topic to be garbage-collected.
	ErrTopicClosed = errors.New("topic is closed")
)
