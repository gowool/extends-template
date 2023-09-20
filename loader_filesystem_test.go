package et_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	et "github.com/gowool/extends-template"
)

func newFilesystemLoader() *et.FilesystemLoader {
	return et.NewFilesystemLoader(os.DirFS("./tests"))
}

func TestFilesystemLoader_Namespaces(t *testing.T) {
	expected := "base"

	loader := newFilesystemLoader()

	assert.Empty(t, loader.Namespaces())

	_ = loader.BaseAppend(expected)
	ns := loader.Namespaces()

	if assert.Len(t, ns, 1) {
		assert.Equal(t, et.BaseNamespace, ns[0])
	}
}

func TestFilesystemLoader_Prepend(t *testing.T) {
	expected := "main"

	loader := newFilesystemLoader()
	_ = loader.BaseAppend("base")
	err := loader.BasePrepend(expected)

	if assert.NoError(t, err) {
		paths := loader.Paths(et.BaseNamespace)

		assert.Len(t, paths, 2)
		assert.Contains(t, paths, expected)
	}
}

func TestFilesystemLoader_AppendError(t *testing.T) {
	loader := newFilesystemLoader()

	err := loader.BaseAppend("no-path")

	assert.Error(t, err)
}

func TestFilesystemLoader_SetPaths(t *testing.T) {
	expected := []string{"main", "base"}

	loader := newFilesystemLoader()

	assert.Empty(t, loader.Paths(et.BaseNamespace))

	err := loader.SetPaths(et.BaseNamespace, expected...)

	if assert.NoError(t, err) {
		paths := loader.Paths(et.BaseNamespace)

		assert.Len(t, paths, len(expected))

		for _, p := range paths {
			assert.Contains(t, expected, p)
		}
	}
}

func TestFilesystemLoader_Get(t *testing.T) {
	loader := newFilesystemLoader()
	_ = loader.BaseAppend("base")

	scenarios := []struct {
		view    string
		isError bool
	}{
		{
			view:    "views/no-home.html",
			isError: true,
		},
		{
			view:    "views/home.html",
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
		}
	}
}

func TestFilesystemLoader_Exists(t *testing.T) {
	loader := newFilesystemLoader()
	_ = loader.BaseAppend("base")

	scenarios := []struct {
		view     string
		expected bool
		isError  bool
	}{
		{
			view:     "views/home.html",
			expected: true,
			isError:  false,
		},
		{
			view:     "views/no-home.html",
			expected: false,
			isError:  true,
		},
	}

	for _, s := range scenarios {
		exists, err := loader.Exists(context.TODO(), s.view)

		assert.Equal(t, s.expected, exists)
		if s.isError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestFilesystemLoader_IsFresh(t *testing.T) {
	loader := newFilesystemLoader()
	_ = loader.BaseAppend("base")

	scenarios := []struct {
		view     string
		t        int64
		expected bool
		isError  bool
	}{
		{
			view:     "views/no-home.html",
			t:        time.Now().Unix(),
			expected: false,
			isError:  true,
		},
		{
			view:     "views/home.html",
			t:        0,
			expected: false,
			isError:  false,
		},
		{
			view:     "views/home.html",
			t:        time.Now().Unix(),
			expected: true,
			isError:  false,
		},
	}

	for _, s := range scenarios {
		isFresh, err := loader.IsFresh(context.TODO(), s.view, s.t)

		assert.Equal(t, s.expected, isFresh)
		if s.isError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
