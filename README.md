# pubsub-go
Simple publisher/subscriber in Go

TBD

# Hypothetical Example

Let's imagine, our application wants to send configuration updates:
- There's a file watcher, which reload file content when it's changed;
- There's a content watcher, which unmarshals config;
- There are several pieces of an app that want to know when the configuration changes.

## Helpers

### Topic termination

First we define a helper to distinguish between loggable errors and
end-of-stream (kinda similar to handling EOFs):

```go
func receive(ctx context.Context, sub pubsub.Subscription) (v interface{}, ok bool, err error) {
    v, err = sub.Receive(ctx)
    ok = err != pubsub.ErrTopicClosed
    return v, ok, err
}
```

### File-system modification pub/sub

```go
type (
    fileContentSender   struct { pub pubsub.Publisher }
    fileContentReceiver struct { sub pubsub.Subscription }
)

func newFileContentSender(t pubsub.Topic) fileContentSender {
    return fileContentSender{pubsub.NewPublisher(t)}
}

func (sender fileContentSender) Send(b []byte) { sender.topic.Publish(b) }

func newFileContentReceiver(t pubsub.Topic) fileContentReceiver {
    return fileContentReceiver{pubsub.Subscribe(t)}
}

func (receiver fileContentReceiver) Receive(ctx context.Context, b *[]byte, e *error) bool {
    v, ok, err := receive(ctx, receiver.sub)
    *b, *e = v.([]byte), err
    return ok
}
```

### Configuration modification pub/sub

```go
type (
    configSender   struct { topic pubsub.Publisher }
    configReceiver struct { sub pubsub.Subscription }
)

func newConfigSender(t pubsub.Topic) configSender {
    return configSender{pubsub.NewPublisher(t)}
}

func (sender configSender) Send(cfg Configuration) { sender.topic.Publish(cfg) }

func newConfigReceiver(t pubsub.Topic) configReceiver {
    return configReceiver{pubsub.Subscribe(t)}
}

func (receiver configReceiver) Receive(ctx context.Context, c *Config, e *error) bool {
    v, ok, err := receive(ctx, receiver, sub)
    *c, *e = v.(Config), err
    return ok
}
```

### Wiring parts together

I've omitted `fileWatcher` and `configurationRepository` implementation, since
I want to show a concept. An idea behind those is:

- `fileWatcher` cares only about the fact that there was a file change. It
  probably yields file content changes (or file names if files are large).
  It seems like [fsnotify] is a suitable library to use.

- `configurationRepository` doesn't care about bytes origins, but it knows how
  to parse those into `Configuration` object (probably a protobuf or whatever
  else). It also acts as a guard to the rest of the application and ensures
  that only real configuration changes (e.g. there was at least a single field
  changed) are published.


[fsnotify]: https://github.com/fsnotify/fsnotify

```go
func main() {
    cfg := flag.String("cfg", "", "configuration file `path`")
    flag.Parse()
    // ...

    repo := pubsub.NewInprocess()

    ctx, cancel := context.WithCancel(context.Background())

    // Listen to file-system changes (probably using fsnotify), publish to the
    // file change topic.
    //
    // Note that we aren't limited with listening to the configuration file only.
    // It can be any other thing - we just care about its content.

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
    //
    // Again, this particular thing is supposed to parse out Configuration. But
    // we can use similar concept to parse and publish whatever else just by
    // injecting parsing and comparison logic.

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
    //
    // Yeah, we pass and pubsub backend repository and inject constructors (in
    // this case it's a single constructor) for the topics that we may want to
    // subscribe to.

    App{
        PubSub:          repo,
        NewConfReceiver: func() configReceiver {
            return configReceiver{pubsub.Subscribe(confChanges)}
        },
        ...
    }.Run()

    // And try to shutdown gracefully. I hope the semantic is clear enough.
    // And yes, I know, it smells a lot like Java here. But this way it's super
    // obvious w/o writing too much code.

    cancel()
    BlockOn(wg.Wait).ForMaxOf(shutdownTimeout)
}
```
