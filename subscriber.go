package pubsub

import "context"

// Subscription uniquely represents a subscriber.
type Subscription struct {
	id SubscriberID
	t  Topic
}

// Subscribe creates new topic subscription.
func Subscribe(topic Topic) (Subscription, error) {
	be, tID := topic.ID()
	sID, err := be.NewSubscriber(tID)
	if err != nil {
		return Subscription{}, err
	}
	return NewSubscription(topic, sID), nil

}

// NewSubscription returns a subscriber associated with the topic and sub ID.
func NewSubscription(topic Topic, sID SubscriberID) Subscription {
	return Subscription{id: sID, t: topic}
}

// Topic returns an associated topic.
func (sub Subscription) Topic() Topic { return sub.t }

// ID returns a backend, topic and subscription IDs associated.
func (sub Subscription) ID() (Backend, TopicID, SubscriberID) {
	be, tID := sub.t.ID()
	return be, tID, sub.id
}

// Receive waits for data.
func (sub Subscription) Receive(ctx context.Context) (interface{}, error) {
	be, tID, sID := sub.ID()
	return be.Receive(ctx, tID, sID)
}

// Disconnect disconnects the subscriber releasing allocated resources.
func (sub Subscription) Disconnect() error {
	be, tID, sID := sub.ID()
	return be.Disconnect(tID, sID)
}
