run: build
	PORT=8080 go run ./cmd/todo-server

build:
	GOOS=js GOARCH=wasm go build -o assets/wasm/todo.wasm ./pages/todo

setup:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" "assets/scripts"