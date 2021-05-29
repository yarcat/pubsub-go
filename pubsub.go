package pubsub

import "context"

// Topic represents a unique pub/sub topic.
type Topic struct {
	id TopicID
	be Backend
}

// NewTopic returns a unique pub/sub topic.
func NewTopic(be Backend) Topic {
	return Topic{id: be.NewTopic(), be: be}
}

// Receive waits for data.
func (t Topic) Receive(ctx context.Context, sID SubscriberID) (interface{}, error) {
	return t.be.Receive(ctx, t.id, sID)
}

// Publish sends the message to all registered subscribers.
func (t Topic) Publish(msg interface{}) error {
	return t.be.Publish(t.id, msg)
}

// NewSubscriber returns new topic subscriber.
func (t Topic) NewSubscriber() (Subscriber, error) {
	sID, err := t.be.NewSubscriber(t.id)
	if err != nil {
		return Subscriber{}, err
	}
	return Subscriber{id: sID, t: t}, nil
}

// Close closes the topic preventing any messages from being published. All
// subscribers will receive ErrTopicClosed and should disconnect.
func (t Topic) Close() error { return t.be.Close(t.id) }

// Disconnect disconnects the subscriber releasing allocated resources.
func (t Topic) Disconnect(sID SubscriberID) error {
	return t.be.Disconnect(t.id, sID)
}

// Subscriber uniquely represents a subscriber.
type Subscriber struct {
	id SubscriberID
	t  Topic
}

// Receive waits for data.
func (sub Subscriber) Receive(ctx context.Context) (interface{}, error) {
	return sub.t.Receive(ctx, sub.id)
}

// Disconnect disconnects the subscriber releasing allocated resources.
func (sub Subscriber) Disconnect() error {
	return sub.t.Disconnect(sub.id)
}
