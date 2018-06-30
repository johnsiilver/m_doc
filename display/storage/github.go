package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func init() {
	cl := http.Client{Timeout: 10 * time.Second}

	register("github", githubReader{client: cl})
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

	u, err := urlJoin(user, proj, "master", projPath)
	if err != nil {
		return nil, err
	}

	log.Println("github fetch: ", u)
	r, err := g.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("probably getting the github page: %s", err)
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("problem reading the response from github: %s", err)
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("no content at: %s", u)
	}
	return b, nil
}

func urlJoin(p ...string) (string, error) {
	j := path.Join(p...)
	u, err := url.Parse(j)
	if err != nil {
		return "", err
	}
	base, err := url.Parse("https://raw.githubusercontent.com/")
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}
