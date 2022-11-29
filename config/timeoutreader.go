package config

import (
	"context"
	"io"
)

type TimeoutReader struct {
	ctx context.Context
	r   io.ReadCloser
}

func (t *TimeoutReader) Async() {
	<-t.ctx.Done()
	t.r.Close()
}

func (t *TimeoutReader) Read(p []byte) (n int, err error) {
	// TODO: Maybe do check on r to see if we can bypass doing it ourselves?
	return t.r.Read(p)
}

func NewTimeoutReader(ctx context.Context, r io.ReadCloser) *TimeoutReader {
	t := &TimeoutReader{ctx: ctx, r: r}
	go t.Async()
	return t
}
