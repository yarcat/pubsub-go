# pubsub-go
Simple publisher/subscriber in Go

TBD

# Hypothetical Example

Let's imagine, our application wants to send configuration updates:
- There's a file watcher, which reload file content when it's changed;
- There's a content watcher, which unmarshals config;
- There are several pieces of an app that want to know when the configuration changes.

First we define typed helpers:

```go
func receive(ctx context.Context, sub pubsub.Subscription) (v interface{}, ok bool, err error) {
	v, err = sub.Receive(ctx)
	ok = err != pubsub.ErrTopicClosed
	return v, ok, err
}

type (
    fileContentSender   struct { pub pubsub.Publisher }
    fileContentReceiver struct { sub pubsub.Subscription }
)

func (sender fileContentSender) Send(b []byte) { sender.topic.Publish(b) }
func (receiver fileContentReceiver) Receive(ctx context.Context, b *[]byte, e *error) bool {
	v, ok, err := receive(ctx, receiver.sub)
	*b, *e = v.([]byte), err
	return ok
}

type (
    configSender   struct { topic pubsub.Publisher }
    configReceiver struct { sub pubsub.Subscription }
)

func (sender configSender) Send(cfg Configuration) { sender.topic.Publish(cfg) }
func (receiver configReceiver) Receive(ctx context.Context, c *Config, e *error) bool {
	v, ok, err := receive(ctx, receiver, sub)
	*c, *e = v.(Config), err
	return ok
}
```

And now we can implement our thing:

```go
func main() {
    cfg := flag.String("cfg", "", "configuration file `path`")
    flag.Parse()
    // ...

    repo := pubsub.NewInprocess()

    ctx, cancel := context.WithCancel(context.Background())

	// Listen to file-system changes (probably using fsnotify), publishes to the
	// file change topic.

    fileChanges := pubsub.CreateTopic(repo)
    wg.Add(1)
    go func() {
        defer wg.Done()
        fileWatcher{
            Path: *cfg,
            Send: fileContentSender{pubsub.NewPublisher(fileChanges)}).Send,
        }.Run(ctx)
    }()

	// Subscribe to configuration bytes updates, parse into Configuration
	// structure, publish new Configuration if changed.

    confChanges := pubsub.CreateTopic(repo)
    wg.Add(1)
    go func() {
        defer wg.Done()
        configurationRepository{
            Recv: fileContentReceiver{pubsub.Subscribe(confChanges)}.Receive,
            Send: configSender{pubsub.NewPublisher(confChanges)}).Send,
        }.Run(ctx)
    }()

	// All other application related things go here.

	App{
		PubSub:          repo,
		NewConfReceiver: func() configReceiver {
			return configReceiver{pubsub.Subscribe(confChanges)}
		},
		...
	}.Run()

	// And try to shutdown gracefully.

	cancel()
	BlockOn(wg.Wait).ForMaxOf(shutdownTimeout)
}
```
