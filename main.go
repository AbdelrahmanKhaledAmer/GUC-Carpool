// Server.go project main.go
package main

import (
	"net/http"
	"encoding/json"
)

type (
	JSON map[string]interface{}
)

func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/", serve)
	http.HandleFunc("/welcome", startSession)
	http.ListenAndServe(":8080", nil)
}

func startSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, JSON {
		"message": "Welocme to GUC Carpool!",
	})
}

func serve(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}
