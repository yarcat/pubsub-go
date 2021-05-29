package pubsub_test

// import (
// 	"log"
// 	"testing"

// 	"github.com/yarcat/pubsub-go"
// )

// func TestPubSub(t *testing.T) {
// 	for _, tc := range []struct {
// 		name     string
// 		subs     int
// 		messages []string
// 	}{{
// 		name: "no subscribers",
// 	}} {
// 		t.Run(tc.name, func(t *testing.T) {
// 			topic := pubsub.NewStringTopic()
// 			sub := newCompositeSub(topic, tc.subs)
// 			pub := pubsub.NewPublisher(topic)
// 			for _, msg := range tc.messages {
// 				if err := pubsub.PublishMany(pub, messages...); err != nil {
// 					log.Fatalf("Publish(%q) got = %v, want = nil", msg, err)
// 				}
// 			}
// 		})
// 	}
// }
