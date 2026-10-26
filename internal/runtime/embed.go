package runtime

import (
	"embed"
)

//go:embed runtime.go
var EmbeddedRuntime []byte

//go:embed ext.go
var EmbeddedExtRuntime []byte

//go:embed ext_fast.go
var EmbeddedExtFastRuntime []byte

//go:embed runtime.mod
var EmbeddedRuntimeMod []byte

//go:embed all:vendor_clean
var EmbeddedVendor embed.FS
