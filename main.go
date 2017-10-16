// Server.go project main.go
package main

import (
	"net/http"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
)

type (
	JSON map[string]interface{}
	Session map[string]interface{}
)

var (
	sessions = map[string]Session{}
)

func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}

func processMessage(session Session, message string) (string, error){
	// Some session logic should go here
	message = strings.ToLower(message)
	return message, nil
}

func main() {
	http.HandleFunc("/", serveAndLog(serve))
	http.HandleFunc("/welcome", serveAndLog(startSession))
	http.HandleFunc("/chat", serveAndLog(handleChat))
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
	gucID := req.Header.Get("Authorization")
	sessions[gucID] = Session{}
	writeJSON(res, JSON {
		"gucID": gucID,
		"message": "Welocme to GUC Carpool! Would you like to get a ride to university? Or are you offering one?",
	})
}

func handleChat(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to POST requests on this route.",
		})
		return
	}

	gucID := req.Header.Get("Authorization")
	if gucID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you don't seem to be logged in. Please log in and try again.",
		})
		return
	}

	session, sessionFound := sessions[gucID]
	if !sessionFound {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON {
			"message": "I'm sorry, but your session has expired. Please log in and try again.",
		})
		return
	}

	data := JSON{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON {
			"message": "I could not understand what you said because it wasn't written in a JSON format!",
		})
		return
	}
	defer req.Body.Close()

	_, received := data["message"]
	if !received {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON {
			"message": "I did not receive a message. Are you sure you sent me something?",
		})
		return
	}

	finalResponse, err := processMessage(session, data["message"].(string))
	if err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(res, JSON {
			"message": string(err.Error()),
		})
		return
	}

	writeJSON(res, JSON {
		"message": finalResponse,
	})
}

func serve(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("hello world"))
}