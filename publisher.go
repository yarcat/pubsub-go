package pubsub

// Publisher sends messages to all topic subsribers.
type Publisher struct{ t Topic }

// NewPublisher returns new message publisher.
func NewPublisher(topic Topic) Publisher { return Publisher{topic} }

// Topic returns an associated topic.
func (pub Publisher) Topic() Topic { return pub.t }

// Publish sends the message to all registered subscribers.
func (pub Publisher) Publish(msg interface{}) error {
	be, tID := pub.t.ID()
	return be.Publish(tID, msg)
}
