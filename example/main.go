package main

import (
	"context"
	"html/template"
	"os"
	"path/filepath"

	et "github.com/gowool/extends-template"
)

func main() {
	p, err := filepath.Abs("tests")
	if err != nil {
		panic(err)
	}
	println(p)

	fsys := os.DirFS(p)
	fsLoader := et.NewFilesystemLoader(fsys)

	if err = fsLoader.SetPaths("test_ns", "main", "base"); err != nil {
		panic(err)
	}

	loader := et.NewChainLoader(fsLoader)

	e := et.NewEnvironment(loader)
	e.Debug(true).Funcs(template.FuncMap{
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
	})

	view := "@test_ns/views/home.html"

	w, err := e.Load(context.TODO(), view)
	if err != nil {
		panic(err)
	}

	if err = w.HTML.ExecuteTemplate(os.Stdout, view, nil); err != nil {
		panic(err)
	}
}
