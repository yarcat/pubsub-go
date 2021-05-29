package pubsub

import "context"

type queue struct {
	empty chan struct{}
	items chan []interface{}
}

func newQueue() queue {
	empty := make(chan struct{}, 1)
	empty <- struct{}{}
	return queue{
		empty: empty,
		items: make(chan []interface{}, 1),
	}
}

// Add appends the value to the queue.
func (q queue) Add(v interface{}) {
	var items []interface{}
	select {
	case <-q.empty:
	case items = <-q.items:
	}
	q.items <- append(items, v)
}

// Get blocks until cancelled or an equeued value was received.
func (q queue) Get(ctx context.Context) (interface{}, error) {
	var items []interface{}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case items = <-q.items:
	}
	v := items[0]
	if items = items[1:]; len(items) == 0 {
		q.empty <- struct{}{}
	} else {
		q.items <- items
	}
	return v, nil
}
