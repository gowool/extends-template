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

var _ Loader = (*FilesystemLoader)(nil)

const BaseNamespace = "base"

var sep = fmt.Sprintf("%c", filepath.Separator)

type FilesystemLoader struct {
	fsys   fs.FS
	paths  *sync.Map
	errors *sync.Map
	cache  *sync.Map
	mu     sync.Mutex
}

func NewFilesystemLoader(fsys fs.FS) *FilesystemLoader {
	return &FilesystemLoader{
		fsys:   fsys,
		paths:  new(sync.Map),
		errors: new(sync.Map),
		cache:  new(sync.Map),
	}
}

func (l *FilesystemLoader) Namespaces() (namespaces []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.paths.Range(func(key, _ any) bool {
		namespaces = append(namespaces, key.(string))
		return true
	})
	return
}

func (l *FilesystemLoader) Paths(namespace string) (paths []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if data, ok := l.paths.Load(namespace); ok {
		paths = append(make([]string, 0, len(data.([]string))), data.([]string)...)
	}
	return
}

func (l *FilesystemLoader) BasePrepend(path string) error {
	return l.Prepend(BaseNamespace, path)
}

func (l *FilesystemLoader) BaseAppend(path string) error {
	return l.Append(BaseNamespace, path)
}

func (l *FilesystemLoader) Prepend(namespace, p string) (err error) {
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

func (l *FilesystemLoader) Append(namespace, p string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reset()

	return l.add(namespace, p)
}

func (l *FilesystemLoader) SetPaths(namespace string, paths ...string) error {
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

func (l *FilesystemLoader) Get(_ context.Context, name string) (*Source, error) {
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

func (l *FilesystemLoader) IsFresh(_ context.Context, name string, t int64) (bool, error) {
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

func (l *FilesystemLoader) Exists(_ context.Context, name string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := l.find(name)
	return err == nil, err
}

func (l *FilesystemLoader) find(name string) (string, error) {
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

func (l *FilesystemLoader) add(namespace, p string) (err error) {
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

func (l *FilesystemLoader) normalize(s string) string {
	return strings.ReplaceAll(s, "/", sep)
}

func (l *FilesystemLoader) path(p string) (string, error) {
	p = strings.Trim(l.normalize(p), sep)

	if stat, err := fs.Stat(l.fsys, p); err != nil {
		return p, errors.Join(fmt.Errorf(ErrDirNotExistsFormat, p), err)
	} else if !stat.IsDir() {
		return p, fmt.Errorf(ErrDirNotExistsFormat, p)
	}

	return p, nil
}

func (l *FilesystemLoader) parse(name string) (string, string) {
	if data := strings.SplitN(name, "/", 2); len(data) == 2 && '@' == data[0][0] {
		return data[0][1:], data[1]
	}
	return BaseNamespace, name
}

func (l *FilesystemLoader) reset() {
	l.errors = new(sync.Map)
	l.cache = new(sync.Map)
}
