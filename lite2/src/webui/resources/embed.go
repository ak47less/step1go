package resources

import "embed"

// WWW contains the single-page application assets deployed by the web server.
//
//go:embed www
var WWW embed.FS
