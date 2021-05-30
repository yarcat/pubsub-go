package pubsub

// Topic represents a unique pub/sub topic.
type Topic struct {
	id TopicID
	be Backend
}

// Create topic creates new topic and returns a topic structure.
func CreateTopic(be Backend) Topic { return NewTopic(be, be.NewTopic()) }

// NewTopic returns a pub/sub topic structure.
func NewTopic(be Backend, tID TopicID) Topic { return Topic{id: tID, be: be} }

// ID returns a backend and its topic ID.
func (t Topic) ID() (Backend, TopicID) { return t.be, t.id }

// Close closes the topic preventing any messages from being published. All
// subscribers will receive ErrTopicClosed and should disconnect.
func (t Topic) Close() error { return t.be.Close(t.id) }
