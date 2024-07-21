build:
	GOOS=js GOARCH=wasm go build -o main.wasm src/wasm/*.go

run: build
	go build -o server server.go

package: build
	zip -9 wasmbugs.zip index.html wasm_exec.js main.wasm styles.css bugs-logo.png bugs-favicon.ico

clean:
	rm -f server
	rm -f main.wasm
	rm -rf tmp