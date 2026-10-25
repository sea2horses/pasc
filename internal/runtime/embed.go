package runtime

import (
	_ "embed"
)

//go:embed runtime.go
var EmbeddedRuntime []byte
