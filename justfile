main_file := 'main.go'

exec_name := if os_family() == "windows" { 'pascc.exe' } else { 'pascc' }
target := f'{{exec_name}}'

build:
	go build -o {{target}} {{main_file}}

generate:
	go generate ./...

test:
	go test ./...
	./coverage.sh

run file: build
	{{target}} run {{file}}

release_targets := "linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64"

release:
	#!/usr/bin/env bash
	set -euo pipefail
	rm -rf dist && mkdir dist
	for t in {{release_targets}}; do
		os=${t%/*}; arch=${t#*/}
		ext=""; [ "$os" = windows ] && ext=".exe"
		name="pascc-$os-$arch"
		echo "building $name"
		GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "dist/$name/pascc$ext" .
		cp README.md LICENSE "dist/$name/"
		if [ "$os" = windows ]; then bsdtar -a -cf "dist/$name.zip" -C dist "$name"
		else tar -czf "dist/$name.tar.gz" -C dist "$name"; fi
	done
	(cd dist && sha256sum *.zip *.tar.gz > checksums.txt)
