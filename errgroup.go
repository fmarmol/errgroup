package errgroup

import (
	"context"
	"sync"
)

type Group struct {
	ctx     context.Context
	cancel  func()
	wg      sync.WaitGroup
	err     error
	errOnce sync.Once
}

func NewGroup(ctx context.Context) (*Group, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	g := &Group{
		ctx:    ctx,
		cancel: cancel,
	}
	return g, ctx
}

func (g *Group) Wait() error {
	g.wg.Wait()
	return g.err
}

// Go add f to the pool of functions to execute. Be careful f must stop on cancel context, to not have goroutines leaks
func (g *Group) Go(f func(ctx context.Context) error, callbacks ...func()) {
	g.wg.Add(1)
	go func() {
		defer func() {
			g.wg.Done()
			for _, callback := range callbacks {
				callback()
			}
		}()
		defer func() {
			if g.cancel != nil {
				g.cancel()
			}
		}()
		err := f(g.ctx)
		g.errOnce.Do(func() {
			g.err = err
		})
	}()
}
