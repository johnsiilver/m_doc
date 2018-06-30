package main

import (
	"context"
	"flag"

	"github.com/johnsiilver/m_doc/display/http"
)

var (
	port  = flag.Int("port", 8111, "The port to start the webserver on")
	debug = flag.Bool("debug", false, "If we are in debug mode")
)

func main() {
	flag.Parse()

	s, err := http.New(*debug)
	if err != nil {
		panic(err)
	}

	if err := s.Start(context.Background(), *port); err != nil {
		panic(err)
	}
}
