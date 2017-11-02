package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GUC-Carpool/DB"
)

type (
	// JSON models a json for sending and recieving in requests and responses
	JSON map[string]interface{}
	// Session models the session of a user
	Session map[string]interface{}
)

var (
	sessions = map[string]Session{}
)

// Main function to start the server and handle all incoming routes.
func main() {
	http.HandleFunc("/", serveAndLog(serve))
	http.HandleFunc("/welcome", serveAndLog(startSession))
	http.HandleFunc("/chat", serveAndLog(handleChat))
	http.ListenAndServe(":8080", nil)
}

// Intermediary function that logs the current request and the status code attached to the response.
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

// Default route handler.
func serve(res http.ResponseWriter, req *http.Request) {
	writeJSON(res, JSON{
		"message": "Please use the route '/welcome' to log in, and the route '/chat' to talk.",
	})
}

// Function that creates the session variable and attaches a uuid (Unique user ID) to it.
func startSession(res http.ResponseWriter, req *http.Request) {
	// Only listen to GET requests.
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to GET requests on this route.",
		})
		return
	}

	// Create a new uuid
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	// Create a new session mapped to the new uuid and reply to the user.
	sessions[uuid] = Session{}
	writeJSON(res, JSON{
		"uuid":    uuid,
		"message": "Welocme to GUC Carpool! Please log in using your GUC-ID and name separated by the delimiter ':'. Make sure to do this first on the '/chat' route.",
	})
}

// Function to handle the chat route
func handleChat(res http.ResponseWriter, req *http.Request) {
	// Only listen to POST requests
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to POST requests on this route.",
		})
		return
	}

	// Make sure the user has a session.
	uuid := req.Header.Get("Authorization")
	if uuid == "" {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you don't seem to be logged in. Please log in and try again.",
		})
		return
	}

	// Make sure the user's session is active.
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON{
			"message": "I'm sorry, but your session has expired. Please log in and try again.",
		})
		return
	}

	// Make sure the data sent is in a JSON format.
	data := JSON{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON{
			"message": "I could not understand what you said because it wasn't written in a JSON format!",
		})
		return
	}
	defer req.Body.Close()

	// Make sure the data sent is in "message"
	messageRecieved, received := data["message"]
	if !received {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON{
			"message": "I did not receive a message. Are you sure you sent me something?",
		})
		return
	}

	// If user is not logged in, log them in.
	_, loggedIn := session["gucID"]
	if !loggedIn {
		// Separate gucID and name
		login := strings.Split(messageRecieved.(string), ":")
		if len(login) < 2 {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again.",
			})
			return
		}
		gucID := login[0]
		name := login[1]
		// Check if gucID and name are empty
		if gucID == "" || name == "" {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again.",
			})
			return
		}
		// Check if gucId is in a valid format (eg. 13-2456)
		exp := regexp.MustCompile(`[0-9]+-[0-9]+`)
		match := exp.MatchString(gucID)
		if !match {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Your GUC ID is invalid. Are you sure you entered it correctly?",
			})
			return
		}
		// Find if an old session has the same user. If found, migrate the information from the old session to the new one, and delete the old one.
		for key1, val1 := range sessions {
			currentGucID, currentIDFound := val1["gucID"]
			if currentIDFound && strings.EqualFold(currentGucID.(string), gucID) && uuid != key1 {
				for key2, val2 := range val1 {
					session[key2] = val2
				}
				delete(sessions, key1)
			}
		}
		session["gucID"] = gucID
		session["name"] = name
		writeJSON(res, JSON{
			"message": "Hello " + name + ". You can view all available carpools by typing 'view all', cancel your request by typing 'cancel request', edit your request by typing 'edit request' or choose an available carpool by typing 'choose ID' where ID is the postID of the carpool of your choice. You can also choose to offer other people a ride by creating a carpool by typing 'create' or 'offer', or specify the details of a carpool you wish to request by typing 'request', 'find' or 'join'.",
		})
		return
	}

	// See if the user wishes to interact with data from the database or edit his session.
	comparable := strings.ToLower(messageRecieved.(string))
	if strings.Contains(comparable, "edit") || strings.Contains(comparable, "cancel") || strings.Contains(comparable, "choose") || strings.Contains(comparable, "view all") {
		postRequestHandler(res, session, data)
		return
	}

	// See if user wishes to request or create a carpool.
	finalResponse, err := processMessage(session, data["message"].(string))
	if err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(res, JSON{
			"message": string(err.Error()),
		})
		return
	}

	writeJSON(res, JSON{
		"message": finalResponse,
	})
}

