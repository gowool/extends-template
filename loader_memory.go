package et

import (
	"context"
	"fmt"
	"sync"
)

var _ Loader = (*MemoryLoader)(nil)

type MemoryLoader struct {
	templates *sync.Map
}

func NewMemoryLoader(templates map[string][]byte) *MemoryLoader {
	t := new(sync.Map)
	for name, code := range templates {
		t.Store(name, code)
	}
	return &MemoryLoader{templates: t}
}

func (l *MemoryLoader) Add(name string, code []byte) *MemoryLoader {
	l.templates.Store(name, code)
	return l
}

func (l *MemoryLoader) Get(_ context.Context, name string) (*Source, error) {
	if code, ok := l.templates.Load(name); ok {
		return &Source{Code: code.([]byte), Name: name}, nil
	}
	return nil, fmt.Errorf(ErrNotDefinedFormat, name)
}

func (l *MemoryLoader) IsFresh(ctx context.Context, name string, _ int64) (bool, error) {
	return l.Exists(ctx, name)
}

func (l *MemoryLoader) Exists(_ context.Context, name string) (bool, error) {
	if _, ok := l.templates.Load(name); ok {
		return true, nil
	}
	return false, fmt.Errorf(ErrNotDefinedFormat, name)
}
