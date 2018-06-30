// Package http provides an HTTP server for serving up markdown pages with
// specialized banners and menus defined by a custom style.
package http

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mdoc/display/meta"
	"mdoc/display/storage"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var templFns = template.FuncMap{
	"join": strings.Join,
	"sub": func(y, x int) int {
		return x - y
	},
}

// Server provides an HTTP server.
type Server struct {
	store  storage.Reader
	server http.Server
	mux    *http.ServeMux

	mu sync.Mutex
}

// New is the contructor for Server.
func New(debug bool) (*Server, error) {
	s, err := storage.NewReader()
	if err != nil {
		return nil, err
	}

	h := &Server{
		server: http.Server{
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 1 * time.Second,
			WriteTimeout:      5 * time.Second,
			MaxHeaderBytes:    1024,
		},
		store: s,
	}
	h.mux = http.NewServeMux()
	h.mux.HandleFunc("/", h.display)
	h.server.Handler = h.mux
	if err := h.registerStatic(debug); err != nil {
		return nil, err
	}
	return h, nil
}

var staticExts = regexp.MustCompile(`.css`)

// TODO(jdoak): This is in the bits server too, should be made into a standard lib.
func (s *Server) registerStatic(debug bool) error {
	err := filepath.Walk(
		"styles",
		func(path string, f os.FileInfo, err error) error {
			log.Println("checking path for static file: ", path)
			path = strings.Replace(path, `\`, `/`, -1)
			if err != nil {
				return err
			}
			if f.IsDir() {
				log.Printf("%s: is dir, skipping", path)
				return nil
			}

			if staticExts.MatchString(path) {
				log.Printf("register: %s", fmt.Sprintf("/%s", path))
				var hf http.HandlerFunc
				var err error
				// Reload content on every request.
				if debug {
					hf, err = debugHandleFunc(path)
					if err != nil {
						return err
					}
					// Load the context only once on load.
				} else {
					hf, err = handleFunc(path)
					if err != nil {
						return err
					}
				}
				urlPath := fmt.Sprintf("/%s", path)
				s.mux.HandleFunc(urlPath, hf)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("problems walking our web/ filepath: %s", err)
	}
	return nil
}

// Start starts the webserver on port.  Blocks until the server stops.
func (s *Server) Start(ctx context.Context, port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server.Addr != "" {
		return errors.New("the server is already running")
	}
	s.server.Addr = fmt.Sprintf(":%d", port)
	return s.server.ListenAndServe()
}

// Close stops the webserver.  If no server is running, this is a no-op.
func (s *Server) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server.Addr == "" {
		return nil
	}

	err := s.server.Close()
	s.server.Addr = ""
	return err
}

// display handles displaying the content for a page.
func (s *Server) display(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	if strings.Count(r.URL.Path, "mdoc/") != 1 {
		http.Error(w, fmt.Sprintf("there must be a single mdoc/ in the path, we had: %s", r.URL.Path), http.StatusNotAcceptable)
		return
	}
	sp := strings.SplitAfter(r.URL.Path, "mdoc/")

	metaData, err := s.getMetaData(r.Context(), storage.Location{filepath.Join(sp[0], metaFile)})
	if err != nil {
		http.Error(w, fmt.Sprintf("had problem locating meta file: %s", err), http.StatusInternalServerError)
		return
	}
	log.Println(metaData)

	tmplName := metaData.Style + ".gotmpl"
	tmpl := template.New("").Funcs(templFns)
	tmpl, err = tmpl.ParseFiles(fmt.Sprintf("./styles/%s/%s", metaData.Style, tmplName))
	if err != nil {
		http.Error(w, fmt.Sprintf("could not find style %s.gotmpl: %s", metaData.Style, err), http.StatusInternalServerError)
		return
	}

	var content []byte
	switch {
	// We want the index.
	case strings.HasSuffix(r.URL.Path, "mdoc/"):
		content, err = s.store.Read(r.Context(), storage.Location{r.URL.Path + "index.mdoc"})
		if err != nil {
			http.Error(w, fmt.Sprintf("could not find index.mdoc for location %s", r.URL.Path), http.StatusInternalServerError)
			return
		}
	// We want a specific page.
	case strings.HasSuffix(r.URL.Path, ".mdoc"):
		content, err = s.store.Read(r.Context(), storage.Location{r.URL.Path})
		if err != nil {
			http.Error(w, fmt.Sprintf("could not find %s", r.URL.Path), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("we don't know how to deal with %s", r.URL.Path), http.StatusNotAcceptable)
		return
	}

	md := blackfriday.Run(content)
	//p := bluemonday.UGCPolicy()
	//p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")

	pass := Passthru{
		Meta:     metaData,
		Home:     sp[0],
		Auth:     Auth{User: "some_user"},
		Markdown: template.HTML(md),
	}

	if err := tmpl.ExecuteTemplate(w, tmplName, pass); err != nil {
		http.Error(w, fmt.Sprintf("problem rendering template: %s", err), http.StatusInternalServerError)
		return
	}
}

const metaFile = "meta"

func (s *Server) getMetaData(ctx context.Context, loc storage.Location) (meta.Data, error) {
	d := meta.Data{}

	b, err := s.store.Read(ctx, loc)
	if err != nil {
		return d, fmt.Errorf("could not find metafile at: %s: %s", loc.Path, err)
	}

	err = d.UnmarshalYAML(b)
	return d, err
}

// Auth contains authorization information.
type Auth struct {
	// User is the user name of the user navigating the page.
	User string
}

// Passthru contains information used in the rendering of an mdoc page.
type Passthru struct {
	// Meta contains information from the site's meta file.
	Meta meta.Data
	// Home is the root URL of the mdoc.
	Home string
	// Auth contains authorization information.
	Auth Auth // TODO(jdoak): Replace with Oath stuff.
	// Markdown contains the markdown to render on the page.
	Markdown template.HTML
}

// debugHandleFunc reloads the content of static files on every load.
// Should only be used when debugging the UI.
func debugHandleFunc(path string) (http.HandlerFunc, error) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not read in file %s: %s", path, err), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(b); err != nil {
			http.Error(w, fmt.Sprintf("could not write to http stream file %s: %s", path, err), http.StatusInternalServerError)
		}
	}, nil
}

// handleFunc only loads the static content a single time.
// Used for production.
func handleFunc(path string) (http.HandlerFunc, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read in file %s: %s", path, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(b); err != nil {
			http.Error(w, fmt.Sprintf("could not write to http stream file %s: %s", path, err), http.StatusInternalServerError)
		}
	}, nil
}
