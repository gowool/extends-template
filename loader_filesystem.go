package et

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

var _ Loader = (*FileSystemLoader)(nil)

const BaseNamespace = "base"

var sep = fmt.Sprintf("%c", filepath.Separator)

type FileSystemLoader struct {
	fsys   fs.FS
	paths  *sync.Map
	errors *sync.Map
	cache  *sync.Map
	mu     sync.Mutex
}

func NewFSLoaderWithNS(fsys fs.FS) (*FileSystemLoader, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	loader := NewFileSystemLoader(fsys)
	for _, entry := range entries {
		if entry.IsDir() {
			if err = loader.SetPaths(entry.Name(), entry.Name()); err != nil {
				return nil, err
			}
		}
	}
	return loader, nil
}

func NewFileSystemLoader(fsys fs.FS) *FileSystemLoader {
	if fsys == nil {
		panic("fs.FS is nil")
	}

	return &FileSystemLoader{
		fsys:   fsys,
		paths:  new(sync.Map),
		errors: new(sync.Map),
		cache:  new(sync.Map),
	}
}

func (l *FileSystemLoader) Namespaces() (namespaces []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.paths.Range(func(key, _ any) bool {
		namespaces = append(namespaces, key.(string))
		return true
	})
	return
}

func (l *FileSystemLoader) Paths(namespace string) (paths []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if data, ok := l.paths.Load(namespace); ok {
		paths = make([]string, len(data.([]string)))
		copy(paths, data.([]string))
	}
	return
}

func (l *FileSystemLoader) BasePrepend(path string) error {
	return l.Prepend(BaseNamespace, path)
}

func (l *FileSystemLoader) BaseAppend(path string) error {
	return l.Append(BaseNamespace, path)
}

func (l *FileSystemLoader) Prepend(namespace, p string) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reset()

	if p, err = l.path(p); err != nil {
		return
	}

	var paths []string
	if data, ok := l.paths.Load(namespace); ok {
		paths = slices.Insert(data.([]string), 0, p)
	} else {
		paths = []string{p}
	}

	l.paths.Store(namespace, paths)
	return
}

func (l *FileSystemLoader) Append(namespace, p string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reset()

	return l.add(namespace, p)
}

func (l *FileSystemLoader) SetPaths(namespace string, paths ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reset()
	l.paths.Store(namespace, make([]string, 0, len(paths)))

	var err error
	for _, p := range paths {
		err = errors.Join(err, l.add(namespace, p))
	}
	return err
}

func (l *FileSystemLoader) Get(_ context.Context, name string) (*Source, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := l.find(name)
	if err != nil {
		return nil, err
	}

	code, err := fs.ReadFile(l.fsys, file)
	if err != nil {
		return nil, err
	}

	return &Source{Name: name, Code: code, File: file}, nil
}

func (l *FileSystemLoader) IsFresh(_ context.Context, name string, t int64) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := l.find(name)
	if err != nil {
		return false, err
	}

	if stat, err := fs.Stat(l.fsys, file); err != nil {
		return false, err
	} else {
		return stat.ModTime().Unix() < t, nil
	}
}

func (l *FileSystemLoader) Exists(_ context.Context, name string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.find(name)
	return err == nil, err
}

func (l *FileSystemLoader) find(name string) (string, error) {
	if p, ok := l.cache.Load(name); ok {
		return p.(string), nil
	}

	if err, ok := l.errors.Load(name); ok {
		return "", err.(error)
	}

	namespace, shortname := l.parse(name)

	var err error
	if paths, ok := l.paths.Load(namespace); ok {
		for _, p := range paths.([]string) {
			file := filepath.Join(p, shortname)
			if _, err1 := fs.Stat(l.fsys, file); err1 == nil {
				l.cache.Store(name, file)
				return file, nil
			}
		}
		err = fmt.Errorf("unable to find template \"%s\" (looked into: %s)", name, strings.Join(paths.([]string), ", "))
	} else {
		err = fmt.Errorf("there are no registered paths for namespace \"%s\"", namespace)
	}

	l.errors.Store(name, err)

	return "", err
}

func (l *FileSystemLoader) add(namespace, p string) (err error) {
	if p, err = l.path(p); err != nil {
		return
	}

	var paths []string
	if data, ok := l.paths.Load(namespace); ok {
		paths = data.([]string)
	}

	l.paths.Store(namespace, append(paths, p))
	return
}

func (l *FileSystemLoader) normalize(s string) string {
	return strings.ReplaceAll(s, "/", sep)
}

func (l *FileSystemLoader) path(p string) (string, error) {
	p = strings.Trim(l.normalize(p), sep)

	if stat, err := fs.Stat(l.fsys, p); err != nil {
		return p, errors.Join(fmt.Errorf(ErrDirNotExistsFormat, p), err)
	} else if !stat.IsDir() {
		return p, fmt.Errorf(ErrDirNotExistsFormat, p)
	}

	return p, nil
}

func (l *FileSystemLoader) parse(name string) (string, string) {
	if data := strings.SplitN(name, "/", 2); len(data) == 2 && '@' == data[0][0] {
		return data[0][1:], data[1]
	}
	return BaseNamespace, name
}

func (l *FileSystemLoader) reset() {
	l.errors = new(sync.Map)
	l.cache = new(sync.Map)
}
