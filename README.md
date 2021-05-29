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
type (
    fileContentSender   struct { topic pubsub.Topic }
    fileContentReceiver struct { sub pubsub.Subscriber }
)

func (sender fileContentSender) Send(b []byte) { sender.topic.Publish(b) }
func (receiver fileContentReceiver) Receive(ctx context.Context, b *[]byte, e *error) bool {
    switch buf, err := receiver.sub.Receive(ctx); {
    case errors.Is(err, pubsub.ErrTopicClosed):
        return false
    case err != nil:
        *e = err
    default:
        *b = buf.([]byte)
    }
    return true
}

type (
    configSender   struct { topic pubsub.Topic }
    configReceiver struct { sub pubsub.Subscriber }
)

func (sender configSender) Send(cfg Configuration) { sender.topic.Publish(cfg) }
func (receiver configReceiver) Receive(ctx context.Context, c *Config, e *error) bool {
    switch cfg, err := receiver.sub.Receive(ctx); {
    case errors.Is(err, pubsub.ErrTopicClosed):
        return false
    case err != nil:
        *e = err
    default:
        *c = cfg.(Config)
    }
    return true
}
```

And now we can implement our thing:

```go
func main() {
    cfg := flag.String("cfg", "", "configuration file `path`")
    flag.Parse()
    // ...

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    repo := pubsub.NewInprocess()
    fileChanges := pubsub.NewTopic(repo)
    wg.Add(1)
    go func() {
        defer wg.Done()
        fileWatcher{
            Path: *cfg,
            Send: fileContentSender{fileChanges}).Send,
        }.Run(ctx)
    }()

    confChanges := pubsub.NewTopic(repo)
    wg.Add(1)
    go func() {
        defer wg.Done()
        confWatcher{
            Recv: fileContentReceiver{fileChanges.NewSubscriber()}.Receive,
            Send: configSender{confChanges}).Send,
        }.Run(ctx)
    }()

    wg.Add(...)
    go func() {
        ...
        runOtherPartsOfApp(
            NewRecv: func() ConfRecvFunc {
                return configReceiver{confChanges.NewSubscriber()}.Receive
            },
            ...,
        }.Run(ctx)
    }()

    waitForExitCondition(cancel)
    wg.Wait()
}
```
