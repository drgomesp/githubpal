package templates

import (
	_ "embed"
)

//go:embed tpl.md
var MarkdownTemplate string
