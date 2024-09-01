package et

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gowool/extends-template/internal"
)

const (
	leftDelim  = "{{"
	rightDelim = "}}"
)

type Environment struct {
	debug      bool
	global     []string
	left       string
	right      string
	loader     Loader
	handlers   []Handler
	reExtends  *regexp.Regexp
	reTemplate *regexp.Regexp
	templates  *sync.Map
	funcMap    template.FuncMap
	hash       atomic.Value
	mu         sync.Mutex
}

func NewEnvironment(loader Loader, handlers ...Handler) *Environment {
	e := &Environment{loader: loader, handlers: handlers, funcMap: template.FuncMap{}}

	return e.Delims(leftDelim, rightDelim)
}

func (e *Environment) Debug(debug bool) *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.debug != debug {
		e.debug = debug
		e.updateHash()
	}

	return e
}

func (e *Environment) Delims(left, right string) *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.left = left
	e.right = right
	e.reExtends = internal.ReExtends(left, right)
	e.reTemplate = internal.ReTemplate(left, right)
	e.updateHash()

	return e
}

func (e *Environment) Funcs(funcMap template.FuncMap) *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	for k, v := range funcMap {
		e.funcMap[k] = v
	}
	e.updateHash()

	return e
}

func (e *Environment) Global(global ...string) *Environment {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.global = make([]string, len(global))
	copy(e.global, global)
	e.updateHash()

	return e
}

func (e *Environment) NewHTMLTemplate(name string) *template.Template {
	return template.New(name).Delims(e.left, e.right).Funcs(e.funcMap)
}

func (e *Environment) NewTemplateWrapper(name string) *TemplateWrapper {
	return NewTemplateWrapper(
		e.NewHTMLTemplate(name),
		e.loader,
		e.handlers,
		e.reExtends,
		e.reTemplate,
		e.global...,
	)
}

func (e *Environment) Load(ctx context.Context, name string) (*TemplateWrapper, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var wrapper *TemplateWrapper

	key := e.key(name)

	v, ok := e.templates.Load(key)
	if ok {
		wrapper = v.(*TemplateWrapper)
	} else {
		wrapper = e.NewTemplateWrapper(name)

		e.templates.Store(key, wrapper)
	}

	if !ok || e.debug || !wrapper.IsFresh(ctx) {
		if err := wrapper.Parse(ctx); err != nil {
			return nil, err
		}
	}
	return wrapper, nil
}

func (e *Environment) updateHash() {
	var buf bytes.Buffer

	buf.WriteString(e.left)
	buf.WriteString(e.right)
	buf.WriteString(strconv.FormatBool(e.debug))
	for _, s := range e.global {
		buf.WriteString(s)
	}
	for name := range e.funcMap {
		buf.WriteString(name)
	}

	e.hash.Store(internal.Hash(buf.Bytes()))
	e.templates = new(sync.Map)
}

func (e *Environment) key(name string) string {
	return internal.Hash(internal.Bytes(fmt.Sprintf("%s:%s", name, e.hash.Load())))
}
