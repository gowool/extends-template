package et_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	et "github.com/gowool/extends-template"
)

const (
	f1 = "file1.html"
	f2 = "file2.html"
)

func TestChainLoader_Add(t *testing.T) {
	chain := et.NewChainLoader()
	c := chain.Add(loader1{})

	assert.NotNil(t, c)
}

func TestChainLoader_Loaders(t *testing.T) {
	chain := et.NewChainLoader()

	assert.Len(t, chain.Loaders(), 0)

	chain.Add(loader1{})

	assert.Len(t, chain.Loaders(), 1)

	chain.Add(loader2{})

	assert.Len(t, chain.Loaders(), 2)
}

func TestChainLoader_Get(t *testing.T) {
	loader := et.NewChainLoader(loader1{}, loader2{})

	scenarios := []struct {
		view    string
		isError bool
	}{
		{
			view:    "no-file.html",
			isError: true,
		},
		{
			view:    f1,
			isError: false,
		},
		{
			view:    f2,
			isError: false,
		},
	}

	for _, s := range scenarios {
		source, err := loader.Get(context.TODO(), s.view)

		if s.isError {
			assert.Nil(t, source)
			assert.Error(t, err)
		} else if assert.NotNil(t, source) && assert.Nil(t, err) {
			assert.Equal(t, s.view, source.Name)
			assert.Equal(t, memData[s.view], source.Code)
			assert.Empty(t, source.File)
		}
	}
}

func TestChainLoader_IsFresh(t *testing.T) {
	loader := et.NewChainLoader(loader1{}, loader2{})

	scenarios := []struct {
		view     string
		expected bool
	}{
		{
			view:     "no-file.html",
			expected: false,
		},
		{
			view:     f1,
			expected: true,
		},
		{
			view:     f2,
			expected: true,
		},
	}

	for _, s := range scenarios {
		isFresh, _ := loader.IsFresh(context.TODO(), s.view, 0)

		assert.Equal(t, s.expected, isFresh)
	}
}

func TestChainLoader_Exists(t *testing.T) {
	loader := et.NewChainLoader(loader1{}, loader2{})

	scenarios := []struct {
		view     string
		expected bool
	}{
		{
			view:     "no-file.html",
			expected: false,
		},
		{
			view:     f1,
			expected: true,
		},
		{
			view:     f2,
			expected: true,
		},
	}

	for _, s := range scenarios {
		isFresh, _ := loader.Exists(context.TODO(), s.view)

		assert.Equal(t, s.expected, isFresh)
	}
}

type loader1 struct{}

func (loader1) Get(_ context.Context, name string) (*et.Source, error) {
	if name == f1 {
		return &et.Source{Name: f1}, nil
	}
	return nil, errors.New("not found")
}

func (loader1) IsFresh(_ context.Context, name string, _ int64) (bool, error) {
	return name == f1, nil
}

func (loader1) Exists(_ context.Context, name string) (bool, error) {
	return name == f1, nil
}

type loader2 struct{}

func (loader2) Get(_ context.Context, name string) (*et.Source, error) {
	if name == f2 {
		return &et.Source{Name: f2}, nil
	}
	return nil, errors.New("not found")
}

func (loader2) IsFresh(_ context.Context, name string, _ int64) (bool, error) {
	return name == f2, nil
}

func (loader2) Exists(_ context.Context, name string) (bool, error) {
	return name == f2, nil
}
