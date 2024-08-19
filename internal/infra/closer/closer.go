package closer

import (
	"context"
	"errors"
	"sync"
)

var CL *Closer

type closer func(ctx context.Context) error

// Closer is a helper to close multiple closers.
type Closer struct {
	closers  []closer
	mu       sync.Mutex
	isDone   chan struct{}
	isClosed bool
	once     sync.Once
}

// New creates a new Closer.
func New() *Closer {
	cl := &Closer{
		closers:  make([]closer, 0),
		isDone:   make(chan struct{}),
		isClosed: false,
	}

	CL = cl

	return cl
}

// Add adds a closer to the Closer.
func (c *Closer) Add(cl closer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return
	}

	c.closers = append(c.closers, cl)
}

// Done returns a channel that's closed when Close is called.
func (c *Closer) Done() chan struct{} {
	return c.isDone
}

// Close closes all closers.
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true

	defer func() {
		c.once.Do(func() {
			close(c.isDone)
		})
	}()

	var resultErr []error

	for i := len(c.closers) - 1; i >= 0; i-- {
		fn := c.closers[i]
		if err := fn(ctx); err != nil {
			resultErr = append(resultErr, err)
		}
	}

	return errors.Join(resultErr...)
}
