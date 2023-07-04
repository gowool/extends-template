package et

import "context"

const (
	ErrNotDefinedFormat   = "template \"%s\" is not defined"
	ErrDirNotExistsFormat = "the \"%s\" directory does not exist"
)

type Source struct {
	Code []byte
	Name string
	File string
}

type Loader interface {
	// Get returns a Source for a given template name
	Get(ctx context.Context, name string) (*Source, error)

	// IsFresh check if template is fresh
	IsFresh(ctx context.Context, name string, t int64) (bool, error)

	// Exists check if template exists
	Exists(ctx context.Context, name string) (bool, error)
}
