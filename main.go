// Server.go project main.go
package main

import (
	"net/http"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
	"fmt"
	"regexp"
	"time"
	//"./DB"
)

type (
	JSON map[string]interface{}
	Session map[string]interface{}
)

var (
	sessions = map[string]Session{}
)

func getDetails(session Session) string {
	str := ""
	
	if session["fromGUC"].(bool) {
		str += "You're leaving the GUC, and going to the location with "
	} else {
		str += "You're coming to the GUC, from the location with "
	}

	str += "latitude " + session["latitude"].(string) + " and longitude " + session["longitude"].(string) + "."

	str += "You want your ride to take place around " + (session["time"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + "."

	return str
}

func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}

func processMessage(session Session, message string) (string, error) {
	requestOrCreate, requestOrCreateFound := session["requestOrCreate"]
	comparable := strings.ToLower(message)
	if !requestOrCreateFound {
		if strings.Contains(comparable, "create") || (strings.Contains(comparable, "offer") && !strings.Contains(comparable, "offered")){
			session["requestOrCreate"] = "create"
			return "You've chosen to create a carpool. Are you going to the GUC, or are you leaving campus?", nil
		} else if strings.Contains(comparable, "request") || strings.Contains(comparable, "find") || strings.Contains(comparable, "join"){
			session["requestOrCreate"] = "request"
			return "You've chosen to request a carpool. Are you going to the GUC, or are you leaving campus?", nil
		} else {
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! Are you offering a ride? Or are you requesting One?")
		}
	}else{
		if requestOrCreate == "create" {
			return createCarpoolChat(session, message)
		} else if requestOrCreate == "request" {
			return requestCarpoolChat(session, message)
		} else {
			return "", fmt.Errorf("Whoops! An error occured in your session. Can you please log out and log back in again?")
		}
	}
}

func createCarpoolChat(session Session, message string) (string, error) {
	return message, nil
}

func requestCarpoolChat(session Session, message string) (string, error) {
	fromGUC, fromGUCFound := session["fromGUC"]
	comparable := strings.ToLower(message)
	if !fromGUCFound {
		if strings.Contains(comparable, "going to") {
			session["fromGUC"] = false
			return "You've chosen to find a carpool going to the GUC! Where would you like to be picked up from?", nil
		} else if strings.Contains(comparable, "leaving") {
			session["fromGUC"] = true
			return "You chose to leave the campus. Where would you like to go?", nil
		} else {
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! Are you going to the GUC? Or are you leaving campus?")
		}
	}

	_, latitudeFound := session["latitude"]
	_, longitudeFound := session["longitude"]
	if (!latitudeFound || !longitudeFound) && fromGUCFound {
		if strings.Contains(comparable, "latitude") && strings.Contains(comparable, "longitude") {
			exp := regexp.MustCompile(`[0-9]+[\.]?[0-9]*`)
			session["latitude"] = exp.FindAllString(message,-1)[0]
			session["longitude"] = exp.FindAllString(message,-1)[1]
			return "You chose the location with the latitude " + session["latitude"].(string) + ", and the longitude " + session["longitude"].(string) + ". What time would you like to your ride to be?", nil
		} else {
			var ret string
			if fromGUC.(bool) {
				ret = "Where would you like to go?"
			} else {
				ret = "Where would you like to be picked up from?"
			}
			return "",fmt.Errorf("I'm sorry, but you didn't answer my question! " + ret)
		}
	}

	_, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "",fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
		}
		session["time"] = stTime
	}

	_, timeFound = session["time"]
	if timeFound && fromGUCFound && latitudeFound && longitudeFound {
		details := getDetails(session)
		return "Your request is complete! Here are the details: " + details + " Please wait while we find a suitable Carpool for you.", nil
	}

	return "", fmt.Errorf("Whoops! An error occured in your session. Can you please log out and log back in again?")
}

func main() {
	http.HandleFunc("/", serveAndLog(serve))
	http.HandleFunc("/welcome", serveAndLog(startSession))
	http.HandleFunc("/chat", serveAndLog(handleChat))
	http.ListenAndServe(":8080", nil)
}

func serveAndLog(handler http.HandlerFunc) http.HandlerFunc {
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
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to GET requests on this route.",
		})
		return
	}

	gucID := req.Header.Get("Authorization")
	if gucID == "" {
		res.WriteHeader(http.StatusForbidden)
		writeJSON(res, JSON {
			"message": "You don't seem to be logged in. You need to login with your GUC email and password.",
		})
		return
	}

	sessions[gucID] = Session{}
	writeJSON(res, JSON {
		"gucID": gucID,
		"message": "Welocme to GUC Carpool! Would you like to get a ride? Or are you offering one?",
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
	writeJSON(res, JSON {
		"message": "Please use the route '/welcome' to log in.",
	})
}