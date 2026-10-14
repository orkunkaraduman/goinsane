package goinsane

import (
	"context"
	"sync"
)

type ContextRunner struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	started   chan struct{}
	startedMu sync.Mutex
}

func NewContextRunner(parent context.Context) *ContextRunner {
	r := &ContextRunner{
		started: make(chan struct{}),
	}
	ctxParent := context.Background()
	if parent != nil {
		ctxParent = parent
	}
	r.ctx, r.cancel = context.WithCancel(ctxParent)
	return r
}

func (r *ContextRunner) Run(fn func(context.Context) error) (done <-chan error) {
	d := make(chan error, 1)
	err := r.run(func(ctx context.Context) {
		defer close(d)
		select {
		case <-ctx.Done():
			d <- ctx.Err()
		case <-r.started:
			d <- fn(ctx)
		}
	}, true)
	if err != nil {
		d <- err
		close(d)
	}
	return d
}

func (r *ContextRunner) RunAsync(fn func(context.Context)) error {
	return r.run(fn, true)
}

func (r *ContextRunner) RunSync(fn func(context.Context)) error {
	return r.run(fn, false)
}

func (r *ContextRunner) Start() {
	r.startedMu.Lock()
	defer r.startedMu.Unlock()
	var ok bool
	select {
	case _, ok = <-r.started:
	default:
		ok = true
	}
	if ok {
		close(r.started)
	}
}

func (r *ContextRunner) Ctx() context.Context {
	return r.ctx
}

func (r *ContextRunner) Cancel() {
	r.cancel()
}

func (r *ContextRunner) Wait() {
	<-r.ctx.Done()
	r.wg.Wait()
}

func (r *ContextRunner) Stop() {
	r.Cancel()
	r.Wait()
}

func (r *ContextRunner) run(fn func(context.Context), async bool) error {
	r.wg.Add(1)
	if e := r.ctx.Err(); e != nil {
		r.wg.Done()
		return e
	}
	f := func() {
		defer r.wg.Done()
		if fn != nil {
			fn(r.ctx)
		}
	}
	if async {
		go f()
	} else {
		f()
	}
	return nil
}
