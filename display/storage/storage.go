// Package storage provide storage methods.
package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Location details where a file is located.
type Location struct {
	// Path is the path to the file.
	Path string
}

// Reader provides a file reader.
type Reader interface {
	// Read reads a file from loc. loc.Path must not have the header.
	Read(ctx context.Context, loc Location) ([]byte, error)
}

// NewReader provides a Reader for all supported location paths.
func NewReader() (Reader, error) {
	return multiReader{}, nil
}

type multiReader struct{}

// Read implements Reader.Read.
func (m multiReader) Read(ctx context.Context, loc Location) ([]byte, error) {
	log.Println("loc.Path: ", loc.Path)
	p := strings.Split(loc.Path, "/")
	log.Println("after split: ", p)
	if len(p) < 3 {
		return nil, fmt.Errorf("Location.Path does not contain a valid header: %s", loc.Path)
	}

	r := registry[p[1]]
	if r == nil {
		log.Println(registry)
		return nil, fmt.Errorf("cannot call Read for a location: %v is not registered", p[1])
	}

	return r.Read(ctx, Location{strings.Join(p[2:], "/")})
}

var registry = map[string]Reader{}

func register(h string, r Reader) error {
	if _, ok := registry[h]; ok {
		return fmt.Errorf("Medium %v has already been registered", h)
	}
	registry[h] = r
	return nil
}
