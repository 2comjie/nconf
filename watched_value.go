package nconf

import (
	"sync/atomic"

	"github.com/2comjie/ntool/logx"
)

type WatchedValue[T any] struct {
	current atomic.Pointer[T]
}

func (w *WatchedValue[T]) Init(center Config, key string) error {
	value := center.Value(key)
	if value == nil {
		return ErrNotFound
	}
	if err := w.store(value); err != nil {
		return err
	}

	if err := center.Watch(key, func(_ string, _ Value) {
		if err := w.store(value); err != nil {
			logx.Errorf("config: reload watched value key=%s: %v", key, err)
		}
	}); err != nil {
		return err
	}

	return w.store(value)
}

func (w *WatchedValue[T]) store(value Value) error {
	if value.Load() == nil {
		return ErrNotFound
	}
	var current T
	if err := value.Scan(&current); err != nil {
		return err
	}
	w.current.Store(&current)
	return nil
}

func (w *WatchedValue[T]) Load() T {
	current := w.current.Load()
	if current == nil {
		var zero T
		return zero
	}
	return *current
}
