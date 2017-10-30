package main

import (
	"crypto/md5"
	"net/http"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strconv"
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

func serve(res http.ResponseWriter, req *http.Request) {
	writeJSON(res, JSON {
		"message": "Please use the route '/welcome' to log in.",
	})
}

func startSession(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to GET requests on this route.",
		})
		return
	}

	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	sessions[uuid] = Session{}
	writeJSON(res, JSON {
		"uuid": uuid,
		"message": "Welocme to GUC Carpool! Please log in using your GUC-ID and name",
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

	uuid := req.Header.Get("Authorization")
	if uuid == "" {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON {
			"message": "I'm sorry, but you don't seem to be logged in. Please log in and try again.",
		})
		return
	}

	session, sessionFound := sessions[uuid]
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

	_, loggedIn := session["gucID"]
	if !loggedIn {
		_, gucIDFound := data["gucID"]
		_, nameFound := data["name"]
		if !gucIDFound || !nameFound {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON {
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again.",
			})
			return
		} else {
			gucID := data["gucID"].(string)
			name := data["name"].(string)
			session["gucID"] = gucID
			session["name"] = name
			writeJSON(res, JSON {
				"message": "Hello " + name + ". Are you offering a ride, or are you requesting one?",
			})
			return
		}
	}

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

func createCarpoolChat(session Session, message string) (string, error) {
	comparable := strings.ToLower(message)
	fromGUC, fromGUCFound := session["fromGUC"]
	if !fromGUCFound {
		if strings.Contains(comparable, "to guc") || strings.Contains(comparable, "to the guc") || strings.Contains(comparable, "going") {
			session["fromGUC"] = false
			return "You've chosen to create a carpool that's going to the GUC. Where can you pick up ppl?", nil
		} else if strings.Contains(comparable, "from guc") || strings.Contains(comparable, "from the guc") || strings.Contains(comparable, "leaving") {
			session["fromGUC"] = true
			return "You've chosen to create a carpool that's leaving the GUC. Where are you going?", nil
		} else {
			return "I'm sorry you didn't answer my question. Are you going to the GUC or leaving the GUC?", nil
		}
	}

	_, latitudeFound := session["latitude"]
	_, longitudeFound := session["longitude"]
	if (!latitudeFound || !longitudeFound) && fromGUCFound {
		if strings.Contains(comparable, "latitude") && strings.Contains(comparable, "longitude") {
			exp := regexp.MustCompile(`[0-9]+[\.]?[0-9]*`)
			session["latitude"] = exp.FindAllString(comparable, -1)[0]
			session["longitude"] = exp.FindAllString(comparable, -1)[1]
			return "You chose the location with the latitude " + session["latitude"].(string) + ", and the longitude " + session["longitude"].(string) + ". What time would you like to your ride to be?", nil
		} else {
			var response string
			if fromGUC.(bool) {
				response = "Where are you going?"
			} else {
				response = "Where can you pick up people?"
			}
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + response)
		}
	}
	_, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "", fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
		} else {
			session["time"] = stTime
			return "You want your ride to take place around " + (session["time"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + ". How many passengers can you take with you?", nil
		}
	}

	_, availableSeats := session["availableSeats"]
	if !availableSeats && timeFound && latitudeFound && longitudeFound && fromGUCFound {
		if !(strings.Contains(comparable, "4")) && !(strings.Contains(comparable, "3")) && !(strings.Contains(comparable, "2")) && !(strings.Contains(comparable, "1")) {
			return "you can only have 1-4 passengers, not including yourself. Please enter a valid number!", nil
		} else {
			exp := regexp.MustCompile(`[1-4]`)
			number := exp.FindAllString(comparable, -1)[0]
			session["availableSeats"] = number
			return "You've chosen to take up to " + number + " more passengers.", nil
		}
	}

	// TODO DATABASE

	return "", fmt.Errorf("Whoops! An error occured in your session. Can you please log out and log back in again?")
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

func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}