func processMessage(session Session, message string) (string, error) {
	requestOrCreate, requestOrCreateFound := session["requestOrCreate"]
	comparable := strings.ToLower(message)
	if !requestOrCreateFound {
		if strings.Contains(comparable, "create") || (strings.Contains(comparable, "offer")) {
			session["requestOrCreate"] = "create"
			return "You've chosen to create a carpool. If you are going to the GUC, please type 'to guc', if you are leaving the GUC, please type 'from guc'.", nil
		} else if strings.Contains(comparable, "request") || strings.Contains(comparable, "find") || strings.Contains(comparable, "join") {
			session["requestOrCreate"] = "request"
			return "You've chosen to request a carpool. If you are going to the GUC, please type 'to guc', if you are leaving the GUC, please type 'from guc'.", nil
		} else {
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! Are you offering a ride? Or are you requesting One?")
		}
	} else {
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
	comparable := strings.ToLower(message)
	FromGUC, fromGUCFound := session["fromGUC"]

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
			session["latitude"], _ = strconv.ParseFloat(exp.FindAllString(comparable, -1)[0], 64)
			session["longitude"], _ = strconv.ParseFloat(exp.FindAllString(comparable, -1)[1], 64)
			return "You chose the location with the latitude " + strconv.FormatFloat(session["latitude"].(float64), 'f', -1, 64) + ", and the longitude " + strconv.FormatFloat(session["longitude"].(float64), 'f', -1, 64) + ". What time would you like to your ride to be?", nil
		}
		var response string
		if FromGUC.(bool) {
			response = "Where are you going?"
		} else {
			response = "Where can you pick up people?"
		}
		return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + response)
	}
	stTime, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "", fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
		}
		session["time"] = stTime
		return "You want your ride to take place around " + (session["time"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + ". How many passengers can you take with you?", nil
	}

	AvailableSeats, availableSeatsFound := session["availableSeats"]
	if !availableSeatsFound && timeFound && latitudeFound && longitudeFound && fromGUCFound {
		if !(strings.Contains(comparable, "4")) && !(strings.Contains(comparable, "3")) && !(strings.Contains(comparable, "2")) && !(strings.Contains(comparable, "1")) {
			return "you can only have 1-4 passengers, not including yourself. Please enter a valid number!", nil
		}
		exp := regexp.MustCompile(`[1-4]`)
		number0, _ := strconv.ParseInt(exp.FindAllString(comparable, -1)[0], 10, 64)
		number := int(number0)
		session["availableSeats"] = number
		return "You've chosen to take up to " + strconv.FormatInt(number0, 10) + " more passengers.", nil
	}
	C, err := DB.NewCarpool(session["gucID"].(string), session["longitude"].(float64), session["latitude"].(float64), session["name"].(string), FromGUC.(bool), AvailableSeats.(int), stTime.(time.Time).Format("Jan 2, 2006 at 3:04pm (EET)"))
	if err == nil {
		DB.InsertDB(&C)
	} else {
		return "kalam 3eeeeeb", fmt.Errorf("kalam 3eeeb awy y3ny")
	}

	return "", fmt.Errorf("Whoops! An error occured in your session. Can you please log out and log back in again?")
}

