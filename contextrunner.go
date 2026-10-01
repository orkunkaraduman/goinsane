package goinsane

import (
	"context"
	"sync"
)

type ContextRunner struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewContextRunner(parent context.Context) *ContextRunner {
	r := &ContextRunner{}
	ctxParent := context.Background()
	if parent != nil {
		ctxParent = parent
	}
	r.ctx, r.cancel = context.WithCancel(ctxParent)
	return r
}

func (r *ContextRunner) RunAsync(fn func(context.Context)) error {
	return r.run(fn, true)
}

func (r *ContextRunner) RunSync(fn func(context.Context)) error {
	return r.run(fn, false)
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
