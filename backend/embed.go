package backend

import (
	"embed"
)

// FrontendFS holds the embedded static assets from frontend_dist
//
//go:embed all:frontend_dist
var FrontendFS embed.FS
