package pubsub

import "fmt"

func errTopicNotFound(tID TopicID) error {
	return fmt.Errorf("%w: topic ID = %v", ErrTopicNotFound, tID)
}

func errSubscriberNotFound(tID TopicID, sID SubscriberID) error {
	return fmt.Errorf("%w: topic ID = %v: subscriber ID = %v",
		ErrSubNotFound, tID, sID)
}

func errTopicClosed(tID TopicID) error {
	return fmt.Errorf("%w: topic ID = %v", ErrTopicClosed, tID)
}
