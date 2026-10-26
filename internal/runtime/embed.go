package runtime

import (
	"embed"
)

//go:embed runtime.go
var EmbeddedRuntime []byte

//go:embed runtime.mod
var EmbeddedRuntimeMod []byte

//go:embed all:vendor_clean
var EmbeddedVendor embed.FS
