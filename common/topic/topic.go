package topic

import "context"

// Receiver is able to receive broadcast messages
type Receiver chan interface{}

type sub struct {
	rx    Receiver
	errCh chan error
}

// Topic is for broadcasting
type Topic struct {
	subCh chan sub
	msgCh chan interface{}

	subs map[Receiver]struct{}
}

// New makes a Topic
func New() *Topic {
	subCh := make(chan sub, 1)
	msgCh := make(chan interface{})

	out := &Topic{
		subCh: subCh,
		msgCh: msgCh,
		subs:  map[Receiver]struct{}{},
	}

	go out.run()

	return out
}

func (t *Topic) run() {
	closed := false
	for {
		select {
		case s := <-t.subCh:
			if !closed {
				t.subs[s.rx] = struct{}{}
			} else {
				close(s.rx)
			}
			s.errCh <- nil
		case m := <-t.msgCh:
			if m == nil {
				closed = true
				for r := range t.subs {
					close(r)
				}
				return
			}
			for r := range t.subs {
				// never wait for the receiver
				select {
				case r <- m:
				default:
					close(r)
					delete(t.subs, r)
				}
			}
		}
	}
}

// Sub subscribes a receiver
func (t *Topic) Sub(ctx context.Context, rx Receiver) error {
	errCh := make(chan error, 1)
	t.subCh <- sub{
		rx:    rx,
		errCh: errCh,
	}
	return <-errCh
}

// Send broadcasts to all receivers
func (t *Topic) Send(m interface{}) {
	t.msgCh <- m
}

// Out gives access to the broadcast channel
func (t *Topic) Out() chan<- interface{} {
	return t.msgCh
}

// Close closes all associated channels in and out
func (t *Topic) Close() {
	close(t.msgCh)
	close(t.subCh)
}
