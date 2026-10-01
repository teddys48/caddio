package main

import (
	"caddy-manager"
	"io/fs"
)

func getFrontendFS() fs.FS {
	sub, err := fs.Sub(backend.FrontendFS, "frontend_dist")
	if err == nil {
		if _, err := sub.Open("index.html"); err == nil {
			return sub
		}
	}
	return nil
}
