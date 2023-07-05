package et

import (
	"context"
	"golang.org/x/exp/slices"
	"html/template"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type TemplateWrapper struct {
	HTML        *template.Template
	orig        *template.Template
	reExtends   *regexp.Regexp
	reTemplates *regexp.Regexp
	names       *sync.Map
	parsed      atomic.Bool
	unix        atomic.Int64
	loader      Loader
	handlers    []Handler
	global      []string
	ns          string
}

func NewTemplateWrapper(
	html *template.Template,
	loader Loader,
	handlers []Handler,
	reExtends *regexp.Regexp,
	reTemplates *regexp.Regexp,
	global ...string,
) *TemplateWrapper {
	w := &TemplateWrapper{
		HTML:        html,
		loader:      loader,
		handlers:    handlers,
		reExtends:   reExtends,
		reTemplates: reTemplates,
		global:      global,
	}

	if data := strings.SplitN(html.Name(), "/", 2); len(data) == 2 && '@' == data[0][0] {
		w.ns = data[0] + "/"
	}

	return w
}

func (w *TemplateWrapper) IsFresh(ctx context.Context) (ok bool) {
	if !w.parsed.Load() {
		if err := w.Parse(ctx); err != nil {
			return ok
		}
	}

	unix := w.unix.Load()
	w.names.Range(func(key, _ any) bool {
		ok, _ = w.loader.IsFresh(ctx, key.(string), unix)
		return ok
	})
	return
}

func (w *TemplateWrapper) Parse(ctx context.Context) (err error) {
	defer func() {
		w.parsed.Store(true)
		w.unix.Store(time.Now().Unix())
	}()

	if w.orig == nil {
		if w.orig, err = w.HTML.Clone(); err != nil {
			return
		}
	} else if w.HTML, err = w.orig.Clone(); err != nil {
		return
	}

	w.names = &sync.Map{}

	node := NewNode(w.HTML.Name(), w, nil)
	if err = node.Init(ctx); err != nil {
		return
	}
	node = node.SelfParent()

	for i := len(w.global) - 1; i >= 0; i-- {
		globalNode := NewNode(w.global[i], w, nil)
		if err = globalNode.Init(ctx); err != nil {
			return
		}
		node.Includes = slices.Insert(node.Includes, 0, globalNode)
	}

	return node.Parse(w.HTML)
}
