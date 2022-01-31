package main

import (
	server "github.com/nndi-oss/ussdproxy/server"
)

func main() {
	err := server.ListenAndServe()
}
