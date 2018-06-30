package storage

import (
	"context"
	"io/ioutil"
	"path/filepath"
)

func init() {
	register("local", localReader{})
}

type localReader struct{}

// Read implements Reader.Read().
func (l localReader) Read(ctx context.Context, loc Location) ([]byte, error) {
	return ioutil.ReadFile(filepath.Clean(`./` + loc.Path))
}
