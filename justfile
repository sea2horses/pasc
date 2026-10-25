main_file := 'main.go'

exec_name := if os_family() == "windows" { 'pascc.exe' } else { 'pascc' }
target := f'{{exec_name}}'

build:
	go build -o {{target}} {{main_file}}

generate:
	go generate ./...

test:
	go test ./...

run file: build
	{{target}} run {{file}}
