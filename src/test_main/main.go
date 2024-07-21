//go:build !js && !wasm
// +build !js,!wasm

/*
This file is just for unit testing purposes. The main files include a WASM-specific
import, which fails when not building for wasm, such as when running unit test.
*/
package main

func main() {

}
