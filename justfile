main_file := 'main.go'

exec_name := if os_family() == "windows" { 'pasc.exe' } else { 'pasc' }
build_folder := 'build'
target := f'{{build_folder}}/{{exec_name}}'

build:
	go build -o {{target}} {{main_file}}

generate:
	go generate ./...

test:
	go test ./...

run file: build
	{{target}} run {{file}}
	
