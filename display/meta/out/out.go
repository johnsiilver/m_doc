package main

import (
	"fmt"
	"mdoc/display/meta"
)

func main() {
	data := meta.Data{
		Name: "blah",
		Style: "base",
		Links: []meta.Link{
			{
				Name: "Owners",
				URL: "link",
			},
			{
				Name: "Authors",
				URL: "link",
			},
			
		},
	}

	b, err := data.MarshalYAML()
	if err != nil {
		panic(err)
	}
	fmt.Println("YAML is:\n", string(b))
}
