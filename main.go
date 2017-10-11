// Server.go project main.go
package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", serve)
	http.HandleFunc("/alaa", alaa)
	http.ListenAndServe(":8080", nil)
}

func serve(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func alaa(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello alaa"))
}
