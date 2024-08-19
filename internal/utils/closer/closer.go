package closer

import (
	"context"
	"errors"
	"sync"
)

// CL Глобальная переменная для хранения экземпляра Closer по умолчанию.
var CL = New()

// Closer управляет списком функций-closer и предоставляет методы для их закрытия.
type Closer struct {
	closers  []func(ctx context.Context) error
	mu       sync.Mutex
	isClosed bool
	done     chan struct{}
}

// New создает и возвращает новый экземпляр Closer.
func New() *Closer {
	return &Closer{
		done: make(chan struct{}),
	}
}

// Add добавляет функцию-closer в список closers.
func (c *Closer) Add(cl func(ctx context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isClosed {
		c.closers = append(c.closers, cl)
	}
}

// Done возвращает канал, который закрывается после вызова Close.
func (c *Closer) Done() <-chan struct{} {
	return c.done
}

// Close вызывает все функции-closer в обратном порядке и возвращает собранные ошибки.
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true
	close(c.done)

	var errs []error
	// Обратный порядок вызова функций-closer
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
