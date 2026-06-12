package templates

import (
	"fmt"
	"io/fs"
	"sort"
	"sync"

	builtin "github.com/qaynaq/qaynaq/templates"
)

type Catalog struct {
	templates map[string]*Manifest
	order     []string
}

var (
	defaultCatalog     *Catalog
	defaultCatalogErr  error
	defaultCatalogOnce sync.Once
)

// Default returns the catalog of built-in templates embedded in the binary.
func Default() (*Catalog, error) {
	defaultCatalogOnce.Do(func() {
		defaultCatalog, defaultCatalogErr = LoadCatalog(builtin.FS)
	})
	return defaultCatalog, defaultCatalogErr
}

func LoadCatalog(fsys fs.FS) (*Catalog, error) {
	entries, err := fs.Glob(fsys, "*.yaml")
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)

	c := &Catalog{templates: make(map[string]*Manifest, len(entries))}
	for _, name := range entries {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", name, err)
		}
		m, err := ParseManifest(data)
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", name, err)
		}
		if _, exists := c.templates[m.ID]; exists {
			return nil, fmt.Errorf("template %s: duplicate template id %q", name, m.ID)
		}
		c.templates[m.ID] = m
		c.order = append(c.order, m.ID)
	}
	return c, nil
}

func (c *Catalog) List() []*Manifest {
	out := make([]*Manifest, 0, len(c.order))
	for _, id := range c.order {
		out = append(out, c.templates[id])
	}
	return out
}

func (c *Catalog) Get(id string) (*Manifest, bool) {
	m, ok := c.templates[id]
	return m, ok
}
