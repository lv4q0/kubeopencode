// Copyright Contributors to the KubeOpenCode project

// Package opencodeapp provides embedded OpenCode Web UI static assets.
// The dist/ directory is populated by "make opencode-app-build" which builds
// the OpenCode Web UI from source. When dist/ only contains a placeholder,
// the agent web handler falls back to proxying app.opencode.ai through the
// OpenCode server.
package opencodeapp

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var assets embed.FS

// Assets returns the embedded dist filesystem.
// Returns nil if only the placeholder .gitkeep exists (not built yet).
func Assets() fs.FS {
	distFS, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil
	}

	// Check if index.html exists — if not, assets haven't been built
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		return nil
	}

	return distFS
}
