// Package meta provides data structures to describe basic page controls
// for a markdown site.  All meta.Data is written as TOML.
package meta

import (
	"fmt"
	"log"

	"github.com/go-yaml/yaml"
)

// Data is used to describe metadata about the markdown documents.
type Data struct {
	// Name sets the banner for all sub documents.
	Name string
	// Style is the name of the style that you want to use.  If not set the "base" style is used.
	Style string
	// Links are a clickable link.
	Links []Link
	// Menus are menus in the banner.
	Menus []Menu
}

// MarshalYAML marshals Data into YAML output.
func (d *Data) MarshalYAML() ([]byte, error) {
	b, err := yaml.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("could not encode the Metadata: %s", err)
	}
	return b, nil
}

// UnmarshalYAML takes the bytes representing a YAML file and unmarshals it into Data.
func (d *Data) UnmarshalYAML(b []byte) error {
	log.Println("YAML file:\n", string(b))
	return yaml.Unmarshal(b, d)
}

// Link describes a link that is used in the banner/sidebar/
type Link struct {
	// Name is the name to display.
	Name string
	// URL is the link.
	URL string
}

// Menu describes a menu in the banner/sidebar.
type Menu struct {
	// Name is the name of the menu that will be displayed.
	Name string
	// Items is a list of MenuItem(s) that will be listed.
	Items []MenuItem
}

// MenuItem is an item in a banner/sidebar menu.
type MenuItem struct {
	// Name is the display name of the menu item.  This is clickable.
	Name string
	// Link is the relative path to a mdoc.  This must be a relative
	// link with a .mdoc extension.
	Link string
}
