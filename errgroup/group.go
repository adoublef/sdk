package errgroup

import (
	"context"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// A Group is a collection of goroutines working on subtasks that are part of the same overall task.
type Group struct {
	g   *errgroup.Group
	ctx context.Context
}

// A New Group with a shared Context
func New(ctx context.Context, funcs ...func(context.Context) error) *Group {
	g := &Group{}
	g.g, g.ctx = errgroup.WithContext(ctx)
	g.Go(funcs...)
	return g
}

// Go calls the given function in a new goroutine. It blocks until the new goroutine can be added without the number of active goroutines in the group exceeding the configured limit.
func (g *Group) Go(funcs ...func(context.Context) error) {
	for _, f := range funcs {
		fn := f
		g.g.Go(func() error {
			return fn(g.ctx)
		})
	}
}

// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	return g.g.Wait()
}

// NotifyContext returns a copy of the parent context that is marked done when one of the listed signals arrives
func NotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
}
