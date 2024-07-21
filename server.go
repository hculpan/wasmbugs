package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
