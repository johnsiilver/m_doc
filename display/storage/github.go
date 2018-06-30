package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

func init() {
	cl := http.Client{Timeout: 10 * time.Second}

	register("github://", githubReader{client: cl})
}

type githubReader struct {
	client http.Client
}

// Read implements Reader.Read().
func (g githubReader) Read(ctx context.Context, loc Location) ([]byte, error) {
	p := strings.Split(loc.Path, "/")
	if len(p) < 3 {
		return nil, fmt.Errorf("%s does not seem like a valid github path", loc.Path)
	}
	user, proj := p[0], p[1]
	projPath := strings.Join(p[2:], "/")

	r, err := g.client.Get(path.Join("http://raw.githubusercontent.com/", user, proj, "master", projPath, "mdoc"))
	if err != nil {
		return nil, fmt.Errorf("probably getting the github page: %s", err)
	}

	b := []byte{}
	_, err = io.ReadFull(r.Body, b)
	if err != nil {
		return nil, fmt.Errorf("problem reading the response from github: %s", err)
	}
	return b, nil
}
