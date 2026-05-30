package web

import "embed"

//go:embed all:dist
var EmbedFS embed.FS
