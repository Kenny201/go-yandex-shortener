package closer

import (
	"context"
	"errors"
	"sync"
)

// CL Глобальная переменная для хранения экземпляра Closer по умолчанию.
var CL *Closer

// closer определяет тип функции, принимающей контекст и возвращающей ошибку.
type closer func(ctx context.Context) error

// Closer управляет списком функций-closer и предоставляет методы для их закрытия.
type Closer struct {
	closers  []closer      // Список функций-closer для вызова при закрытии
	mu       sync.Mutex    // Мьютекс для потокобезопасного доступа к списку closers
	isDone   chan struct{} // Канал для уведомления о завершении закрытия
	isClosed bool          // Флаг, указывающий, был ли Closer уже закрыт
	once     sync.Once     // Объект sync.Once для однократного вызова функции
}

// New создает и возвращает новый экземпляр Closer.
func New() *Closer {
	cl := &Closer{
		closers:  []closer{},          // Инициализация пустого списка closer-ов
		isDone:   make(chan struct{}), // Создание канала для уведомления о завершении
		isClosed: false,               // Установка флага, что Closer еще не закрыт
	}

	CL = cl // Установка глобальной переменной на созданный экземпляр

	return cl
}

// Add добавляет функцию-closer в список closers.
func (c *Closer) Add(cl closer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Если Closer уже закрыт, новые closer-ы не добавляются
	if c.isClosed {
		return
	}

	c.closers = append(c.closers, cl)
}

// Done возвращает канал, который закрывается после вызова Close.
func (c *Closer) Done() <-chan struct{} {
	return c.isDone
}

// Close вызывает все функции-closer в обратном порядке и возвращает собранные ошибки.
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Если Closer уже закрыт, ничего не делаем
	if c.isClosed {
		return nil
	}

	c.isClosed = true

	// Закрытие канала isDone один раз
	defer c.once.Do(func() {
		close(c.isDone)
	})

	var resultErr []error

	// Вызов функций-closer в обратном порядке
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i](ctx); err != nil {
			resultErr = append(resultErr, err)
		}
	}

	// Объединение и возврат всех ошибок
	return errors.Join(resultErr...)
}
