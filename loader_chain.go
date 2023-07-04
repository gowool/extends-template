package et

import (
	"context"
	"fmt"
	"sync"
)

var _ Loader = (*ChainLoader)(nil)

type ChainLoader struct {
	loaders []Loader
	cache   *sync.Map
	mu      sync.RWMutex
}

func NewChainLoader(loaders ...Loader) *ChainLoader {
	return &ChainLoader{
		loaders: loaders,
		cache:   new(sync.Map),
	}
}

func (l *ChainLoader) Add(loader Loader) *ChainLoader {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.loaders = append(l.loaders, loader)
	l.cache = new(sync.Map)

	return l
}

func (l *ChainLoader) Loaders() (loaders []Loader) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	loaders = append(make([]Loader, 0, len(l.loaders)), l.loaders...)
	return
}

func (l *ChainLoader) Get(ctx context.Context, name string) (*Source, error) {
	r, err := l.loop(ctx, name, func(loader Loader) (any, error) {
		return loader.Get(ctx, name)
	})
	if err != nil {
		return nil, err
	}
	return r.(*Source), nil
}

func (l *ChainLoader) IsFresh(ctx context.Context, name string, t int64) (bool, error) {
	r, err := l.loop(ctx, name, func(loader Loader) (any, error) {
		return loader.IsFresh(ctx, name, t)
	})
	if err != nil {
		return false, err
	}
	return r.(bool), nil
}

func (l *ChainLoader) Exists(ctx context.Context, name string) (bool, error) {
	if r, ok := l.cache.Load(name); ok {
		return r.(bool), nil
	}
	r, err := l.loop(ctx, name, func(loader Loader) (any, error) {
		return loader.Exists(ctx, name)
	})
	if err != nil {
		l.cache.Store(name, false)
		return false, err
	}

	l.cache.Store(name, r.(bool))
	return r.(bool), nil
}

func (l *ChainLoader) loop(ctx context.Context, name string, fn func(loader Loader) (any, error)) (any, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var err error

	for _, loader := range l.loaders {
		if ok, err1 := loader.Exists(ctx, name); !ok {
			err = merge(err, err1)
			continue
		}

		if r, err1 := fn(loader); err1 == nil {
			return r, nil
		} else {
			err = merge(err, fmt.Errorf("[%s]: %w", typeName(loader), err1))
		}
	}

	return nil, errorf(err, ErrNotDefinedFormat, name)
}
