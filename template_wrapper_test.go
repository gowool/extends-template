package et_test

import (
	"bytes"
	"context"
	"fmt"
	et "github.com/gowool/extends-template"
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
	"time"
)

const (
	htmlLayout   = `<body>{{block "content" .}}{{end}}</body>`
	htmlTitle    = `<h1>Title Test</h1>`
	htmlSubtitle = `<h2>Subtitle Test</h2>`
	htmlView     = `{{extends "layout.html"}}{{define "content"}}{{template "title.html"}}{{template "subtitle.html"}}{{end}}`
	htmlResult   = `<body><h1>Title Test</h1><h2>Subtitle Test</h2></body>`
)

var htmlViews = map[string][]byte{
	"@main/layout.html":   []byte(htmlLayout),
	"@main/title.html":    []byte(htmlTitle),
	"@main/subtitle.html": []byte(htmlSubtitle),
	"@main/view.html":     []byte(htmlView),
}

type wrapLoader struct {
	t int64
}

func (l wrapLoader) Get(ctx context.Context, name string) (*et.Source, error) {
	if ok, _ := l.Exists(ctx, name); ok {
		return &et.Source{Name: name, Code: htmlViews[name]}, nil
	}
	return nil, fmt.Errorf("template %s not found", name)
}

func (l wrapLoader) IsFresh(ctx context.Context, name string, t int64) (bool, error) {
	ok, _ := l.Exists(ctx, name)
	return ok && l.t < t, nil
}

func (l wrapLoader) Exists(_ context.Context, name string) (bool, error) {
	_, ok := htmlViews[name]
	return ok, nil
}

func TestTemplateWrapper_IsFresh(t *testing.T) {
	scenarios := []struct {
		t        int64
		expected bool
	}{
		{
			t:        time.Now().Add(24 * time.Hour).Unix(),
			expected: false,
		},
		{
			t:        time.Now().Add(-24 * time.Hour).Unix(),
			expected: true,
		},
	}

	for _, s := range scenarios {
		name := "@main/view.html"
		wrapper := et.NewTemplateWrapper(
			template.New(name),
			wrapLoader{t: s.t},
			et.ReExtends("{{", "}}"),
			et.ReTemplate("{{", "}}"))

		isFresh := wrapper.IsFresh(context.TODO())

		assert.Equal(t, s.expected, isFresh)
	}
}

func TestTemplateWrapper_Parse(t *testing.T) {
	name := "@main/view.html"
	wrapper := et.NewTemplateWrapper(
		template.New(name),
		wrapLoader{},
		et.ReExtends("{{", "}}"),
		et.ReTemplate("{{", "}}"))

	for range []struct{}{{}, {}} {
		if err := wrapper.Parse(context.TODO()); assert.NoError(t, err) && assert.NotNil(t, wrapper.HTML) {
			var out bytes.Buffer
			if err = wrapper.HTML.ExecuteTemplate(&out, name, nil); assert.NoError(t, err) {
				assert.Equal(t, htmlResult, out.String())
			}
		}
	}
}
