package et

import (
	"bytes"
	"context"
	"golang.org/x/exp/slices"
	"html/template"
	"path"
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
	global      []string
	ns          string
}

func NewTemplateWrapper(
	html *template.Template,
	loader Loader,
	reExtends *regexp.Regexp,
	reTemplates *regexp.Regexp,
	global ...string,
) *TemplateWrapper {
	t := &TemplateWrapper{
		HTML:        html,
		loader:      loader,
		reExtends:   reExtends,
		reTemplates: reTemplates,
		global:      global,
	}

	if data := strings.SplitN(html.Name(), "/", 2); len(data) == 2 && '@' == data[0][0] {
		t.ns = data[0] + "/"
	}

	return t
}

func (t *TemplateWrapper) IsFresh(ctx context.Context) (ok bool) {
	if !t.parsed.Load() {
		if err := t.Parse(ctx); err != nil {
			return ok
		}
	}

	unix := t.unix.Load()
	t.names.Range(func(key, _ any) bool {
		ok, _ = t.loader.IsFresh(ctx, key.(string), unix)
		return ok
	})
	return
}

func (t *TemplateWrapper) Parse(ctx context.Context) (err error) {
	defer func() {
		t.parsed.Store(true)
		t.unix.Store(time.Now().Unix())
	}()

	if t.orig == nil {
		if t.orig, err = t.HTML.Clone(); err != nil {
			return
		}
	} else if t.HTML, err = t.orig.Clone(); err != nil {
		return
	}

	t.names = &sync.Map{}

	n := &node{}
	if err = n.compile(ctx, t, t.HTML.Name()); err != nil {
		return
	}
	n = n.selfParent()

	for i := len(t.global) - 1; i >= 0; i-- {
		s := &node{}
		if err = s.compile(ctx, t, t.global[i]); err != nil {
			return
		}
		n.includes = slices.Insert(n.includes, 0, s)
	}

	return n.parse(t.HTML)
}

type node struct {
	source   *Source
	parent   *node
	child    *node
	includes []*node
}

func (n *node) selfParent() *node {
	if n.parent == nil {
		return n
	}
	return n.parent.selfParent()
}

func (n *node) compile(ctx context.Context, w *TemplateWrapper, name string) (err error) {
	if w.ns != "" && '@' == w.ns[0] && '@' != name[0] {
		name = w.ns + name
	}

	if n.source, err = w.loader.Get(ctx, name); err != nil {
		return
	}

	w.names.Store(name, struct{}{})

	if extends := w.reExtends.FindAllSubmatch(n.source.Code, -1); len(extends) > 0 {
		n.source.Code = w.reExtends.ReplaceAll(n.source.Code, []byte{})
		n.parent = &node{child: n}
		if err = n.parent.compile(ctx, w, toString(extends[0][1])); err != nil {
			return
		}
	}

	if includes := w.reTemplates.FindAllSubmatch(n.source.Code, -1); len(includes) > 0 {
		for _, tpl := range includes {
			include := &node{}
			if err = include.compile(ctx, w, toString(tpl[1])); err != nil {
				return
			}
			n.includes = append(n.includes, include)
			if w.ns != "" && '@' == w.ns[0] && '@' != rune(tpl[1][0]) {
				n.source.Code = bytes.Replace(n.source.Code, tpl[1], append(toBytes(w.ns), tpl[1]...), 1)
			}
		}
	}

	return
}

func (n *node) parse(t *template.Template) error {
	if _, err := t.Parse(toString(n.source.Code)); err != nil {
		return err
	}

	for _, include := range n.includes {
		if err := include.selfParent().parse(t.New(include.source.Name)); err != nil {
			return err
		}
	}

	if n.child == nil {
		return nil
	}

	name := n.child.source.Name
	if n.child.child == nil {
		d, suffix := path.Split(name)
		name = path.Join(d, "child_"+suffix)
	}

	return n.child.parse(t.New(name))
}
