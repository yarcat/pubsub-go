package pubsub

import (
	"context"
	"log"
	"sync/atomic"
)

// Inprocess represents a local process topic container. Inprocess implements
// Backend interface.
type Inprocess struct {
	ts *topics
}

// NewInprocess returns new inprocess topic container.
func NewInprocess() Inprocess {
	return Inprocess{ts: newTopics()}
}

// NewTopic registers a new topic.
func (be Inprocess) NewTopic() TopicID {
	return be.ts.Add()
}

// NewSubscriber registers a new subscriber for the topic.
func (be Inprocess) NewSubscriber(tID TopicID) (SubscriberID, error) {
	return be.ts.AddSubscriber(tID)
}

// Publish sends the message the all registered subscribers.
func (be Inprocess) Publish(tID TopicID, msg interface{}) error {
	return be.ts.Publish(tID, msg)
}

// Receive waits for a data received by the subscriber on the topic.
func (be Inprocess) Receive(ctx context.Context, tID TopicID, sID SubscriberID) (interface{}, error) {
	return be.ts.Receive(ctx, tID, sID)
}

// Close stops publishing message for the topic. Subscribers will receive
// ErrTopicClosed and should disconnect.
func (be Inprocess) Close(tID TopicID) error {
	return be.ts.Close(tID)
}

// Disconnect disconnects the subscriber releasing allocated resources.
func (be Inprocess) Disconnect(tID TopicID, sID SubscriberID) error {
	return be.ts.Disconnect(tID, sID)
}

type topics struct {
	counter atomicCounter
	reg     chan map[TopicID]*topic
}

func newTopics() *topics {
	reg := make(chan map[TopicID]*topic, 1)
	reg <- make(map[TopicID]*topic)
	return &topics{reg: reg}
}

// Add adds  new topic.
func (t *topics) Add() TopicID {
	tID := TopicID(t.counter.Next())
	reg := <-t.reg
	reg[tID] = newTopic()
	t.reg <- reg
	return tID
}

func (t *topics) topic(tID TopicID) *topic {
	reg := <-t.reg
	topic := reg[tID]
	t.reg <- reg
	return topic
}

// AddSubscriber adds new subscriber to the topic.
func (t *topics) AddSubscriber(tID TopicID) (SubscriberID, error) {
	return t.topic(tID).Add(tID)
}

// Publish sends the message for all subscribers of the topic.
func (t *topics) Publish(tID TopicID, msg interface{}) error {
	return t.topic(tID).Publish(tID, msg, nil)
}

// Receive waits for data.
func (t *topics) Receive(ctx context.Context, tID TopicID, sID SubscriberID) (interface{}, error) {
	return t.topic(tID).Receive(ctx, tID, sID)
}

// Close prevents messages from being published on the topic.
func (t *topics) Close(tID TopicID) error {
	return t.topic(tID).Close(tID)
}

// Disconnect disconnects the subscribers releasing allocated resources.
func (t *topics) Disconnect(tID TopicID, sID SubscriberID) error {
	return t.topic(tID).Disconnect(tID, sID)
}

type topic struct {
	counter atomicCounter
	closed  atomic.Value
	reg     chan map[SubscriberID]queue
}

func newTopic() *topic {
	reg := make(chan map[SubscriberID]queue, 1)
	reg <- make(map[SubscriberID]queue)
	return &topic{reg: reg}
}

// Add adds new subscriber queue to this topic.
func (t *topic) Add(tID TopicID) (SubscriberID, error) {
	if t == nil {
		return 0, errTopicNotFound(tID)
	}
	sID := SubscriberID(t.counter.Next())
	reg := <-t.reg
	reg[sID] = newQueue()
	t.reg <- reg
	return sID, nil
}

// Publish enqueues the message for all subscribers of this topic.
func (t *topic) Publish(tID TopicID, msg interface{}, err error) error {
	if t == nil {
		return errTopicNotFound(tID)
	}
	if err := t.checkClosed(tID); err != nil {
		return err
	}
	reg := <-t.reg
	for _, q := range reg {
		q.Add(msgOrError{Msg: msg, Err: err})
	}
	regLen := len(reg)
	t.reg <- reg
	log.Printf("Publish(tID=%v, %v, %v): published to %d subscribers",
		tID, msg, err, regLen)
	return nil
}

// Receive waits for data.
func (t *topic) Receive(ctx context.Context, tID TopicID, sID SubscriberID) (interface{}, error) {
	if t == nil {
		return nil, errTopicNotFound(tID)
	}
	reg := <-t.reg
	q, ok := reg[sID]
	t.reg <- reg
	if !ok {
		return nil, errSubscriberNotFound(tID, sID)
	}
	msg, err := q.Get(ctx)
	if err != nil {
		return nil, err
	}
	if msg := msg.(msgOrError); msg.Err != nil {
		return nil, msg.Err
	} else {
		return msg.Msg, nil
	}
}

// Close prevents further publishes.
func (t *topic) Close(tID TopicID) error {
	if t == nil {
		return errTopicNotFound(tID)
	}
	if err := t.checkClosed(tID); err != nil {
		return err
	}
	t.closed.Store(true)
	return t.Publish(tID, nil, ErrTopicClosed)
}

func (t *topic) Disconnect(tID TopicID, sID SubscriberID) error {
	if t == nil {
		return errTopicNotFound(tID)
	}
	if err := t.checkClosed(tID); err != nil {
		return err
	}
	reg := <-t.reg
	delete(reg, sID)
	t.reg <- reg
	return nil
}

func (t *topic) checkClosed(tID TopicID) error {
	if v := t.closed.Load(); v != nil && v.(bool) {
		errTopicClosed(tID)
	}
	return nil
}

type msgOrError struct {
	Msg interface{}
	Err error
}
