package assets

import (
	"embed"
)

//go:embed "emails" "migrations" "templates" "static" "js" "css"
var EmbeddedFiles embed.FS
