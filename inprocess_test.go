package pubsub_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/yarcat/pubsub-go"
)

func TestLocalBackend(t *testing.T) {
	for _, tc := range []struct {
		name         string
		topics, subs int
		msgs         []string
		want         string
	}{{
		name:   "topics=1 subs=1 no messages",
		topics: 1, subs: 1,
		want: "disconnect 0 0\n",
	}, {
		name:   "topics=1 subs=2 no messages",
		topics: 1, subs: 2,
		want: "disconnect 0 0\ndisconnect 0 1\n",
	}, {
		name:   "topics=1 subs=2",
		topics: 1, subs: 2,
		msgs: []string{"foo", "bar"},
		want: "00-foo\n00-bar\ndisconnect 0 0\n00-foo\n00-bar\ndisconnect 0 1\n",
	}, {
		name:   "topics=2 subs=2",
		topics: 2, subs: 2,
		msgs: []string{"foo", "bar"},
		want: "00-foo\n00-bar\ndisconnect 0 0\n00-foo\n00-bar\ndisconnect 0 1\n" +
			"01-foo\n01-bar\ndisconnect 1 0\n01-foo\n01-bar\ndisconnect 1 1\n",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			be := pubsub.NewInprocess()
			ts := newTopics(tc.topics, be)
			subs := newSubs(t, tc.topics, ts, tc.subs)
			sendMessages(t, ts, tc.msgs)
			got := receiveAll(t, ts, subs)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("pubsub got unexpected diff (-want, +got): %v", diff)
			}
		})
	}
}

func receiveAll(t *testing.T, ts []pubsub.Topic,
	subs map[pubsub.Topic][]pubsub.Subscriber,
) string {
	var b strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for i, topic := range ts {
		for j, s := range subs[topic] {
			receiveOne(ctx, t, s, i, j, &b)
		}
	}
	return b.String()
}

func receiveOne(ctx context.Context, t *testing.T,
	s pubsub.Subscriber, topicNo, subNo int, b *strings.Builder,
) {
	for {
		data, err := s.Receive(ctx)
		if errors.Is(err, pubsub.ErrTopicClosed) {
			if err := s.Disconnect(); err != nil {
				t.Errorf("topic#%02d.sub#%02d.Disconnect() err = %v, want = nil",
					topicNo, subNo, err)
			} else {
				fmt.Fprintln(b, "disconnect", topicNo, subNo)
			}
			break
		} else if err != nil {
			t.Errorf("topic#%02d.sub#%02d.Receive() err = %v, want = nil",
				topicNo, subNo, err)
			break
		}
		fmt.Fprintln(b, data)
	}
}

func sendMessages(t *testing.T, ts []pubsub.Topic, messages []string) {
	for i, topic := range ts {
		for _, m := range messages {
			err := topic.Publish(fmt.Sprintf("%02d-%s", i, m))
			if err != nil {
				t.Errorf("topic#%02d.Publish() err = %v, want = nil",
					i, err)
			}
		}
		err := topic.Close()
		if err != nil {
			t.Errorf("topic#%02d.Close() err = %v, want = nil", i, err)
		}
	}
}

func newSubs(t *testing.T, topics int, ts []pubsub.Topic, subscribers int,
) map[pubsub.Topic][]pubsub.Subscriber {
	subs := make(map[pubsub.Topic][]pubsub.Subscriber, topics)
	for _, topic := range ts {
		for i := 0; i < subscribers; i++ {
			sub, err := topic.NewSubscriber()
			if err != nil {
				t.Fatalf("NewSubscriber() err = %v, want = nil", err)
			}
			subs[topic] = append(subs[topic], sub)
		}
	}
	return subs
}

func newTopics(topics int, be pubsub.Inprocess) []pubsub.Topic {
	var ts []pubsub.Topic
	for i := 0; i < topics; i++ {
		ts = append(ts, pubsub.NewTopic(be))
	}
	return ts
}
