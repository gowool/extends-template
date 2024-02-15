package et

import (
	"bytes"
	"context"
	"html/template"
	"path"

	"github.com/gowool/extends-template/internal"
)

type Handler func(ctx context.Context, node *Node, namespace string) error

type Node struct {
	name      string
	w         *TemplateWrapper
	Source    *Source
	Extends   *Node
	Successor *Node
	Includes  []*Node
}

func NewNode(name string, w *TemplateWrapper, successor *Node) *Node {
	if w.ns != "" && '@' == w.ns[0] && '@' != name[0] {
		name = w.ns + name
	}

	n := &Node{
		name:      name,
		w:         w,
		Successor: successor,
	}

	if successor != nil {
		successor.Extends = n
	}

	return n
}

func (n *Node) SelfParent() *Node {
	if n.Extends == nil {
		return n
	}
	return n.Extends.SelfParent()
}

func (n *Node) Init(ctx context.Context) (err error) {
	if n.Source, err = n.w.loader.Get(ctx, n.name); err != nil {
		return
	}

	n.w.names.Store(n.name, struct{}{})

	if extends := n.w.reExtends.FindAllSubmatch(n.Source.Code, -1); len(extends) > 0 {
		n.Source.Code = n.w.reExtends.ReplaceAll(n.Source.Code, []byte{})
		if err = NewNode(internal.String(extends[0][1]), n.w, n).Init(ctx); err != nil {
			return
		}
	}

	if includes := n.w.reTemplates.FindAllSubmatch(n.Source.Code, -1); len(includes) > 0 {
		for _, tpl := range includes {
			include := NewNode(internal.String(tpl[1]), n.w, nil)
			if err = include.Init(ctx); err != nil {
				return
			}
			n.Includes = append(n.Includes, include)
			if n.w.ns != "" && '@' == n.w.ns[0] && '@' != rune(tpl[1][0]) {
				n.Source.Code = bytes.Replace(n.Source.Code, tpl[1], append(internal.Bytes(n.w.ns), tpl[1]...), 1)
			}
		}
	}

	for _, h := range n.w.handlers {
		if err = h(ctx, n, n.w.ns); err != nil {
			return
		}
	}

	return
}

func (n *Node) Parse(t *template.Template) error {
	if _, err := t.Parse(internal.String(n.Source.Code)); err != nil {
		return err
	}

	for _, include := range n.Includes {
		if err := include.SelfParent().Parse(t.New(include.Source.Name)); err != nil {
			return err
		}
	}

	if n.Successor == nil {
		return nil
	}

	name := n.Successor.Source.Name
	if n.Successor.Successor == nil {
		d, suffix := path.Split(name)
		name = path.Join(d, "child_"+suffix)
	}

	return n.Successor.Parse(t.New(name))
}
