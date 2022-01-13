package messaging

import "context"

type Event interface {
	Name() string
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event Event) error
}
