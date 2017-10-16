// Server.go project main.go
package main

import (
	"net/http"
	"encoding/json"
	"log"
	"net/http/httptest"
)

type (
	JSON map[string]interface{}
	//Session map[string]interface{}
)

func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}

func main() {
	http.HandleFunc("/", serveAndLog(serve))
	http.HandleFunc("/welcome", serveAndLog(startSession))
	http.HandleFunc("/chat", serveAndLog(chatBot))
	http.ListenAndServe(":8080", nil)
}

func serveAndLog(handler http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, req *http.Request) {
		res := httptest.NewRecorder()
		handler(res, req)
		log.Printf("[%d] %-4s %s\n", res.Code, req.Method, req.URL.Path)

		for k, v := range res.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(res.Code)
		res.Body.WriteTo(w)
	}
}

func startSession(res http.ResponseWriter, req *http.Request) {
	// Ask about uuid
	writeJSON(res, JSON {
		"message": "Welocme to GUC Carpool!",
	})
}

func chatBot(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to POST requests on this route.",
		})
		return
	}

	// Some logic for the session should go here

	data := JSON{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON {
			"message": "I could not understand what you said because it wasn't written in a JSON format!",
		})
		return
	}
	// Closed in order to stop resource leak.
	defer req.Body.Close()

	message, received := data["message"]
	if !received {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON {
			"message": "I did not receive a message. Are you sure you sent me something?",
		})
	}
	log.Println(message)
}

func serve(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("hello world"))
}
