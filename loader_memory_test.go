package et_test

import (
	et "github.com/gowool/extends-template"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var memData = map[string][]byte{
	"file.html": []byte("<p>Lorem Ipsum is simply dummy text of the printing and typesetting industry.</p>"),
}

func TestMemoryLoader_Get(t *testing.T) {
	loader := et.NewMemoryLoader(memData)

	scenarios := []struct {
		view    string
		isError bool
	}{
		{
			view:    "no-file.html",
			isError: true,
		},
		{
			view:    "file.html",
			isError: false,
		},
	}

	for _, s := range scenarios {
		source, err := loader.Get(nil, s.view)

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

func TestMemoryLoader_IsFresh(t *testing.T) {
	loader := et.NewMemoryLoader(memData)

	scenarios := []struct {
		view     string
		t        int64
		expected bool
		isError  bool
	}{
		{
			view:     "no-file.html",
			t:        time.Now().Unix(),
			expected: false,
			isError:  true,
		},
		{
			view:     "file.html",
			t:        0,
			expected: true,
			isError:  false,
		},
		{
			view:     "file.html",
			t:        time.Now().Unix(),
			expected: true,
			isError:  false,
		},
	}

	for _, s := range scenarios {
		isFresh, err := loader.IsFresh(nil, s.view, s.t)

		assert.Equal(t, s.expected, isFresh)
		if s.isError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestMemoryLoader_Exists(t *testing.T) {
	loader := et.NewMemoryLoader(memData)

	scenarios := []struct {
		view     string
		expected bool
		isError  bool
	}{
		{
			view:     "file.html",
			expected: true,
			isError:  false,
		},
		{
			view:     "no-file.html",
			expected: false,
			isError:  true,
		},
	}

	for _, s := range scenarios {
		exists, err := loader.Exists(nil, s.view)

		assert.Equal(t, s.expected, exists)
		if s.isError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
