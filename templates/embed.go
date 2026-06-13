// Package templates holds the built-in template manifests.
// Each *.yaml file is a single template. See README.md for the format.
package templates

import "embed"

//go:embed *.yaml
var FS embed.FS