// Function to handle the specifics that the user wants in the carpool he requested.
func requestCarpoolChat(session Session, message string) (string, error) {
	// Check if user is going to or leaving the GUC.
	fromGUC, fromGUCFound := session["fromGUCreq"]
	comparable := strings.ToLower(message)
	// Set if the user is going to GUC or leaving.
	if !fromGUCFound {
		if strings.Contains(comparable, "going to") || strings.Contains(comparable, "to guc") {
			session["fromGUCreq"] = false
			return "You've chosen to find a carpool going to the GUC! Where would you like to be picked up from?", nil
		} else if strings.Contains(comparable, "leaving") || strings.Contains(comparable, "from guc") {
			session["fromGUCreq"] = true
			return "You chose to leave the campus. Where would you like to go?", nil
		} else {
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! Are you going to the GUC? Or are you leaving campus?")
		}
	}

	// Check the location the user wants.
	_, latitudeFound := session["latitudereq"]
	_, longitudeFound := session["longitudereq"]
	// Get the location the user wants to go to.
	if (!latitudeFound || !longitudeFound) && fromGUCFound {
		if strings.Contains(comparable, "latitude") && strings.Contains(comparable, "longitude") {
			exp := regexp.MustCompile(`[0-9]+[\.]?[0-9]*`)
			session["latitudereq"] = exp.FindAllString(message, -1)[0]
			session["longitudereq"] = exp.FindAllString(message, -1)[1]
			return "You chose the location with the latitude " + session["latitudereq"].(string) + ", and the longitude " + session["longitudereq"].(string) + ". What time would you like to your ride to be?", nil
		}
		var ret string
		if fromGUC.(bool) {
			ret = "Where would you like to go?"
		} else {
			ret = "Where would you like to be picked up from?"
		}
		return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + ret)
	}

	// Get the time the user wants to leave.
	createTime, createTimeFound := session["time"]
	_, timeFound := session["timereq"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "", fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
		}
		if createTimeFound {
			duration := stTime.Sub(createTime.(time.Time))
			if duration.Hours() <= 4 {
				return "", fmt.Errorf("You already have a carpool around that same time! Please choose a different time")
			}
		}
		session["timereq"] = stTime
	}

	// The user's request is complete. Set and delete the proper session variables.
	_, timeFound = session["timereq"]
	if timeFound && fromGUCFound && latitudeFound && longitudeFound {
		details := getDetails(session)
		session["requestComplete"] = true
		delete(session, "requestOrCreate")
		return "Your request is complete! Here are the details: " + details + " You can now view the most suitable carpools, view all carpools, cancel your request, edit your request or choose one of the available carpools. So, what do you want to do?", nil
	}

	return "", fmt.Errorf("Whoops! An error occured in your session. Can you please log out and log back in again?")
}

