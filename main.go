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
	JSON    map[string]interface{}
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
		"message": "Please use the route '/welcome' to log in.",
	})
}

//
func startSession(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to GET requests on this route.",
		})
		return
	}

	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))

	sessions[uuid] = Session{}
	writeJSON(res, JSON{
		"uuid":    uuid,
		"message": "Welocme to GUC Carpool! Please log in using your GUC-ID and name",
	})

}

func handleChat(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you didn't send any proper data with that " + req.Method + " request. I can only listen to POST requests on this route.",
		})
		return
	}

	uuid := req.Header.Get("Authorization")
	if uuid == "" {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON{
			"message": "I'm sorry, but you don't seem to be logged in. Please log in and try again.",
		})
		return
	}

	session, sessionFound := sessions[uuid]
	if !sessionFound {
		res.WriteHeader(http.StatusUnauthorized)
		writeJSON(res, JSON{
			"message": "I'm sorry, but your session has expired. Please log in and try again.",
		})
		return
	}

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

	_, loggedIn := session["gucID"]
	if !loggedIn {
		_, gucIDFound := data["gucID"]
		_, nameFound := data["name"]
		if !gucIDFound || !nameFound {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again.",
			})
			return
		} else {
			gucID := data["gucID"].(string)
			name := data["name"].(string)
			session["gucID"] = gucID
			session["name"] = name
			writeJSON(res, JSON{
				"message": "Hello " + name + ". Are you offering a ride, or are you requesting one?",
			})
			return
		}
	}

	messageRecieved, received := data["message"]
	if !received {
		res.WriteHeader(http.StatusBadRequest)
		writeJSON(res, JSON{
			"message": "I did not receive a message. Are you sure you sent me something?",
		})
		return
	}

	comparable := strings.ToLower(messageRecieved.(string))
	if strings.Contains(comparable, "edit") || strings.Contains(comparable, "cancel") || strings.Contains(comparable, "view") || strings.Contains(comparable, "delete") || strings.Contains(comparable, "accept") || strings.Contains(comparable, "reject") {
		postCreateHandler(res, session, data)
		return
	}

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
		if strings.Contains(comparable, "create") || (strings.Contains(comparable, "offer") && !strings.Contains(comparable, "offered")) {
			session["requestOrCreate"] = "create"
			return "You've chosen to create a carpool. Are you going to the GUC, or are you leaving campus?", nil
		} else if strings.Contains(comparable, "request") || strings.Contains(comparable, "find") || strings.Contains(comparable, "join") {
			session["requestOrCreate"] = "request"
			return "You've chosen to request a carpool. Are you going to the GUC, or are you leaving campus?", nil
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
	// make the message in lowercase to pass on

	comparable := strings.ToLower(message)
	FromGUC, fromGUCFound := session["fromGUC"]

	//check if he's going to the guc or leaving the guc
	if !fromGUCFound {
		if strings.Contains(comparable, "to guc") || strings.Contains(comparable, "to the guc") || strings.Contains(comparable, "going") {
			session["fromGUC"] = false
			return "You've chosen to create a carpool that's going to the GUC. Where can you pick up people? please enter your latitude and longitude", nil
		} else if strings.Contains(comparable, "from guc") || strings.Contains(comparable, "from the guc") || strings.Contains(comparable, "leaving") {
			session["fromGUC"] = true
			return "You've chosen to create a carpool that's leaving the GUC. Where are you going? please enter your latitude and longitude", nil
		} else {
			return "I'm sorry you didn't answer my question. Are you going to the GUC or leaving the GUC?", nil
		}
	}

	//take his latitude and longitude
	_, latitudeFound := session["latitude"]
	_, longitudeFound := session["longitude"]
	if (!latitudeFound || !longitudeFound) && fromGUCFound {
		if strings.Contains(comparable, "latitude") && strings.Contains(comparable, "longitude") {
			exp := regexp.MustCompile(`[0-9]+[\.]?[0-9]*`)
			session["latitude"], _ = strconv.ParseFloat(exp.FindAllString(comparable, -1)[0], 64)
			session["longitude"], _ = strconv.ParseFloat(exp.FindAllString(comparable, -1)[1], 64)
			return "You chose the location with the latitude " + strconv.FormatFloat(session["latitude"].(float64), 'f', -1, 64) + ", and the longitude " + strconv.FormatFloat(session["longitude"].(float64), 'f', -1, 64) + ". What time would you like to your ride to be?", nil
		} else {
			var response string
			if FromGUC.(bool) {
				response = "Where are you going?"
			} else {
				response = "Where can you pick up people?"
			}
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + response)
		}
	}

	//take his start time
	stTime, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "", fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
		} else {
			session["time"] = stTime
			return "You want your ride to take place around " + (session["time"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + ". How many passengers can you take with you?", nil
		}
	}
	//take how many available seats
	_, availableSeatsFound := session["availableSeats"]
	if !availableSeatsFound && timeFound && latitudeFound && longitudeFound && fromGUCFound {
		if !(strings.Contains(comparable, "4")) && !(strings.Contains(comparable, "3")) && !(strings.Contains(comparable, "2")) && !(strings.Contains(comparable, "1")) {
			return "you can only have 1 to 4 passengers, not including yourself. Please enter a valid number!", nil
		}
		exp := regexp.MustCompile(`[1-4]`)
		number0, _ := strconv.ParseInt(exp.FindAllString(comparable, -1)[0], 10, 64)
		number := int(number0)
		session["availableSeats"] = number

		//make a new carpool
		_, postFound := session["postID"]
		if !postFound {
			C, err := DB.NewCarpool(session["gucID"].(string), session["longitude"].(float64), session["latitude"].(float64), session["name"].(string), FromGUC.(bool), session["availableSeats"].(int), stTime.(time.Time).Format("Jan 2, 2006 at 3:04pm (EET)"))
			if err != nil {
				return "", fmt.Errorf("whoops we couldn't make you a new carpool")
			}
			//insert that new carpool into the database
			DB.InsertDB(&C)
			session["createComplete"] = false
			session["postID"] = C.PostID
			session["currentPassengers"] = C.CurrentPassengers
			session["possiblePassengers"] = C.PossiblePassengers
		} else {
			k := DB.UpdateDB(session["postID"].(uint64), session["longitude"].(float64), session["latitude"].(float64), session["fromGUC"].(bool), session["availableSeats"].(int), session["currentPassengers"].([]string), session["possiblePassengers"].([]string), stTime.(time.Time).Format("Jan 2, 2006 at 3:04pm (EET)"))
			if k != nil {
				return "", fmt.Errorf("We couldn't update your carpool, please try again")
			}

		}

		return "You've chosen to take up to " + strconv.FormatInt(number0, 10) + " more passengers. Now you can Edit your Carpool, Delete your Carpool, Accept or Reject a request, or View yur Carpool. Choose wisely my friend.", nil
	}

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
			session["latitude"] = exp.FindAllString(message, -1)[0]
			session["longitude"] = exp.FindAllString(message, -1)[1]
			return "You chose the location with the latitude " + session["latitude"].(string) + ", and the longitude " + session["longitude"].(string) + ". What time would you like to your ride to be?", nil
		} else {
			var ret string
			if fromGUC.(bool) {
				ret = "Where would you like to go?"
			} else {
				ret = "Where would you like to be picked up from?"
			}
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + ret)
		}
	}

	_, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", message)
		if err != nil {
			return "", fmt.Errorf("An error occured when parsing the time. Can you please tell me again when you want your ride to be?")
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

func postCreateHandler(res http.ResponseWriter, session Session, data JSON) {
	_, createFound := session["createComplete"]
	postID := session["postID"]
	comparable := strings.ToLower(data["message"].(string))
	if strings.Contains(comparable, "delete") && createFound {
		err := DB.DeleteDB(postID.(uint64))
		if err != nil {
			return
		}
		delete(session, "fromGUC")
		delete(session, "latitude")
		delete(session, "longitude")
		delete(session, "time")
		delete(session, "availableSeats")
		delete(session, "createComplete")
		writeJSON(res, JSON{
			"message": "You chose to delete your carpool. Now you can create or request a carpool."})
		return

	} else if strings.Contains(comparable, "edit") {
		delete(session, "fromGUC")
		delete(session, "latitude")
		delete(session, "longitude")
		delete(session, "time")
		delete(session, "availableSeats")
		delete(session, "createComplete")
		writeJSON(res, JSON{
			"message": "You chose to edit your carpool. Let's do this piece by piece. Firstly, are you going to the GUC, or are you leaving campus?	"})
		return
	} else if strings.Contains(comparable, "Accept") && createFound {

	} else if strings.Contains(comparable, "reject") && createFound {

	} else if strings.Contains(comparable, "view") && createFound {
		myRequest, err := DB.GetPostByID(postID.(uint64))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while retrieving the data from our database. Please try again in a moment.",
			})
			return
		}
		writeJSON(res, JSON{
			"message": "Here is your carpool details!",
			"data":    myRequest,
		})
	} /*else if strings.Contains(comparable, "cancel") && createFound{
		session["choice"] = postID
		Choice, ChoiceFound := session["choice"]
		if ChoiceFound {
			delete(session, "choice")
			name := session["name"].(string)
			carpools, err := DB.GetPostByID(Choice.(uint64))
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error while retrieving the data from our database. Please try again in a moment.",
				})
				return
			}
			carpool := carpools[0]
			possiblePassengers := carpool.PossiblePassengers
			currentPassengers := carpool.CurrentPassengers
			for idx, val := range currentPassengers {
				if strings.EqualFold(val, name) {
					possiblePassengers = append(possiblePassengers[:idx], possiblePassengers[idx+1:]...)
					break
				}
			}
			for idx, val := range currentPassengers {
				if strings.EqualFold(val, name) {
					possiblePassengers = append(currentPassengers[:idx], currentPassengers[idx+1:]...)
					break
				}
			}
			err = DB.UpdateDB(Choice.(uint64), carpool.Longitude, carpool.Latitude, carpool.FromGUC, carpool.AvailableSeats, currentPassengers, possiblePassengers)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error removing you from the carpool. Try again later.",
				})
				return
			}
		}
		writeJSON(res, JSON{
			"message": "Your carpool pickup has been cancelled successfully. You can now pick uo other people or delete your Carpool.",
		})
		return
	}*/

	res.WriteHeader(http.StatusUnprocessableEntity)
	writeJSON(res, JSON{
		"message": "I did not understand what you said. Would you like to view your carpool, delete carpool, edit your request, accept or reject a request, or cancel an acceptedd request. So, what do you want to do?",
	})
	return
}

func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(data)
}