func postRequestHandler(res http.ResponseWriter, session Session, data JSON) {
	_, requestExists := session["requestComplete"]
	comparable := strings.ToLower(data["message"].(string))
	if strings.Contains(comparable, "view all") {
		allRequests, err := DB.QueryAll()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while retrieving the data from our database. Please try again in a moment.",
			})
			return
		}
		writeJSON(res, JSON{
			"message": "Here are all the available carpools!",
			"data":    allRequests,
		})
		return
	} else if strings.Contains(comparable, "edit") && strings.Contains(comparable, "request") && requestExists {
		delete(session, "fromGUCreq")
		delete(session, "latitudereq")
		delete(session, "longitudereq")
		delete(session, "timereq")
		delete(session, "requestComplete")
		session["requestOrCreate"] = "request"
		writeJSON(res, JSON{
			"message": "You chose to edit your carpool request. Let's do this piece by piece. Firstly, are you going to the GUC, or are you leaving campus?",
		})
	} else if strings.Contains(comparable, "cancel") && strings.Contains(comparable, "request") {
		delete(session, "fromGUCreq")
		delete(session, "latitudereq")
		delete(session, "longitudereq")
		delete(session, "timereq")
		delete(session, "requestComplete")
		delete(session, "requestOrCreate")
		previousChoice, myChoiceExists := session["myChoice"]
		if myChoiceExists {
			delete(session, "myChoice")
			gucID := session["gucID"].(string)
			carpoolRequests, err := DB.GetPostByID(previousChoice.(uint64))
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error while retrieving the data from our database. Please try again in a moment.",
				})
				return
			}
			wasCurrent := false
			carpoolRequest := carpoolRequests[0]
			possiblePassengers := carpoolRequest.PossiblePassengers
			currentPassengers := carpoolRequest.CurrentPassengers
			for idx, val := range possiblePassengers {
				if strings.EqualFold(val, gucID) {
					possiblePassengers = append(possiblePassengers[:idx], possiblePassengers[idx+1:]...)
					wasCurrent = true
					break
				}
			}
			for idx, val := range currentPassengers {
				if strings.EqualFold(val, gucID) {
					possiblePassengers = append(currentPassengers[:idx], currentPassengers[idx+1:]...)
					break
				}
			}
			availableSeats := carpoolRequest.AvailableSeats
			if wasCurrent {
				availableSeats--
			}
			err = DB.UpdateDB(previousChoice.(uint64), carpoolRequest.Longitude, carpoolRequest.Latitude, carpoolRequest.FromGUC, availableSeats, currentPassengers, possiblePassengers)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error removing you from the carpool. Try again later.",
				})
				return
			}
		}
		writeJSON(res, JSON{
			"message": "Your carpool request has been cancelled successfully. You can now start over. Do you want to request a carpool, or are you offering one?",
		})
		return
	} else if strings.Contains(comparable, "choose") {
		_, myChoiceExists := session["myChoice"]
		if myChoiceExists {
			res.WriteHeader(http.StatusForbidden)
			writeJSON(res, JSON{
				"message": "You already chose a carpool. Please cancel before choosing a new one.",
			})
			return
		}
		exp := regexp.MustCompile(`[0-9]+`)
		postID := exp.FindString(comparable)
		postIDint, err := strconv.ParseUint(postID, 10, 64)
		if err != nil {
			res.WriteHeader(http.StatusUnprocessableEntity)
			writeJSON(res, JSON{
				"message": "There was an error when converting the postID from string to int. Please try again.",
			})
			return
		}
		carpoolRequests, err := DB.GetPostByID(postIDint)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while retrieving the data from our database. Please try again in a moment.",
			})
			return
		}
		carpoolRequest := carpoolRequests[0]
		if strings.EqualFold(carpoolRequest.GUCID, session["gucID"].(string)) {
			res.WriteHeader(http.StatusForbidden)
			writeJSON(res, JSON{
				"message": "You can't join your own carpool!",
			})
			return
		}
		possiblePassengers := carpoolRequest.PossiblePassengers
		possiblePassengers = append(possiblePassengers, session["gucID"].(string))
		err = DB.UpdateDB(postIDint, carpoolRequest.Longitude, carpoolRequest.Latitude, carpoolRequest.FromGUC, carpoolRequest.AvailableSeats, carpoolRequest.CurrentPassengers, possiblePassengers)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error updating. Try again later.",
			})
			return
		}
		session["myChoice"] = postIDint
		writeJSON(res, JSON{
			"message": "You've successfully chosen a carpool! Now you have to wait for the original poster to accept your request.",
		})
		return
	}
	res.WriteHeader(http.StatusUnprocessableEntity)
	writeJSON(res, JSON{
		"message": "I did not understand what you said. Would you like to view all the available carpools,  cancel your request, edit your request or choose an available carpool?",
	})
	return
}

// Function to write out a JSON response.
func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}

// Function that gets the details of a request session.
func getDetails(session Session) string {
	str := ""

	if session["fromGUCreq"].(bool) {
		str += "You're leaving the GUC, and going to the location with "
	} else {
		str += "You're coming to the GUC, from the location with "
	}

	str += "latitude " + session["latitudereq"].(string) + " and longitude " + session["longitudereq"].(string) + "."

	str += "You want your ride to take place around " + (session["timereq"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + "."

	return str
}
