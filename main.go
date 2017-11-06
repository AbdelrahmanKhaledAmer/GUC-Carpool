package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"

	"github.com/AbdelrahmanKhaledAmer/GUC-Carpool/DB"
	"github.com/AbdelrahmanKhaledAmer/GUC-Carpool/DirectionsAPI"
	cors "github.com/heppu/simple-cors"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("GUC-Carpool server listening on port " + port)

	now.TimeFormats = append(now.TimeFormats, "02 Jan 2006 15:04", "5/11/2017 8.30", "nov 5,2017 at 8.30")

	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", serveAndLog(startSession))
	mux.HandleFunc("/chat", serveAndLog(handleChat))
	mux.HandleFunc("/", serveAndLog(serve))

	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, cors.CORS(mux)))

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
		"message": "Welcome to GUC Carpool! Please tell me your GUC-ID and name separated by ':'.",
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
			"message": "I'm sorry, but it seems that I forgot who yo are in. Please log in and try again.",
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
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again.(ex. '12-3456:MyName') It is so easy you just write them :D",
			})
			return
		}
		gucID := login[0]
		name := login[1]
		// Check if gucID and name are empty
		if gucID == "" || name == "" {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Something went wrong. You have to give me both your name and your GUC-ID in order to successfully start your session. Please try again. I can not infer this Info but when I grow up I may. (ex. '12-3456:MyName')",
			})
			return
		}
		// Check if gucId is in a valid format (eg. 13-2456)
		exp := regexp.MustCompile(`[0-9]+-[0-9]+`)
		match := exp.MatchString(gucID)
		if !match {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "Your GUC ID is invalid. Are you sure you entered it correctly? type it correctly or I will keep anoying you with this message. (ex. '12-3456')",
			})
			return
		}
		oldSession := false
		// Find if an old session has the same user. If found, migrate the information from the old session to the new one, and delete the old one.
		for key1, val1 := range sessions {
			currentGucID, currentIDFound := val1["gucID"]
			if currentIDFound && strings.EqualFold(currentGucID.(string), gucID) && uuid != key1 {
				for key2, val2 := range val1 {
					session[key2] = val2
				}
				delete(sessions, key1)
				oldSession = true
				break
			}
		}
		session["gucID"] = gucID
		session["name"] = name
		// If no old session is found, check if user has a previous carpool
		if !oldSession {
			carpoolRequests, err := DB.QueryAll()
			passengerRequests, err2 := DB.QueryAllPassengerRequests()
			if err != nil || err2 != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error in getting your data from the database. Error: " + err.Error(),
				})
				return
			}
			for i := 0; i < len(carpoolRequests); i++ {
				currentCarpool := carpoolRequests[i]
				if strings.EqualFold(currentCarpool.GUCID, gucID) {
					session["fromGUC"] = currentCarpool.FromGUC
					session["latitude"] = currentCarpool.Latitude
					session["longitude"] = currentCarpool.Longitude
					session["time"] = currentCarpool.Time
					session["availableSeats"] = currentCarpool.AvailableSeats
					session["createComplete"] = true
					session["postID"] = currentCarpool.PostID
				}
			}
			for i := 0; i < len(passengerRequests); i++ {
				currentPassenger := passengerRequests[i]
				if strings.EqualFold(currentPassenger.Passenger.GUCID, gucID) && currentPassenger.Notify != 0 && currentPassenger.Notify != 3 {
					session["myChoice"] = currentPassenger.PostID
				}
			}
		}

		responseMessage := "Hello " + name + ". You can view all available carpools by typing 'view all', or 'view carpool' to view one you already have, cancel your request by typing 'cancel request', edit your request by typing 'edit request' or choose an available carpool by typing 'choose ID' where ID is the postID of the carpool of your choice. You can also choose to offer other people a ride by creating a carpool by typing 'create', or specify the details of a carpool you wish to request by typing 'request'."
		// responseNotify, err := getNotifications(session)
		// if err != nil {
		// 	responseNotify = "can not retrive your notifications right now  "
		// }
		writeJSON(res, JSON{
			"message": responseMessage,
		})
		return
	}

	// See if the user wishes to interact with data from the database or edit his session.
	comparable := strings.ToLower(messageRecieved.(string))
	//_, carpoolRequestFound := session["postID"]
	//_, passengerRequestFound := session["myChoice"]
	if /*(carpoolRequestFound || passengerRequestFound) &&*/ strings.Contains(comparable, "notifications") || strings.Contains(comparable, "notify") {
		notifications, err := getNotifications(session)
		if err != nil {
			writeJSON(res, JSON{
				"message": "I could not retrieve your notifications at the moment, please try again later.",
			})
			return
		}
		writeJSON(res, JSON{
			"message": notifications,
		})
		return
	}

	if strings.Contains(comparable, "edit") || strings.Contains(comparable, "cancel") || strings.Contains(comparable, "choose") || (strings.Contains(comparable, "view") && (strings.Contains(comparable, "all") || strings.Contains(comparable, "carpool"))) || strings.Contains(comparable, "delete") || strings.Contains(comparable, "reject") || strings.Contains(comparable, "accept") || strings.Contains(comparable, "directions") {
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

// This function is used when the user is trying to create or request a carpool. It checks the proper variables, and calls one of two helper functions
func processMessage(session Session, message string) (string, error) {
	requestOrCreate, requestOrCreateFound := session["requestOrCreate"]
	comparable := strings.ToLower(message)
	if !requestOrCreateFound {
		if strings.Contains(comparable, "create") || (strings.Contains(comparable, "offer")) {
			_, postFound := session["postID"]
			if postFound {
				return "You already have a carpool! I have a very good memory, I can remember things you know :P. ", nil
			}
			session["requestOrCreate"] = "create"
			return "You've chosen to create a carpool. Are you going to the GUC, or are you leaving the GUC?", nil
		} else if strings.Contains(comparable, "request") || strings.Contains(comparable, "find") || strings.Contains(comparable, "join") {
			session["requestOrCreate"] = "request"
			return "You've chosen to request a carpool. Are you going to the GUC, or are you leaving the GUC?", nil
		} else {
			return "", fmt.Errorf("I'm sorry, but you didn't answer my question! Are you offering a ride? Or are you requesting One? I am not busy I can do this all day. type 'create' to make a carpool or 'request' to make a request")
		}
	} else {
		if requestOrCreate == "create" {
			return createCarpoolChat(session, message)
		} else if requestOrCreate == "request" {
			return requestCarpoolChat(session, message)
		} else {
			return "", fmt.Errorf("I don't seem to understand what your message '" + comparable + "' means. Can you please clarify?")
		}
	}
}

func createCarpoolChat(session Session, message string) (string, error) {
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
			return "I'm sorry you didn't answer my question. Are you going to the GUC or leaving the GUC? (ex. if you're leaving you can type 'from guc' or if you're going to campus you can type 'to guc'.", nil
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
		}
		var response string
		if FromGUC.(bool) {
			response = "Where are you going?"
		} else {
			response = "Where can you pick up people?"
		}
		return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + response + " Please enter the latitude and longitude like this 'latitude 30.08 longitude 31.33'")
	}

	//take his start time
	requestTime, requestTimeFound := session["timereq"]
	stTime, timeFound := session["time"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := now.Parse(message)
		if err != nil {
			return "", fmt.Errorf("This is not a valid time format. Can you please tell me again when you want your ride to be? One possible format you can use is 'yyyy-mm-dd hh:mm'")
		}
		if requestTimeFound {
			duration := stTime.Sub(requestTime.(time.Time))
			if math.Abs(duration.Hours()) <= 4 {

				return "", fmt.Errorf("You already have a carpool around that same time! Did you forget? Please choose a different time")
			}
		}
		now := time.Now()
		valid := stTime.After(now)
		if !valid {
			return "", fmt.Errorf("This time doesn't make sense! You need to choose a time in the future. I am not that dumb you know")
		}
		session["time"] = stTime
		return "You want your ride to take place around " + (session["time"].(time.Time)).Format("Jan 2, 2006 at 3:04pm (EET)") + ". How many passengers can you take with you?", nil
	}

	//take how many available seats
	_, availableSeatsFound := session["availableSeats"]
	if !availableSeatsFound && timeFound && latitudeFound && longitudeFound && fromGUCFound {
		if !(strings.Contains(comparable, "4")) && !(strings.Contains(comparable, "3")) && !(strings.Contains(comparable, "2")) && !(strings.Contains(comparable, "1")) {
			return "you can only have 1-4 passengers, not including yourself. Please enter a valid number!", nil
		}
		exp := regexp.MustCompile(`[1-4]`)
		number0, _ := strconv.ParseInt(exp.FindAllString(comparable, -1)[0], 10, 64)
		number := int(number0)
		session["availableSeats"] = number

		_, postFound := session["postID"]
		if !postFound {
			C, err := DB.NewCarpool(session["gucID"].(string), session["longitude"].(float64), session["latitude"].(float64), session["name"].(string), FromGUC.(bool), session["availableSeats"].(int), stTime.(time.Time).Format("Jan 2, 2006 at 3:04pm (EET)"))
			if err != nil {
				return "", fmt.Errorf("An error occured when creating your carpool. Please try again later")
			}
			//insert that new carpool into the database
			err = DB.InsertDB(&C)
			if err != nil {
				return "", fmt.Errorf("An error occured while inserting into the database. Error: " + err.Error())
			}
			session["createComplete"] = true
			session["postID"] = C.PostID
			session["currentPassengers"] = C.CurrentPassengers
			session["possiblePassengers"] = C.PossiblePassengers
		} else {
			err := DB.UpdateDB(session["postID"].(uint64), session["longitude"].(float64), session["latitude"].(float64), session["fromGUC"].(bool), session["availableSeats"].(int), session["currentPassengers"].([]string), session["possiblePassengers"].([]string), stTime.(time.Time))
			if err != nil {
				return "", fmt.Errorf("An error occured when creating your carpool. Please try again later Error: " + err.Error())
			}
		}
		delete(session, "requestOrCreate")
		return "You've chosen to take up to " + strconv.FormatInt(number0, 10) + " more passengers. Your carpool is now complete! You can see it by typing 'view carpool'.", nil
	}

	return "I did not understand what you said. I am only a computer after all.", nil
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
			return "You've chosen to find a carpool going to the GUC! Where would you like to be picked up from? Please eneter a latitude and longitude.", nil
		} else if strings.Contains(comparable, "leaving") || strings.Contains(comparable, "from guc") {
			session["fromGUCreq"] = true
			return "You chose to leave the campus. Where would you like to go? Please enter a latitude and longitude.", nil
		} else {
			return "I'm sorry you didn't answer my question. Are you going to the GUC or leaving the GUC? (ex. if you're leaving you can type 'from guc' or if you're going to campus you can type 'to guc'.", nil
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
		return "", fmt.Errorf("I'm sorry, but you didn't answer my question! " + ret + " Please enter the latitude and longitude like this 'latitude 30.08 longitude 31.33'")
	}

	// Get the time the user wants to leave.
	createTime, createTimeFound := session["time"]
	_, timeFound := session["timereq"]
	if !timeFound && fromGUCFound && latitudeFound && longitudeFound {
		stTime, err := now.Parse(message)
		if err != nil {
			return "", fmt.Errorf("This is not a valid time format. Can you please tell me again when you want your ride to be? One possible format you can use is 'yyyy-mm-dd hh:mm'")
		}
		if createTimeFound {
			duration := stTime.Sub(createTime.(time.Time))
			if math.Abs(duration.Hours()) <= 4 {
				return "", fmt.Errorf("You already have a carpool around that same time! You can't be in two places at once! Please choose a different time")
			}
		}
		now := time.Now()
		valid := stTime.After(now)
		if !valid {
			return "", fmt.Errorf("This time doesn't make sense! You need to choose a time in the past! I do not have a time machine")
		}
		session["timereq"] = stTime
	}

	// The user's request is complete. Set and delete the proper session variables.
	_, timeFound = session["timereq"]
	if timeFound && fromGUCFound && latitudeFound && longitudeFound {
		details := getDetails(session)
		session["requestComplete"] = true
		delete(session, "requestOrCreate")
		return "Your request is complete! Here are the details: " + details + " You can now view all carpools, cancel your request, edit your request or choose one of the available carpools. So, what do you want to do?", nil
	}

	return "Looks like you have a carpool already. You can edit it if you want.", nil
}

func postRequestHandler(res http.ResponseWriter, session Session, data JSON) {
	_, requestExists := session["requestComplete"]
	_, createFound := session["createComplete"]
	postID, postIDExists := session["postID"]
	comparable := strings.ToLower(data["message"].(string))
	if strings.Contains(comparable, "view all") {
		allRequests, err := DB.QueryAll()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while retrieving the data from our database. Error: " + err.Error(),
			})
			return
		}
		cpString := ""
		for i := 1; i < len(allRequests); i++ {
			cpString += allRequests[i].CarpoolToString() + ",</br>"
		}
		if len(allRequests) < 2 {
			writeJSON(res, JSON{
				"message": "Oops no current carpool offers available looks like you'll be walking. :P",
			})
			return
		}
		writeJSON(res, JSON{
			"message": "Here are all the available carpools!</br>" + cpString,
		})
		return
	} else if strings.Contains(comparable, "edit") && strings.Contains(comparable, "request") {
		delete(session, "fromGUCreq")
		delete(session, "latitudereq")
		delete(session, "longitudereq")
		delete(session, "timereq")
		if !requestExists {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "You can't edit a request if you don't have one. I am starting to doubt your inteligence.",
			})
			return
		}
		delete(session, "requestComplete")
		session["requestOrCreate"] = "request"
		writeJSON(res, JSON{
			"message": "You chose to edit your carpool request. Let's do this piece by piece. Firstly, are you going to the GUC, or are you leaving campus?",
		})
		return
	} else if strings.Contains(comparable, "cancel") && strings.Contains(comparable, "request") {
		delete(session, "fromGUCreq")
		delete(session, "latitudereq")
		delete(session, "longitudereq")
		delete(session, "timereq")
		delete(session, "requestComplete")
		delete(session, "requestOrCreate")
		previousChoice, myChoiceExists := session["myChoice"]

		if myChoiceExists {
			gucID := session["gucID"].(string)
			carpoolRequests, err := DB.GetPostByID(previousChoice.(uint64))
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error while retrieving the data from our database. Error: " + err.Error(),
				})
				return
			}
			delete(session, "myChoice")

			wasCurrent := false
			if len(carpoolRequests) == 0 {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "No posts exist with this ID. It must have been deleted!",
				})
				return
			}

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
					currentPassengers = append(currentPassengers[:idx], currentPassengers[idx+1:]...)
					break
				}
			}
			availableSeats := carpoolRequest.AvailableSeats
			if wasCurrent {
				availableSeats++
			}
			err = DB.UpdateDB(previousChoice.(uint64), carpoolRequest.Longitude, carpoolRequest.Latitude, carpoolRequest.FromGUC, availableSeats, currentPassengers, possiblePassengers, carpoolRequest.StartTime)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error removing you from the carpool. Error: " + err.Error(),
				})
				return
			}
			passengerRequests, err := DB.GetPassengerRequestsByGUCID(gucID)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "There was an error while retrieving the data from our database. Error: " + err.Error(),
				})
				return
			}
			if len(passengerRequests) == 0 {
				res.WriteHeader(http.StatusUnauthorized)
				writeJSON(res, JSON{
					"message": "I can't find your carpool request. Please try again in a moment",
				})
				return
			}
			for index := 0; index < len(passengerRequests); index++ {

				passengerRequest := passengerRequests[index]
				if passengerRequest.PostID == previousChoice.(uint64) {
					err = DB.UpdatePassengerRequest(passengerRequest.Passenger.GUCID, passengerRequest.Passenger.Name, passengerRequest.PostID, 3)
					if err != nil {
						res.WriteHeader(http.StatusInternalServerError)
						writeJSON(res, JSON{
							"message": "I can't update your information in our database. Error: " + err.Error(),
						})
						return
					}
				}
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
				"message": "You already chose a carpool. Please cancel before choosing a new one. You can only be in one carpool at a time.",
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
				"message": "There was an error while retrieving the data from our database. Error: " + err.Error(),
			})
			return
		}
		if len(carpoolRequests) < 1 {
			writeJSON(res, JSON{
				"message": "not a valid post ID. I will say it was a typo! Try again like this 'choose ID' but replace ID with the post number!",
			})
			return
		}
		myDetails, err := DB.NewPassengerRequest(session["gucID"].(string), session["name"].(string), postIDint, 1)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while creating your information. Error: " + err.Error(),
			})
			return
		}

		carpoolRequest := carpoolRequests[0]
		if strings.EqualFold(carpoolRequest.GUCID, session["gucID"].(string)) {
			res.WriteHeader(http.StatusForbidden)
			writeJSON(res, JSON{
				"message": "You can't join your own carpool! I mean ... why would you even do that?",
			})
			return
		}
		//insert after check
		err = DB.InsertPassengerRequest(&myDetails)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error while saving your information. Error: " + err.Error(),
			})
			return
		}

		possiblePassengers := carpoolRequest.PossiblePassengers
		possiblePassengers = append(possiblePassengers, session["gucID"].(string))
		err = DB.UpdateDB(postIDint, carpoolRequest.Longitude, carpoolRequest.Latitude, carpoolRequest.FromGUC, carpoolRequest.AvailableSeats, carpoolRequest.CurrentPassengers, possiblePassengers, carpoolRequest.StartTime)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error updating. Error: " + err.Error(),
			})
			return
		}
		session["myChoice"] = postIDint
		writeJSON(res, JSON{
			"message": "You've successfully chosen a carpool! Now you have to wait for the original poster to accept your request. He may cancel the carpool, or choose to accept others, so plesae stay alert to your notifications. You can view them by typing 'notify' or 'notifications'.",
		})
		return
	} else if strings.Contains(comparable, "delete") && strings.Contains(comparable, "carpool") {
		delete(session, "fromGUC")
		delete(session, "latitude")
		delete(session, "longitude")
		delete(session, "time")
		delete(session, "availableSeats")
		delete(session, "createComplete")
		delete(session, "requestOrCreate")
		delete(session, "postID")
		if !createFound || !postIDExists {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "You cannot delete a carpool if you've never created one. I can pretend that I deleted one that doesn't exist though!",
			})
			return
		}
		err := DB.DeleteDB(postID.(uint64))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "There was an error deleting the carpool from our database. Error: " + err.Error(),
			})
			return
		}
		writeJSON(res, JSON{
			"message": "You chose to delete your carpool. Now you can create a new one in its place if you wish. You can also request ne or view all the available ones.",
		})
		return
	} else if strings.Contains(comparable, "edit") && strings.Contains(comparable, "carpool") {
		delete(session, "postID")
		delete(session, "fromGUC")
		delete(session, "latitude")
		delete(session, "longitude")
		delete(session, "time")
		delete(session, "availableSeats")
		delete(session, "createComplete")
		session["requestOrCreate"] = "create"
		writeJSON(res, JSON{
			"message": "You chose to edit your carpool. Let's do this piece by piece. Firstly, are you going to the GUC, or are you leaving campus?	",
		})
		return
	} else if strings.Contains(comparable, "view") && strings.Contains(comparable, "carpool") {
		if !postIDExists {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "You can't view your carpool because you didn't make one yet. Go make one if you really want to do that. Just type something like 'create'.",
			})
			return
		}
		myRequest, err := DB.GetPostByID(postID.(uint64))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "Could not get the carpool request. Error: " + err.Error(),
			})
			return
		}
		if len(myRequest) == 0 {
			res.WriteHeader(http.StatusNotFound)
			writeJSON(res, JSON{
				"message": "There seem to be no requests in our databse matching yours.",
			})
			return
		}
		writeJSON(res, JSON{
			"message": "Here are your carpool details!</br>" + myRequest[0].CarpoolToString(),
		})
		return
	} else if strings.Contains(comparable, "reject") {
		if !postIDExists || !createFound {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "You can't reject passengers from your carpool because you didn't make one yet. What are you even rejecting them for?",
			})
			return
		}
		exp := regexp.MustCompile("[0-9]+-[0-9]+")
		passengerID := exp.FindString(comparable)
		err := DB.RejectPassenger(passengerID, postID.(uint64))
		if err != nil {
			res.WriteHeader(http.StatusUnprocessableEntity)
			writeJSON(res, JSON{
				"message": "There was an error in rejecting this passenger. Error: " + err.Error(),
			})
			return
		}
		writeJSON(res, JSON{
			"message": "You successfully rejected the passenger with ID " + passengerID + ". What else would you like to do?",
		})
		return
	} else if strings.Contains(comparable, "accept") {
		if !createFound || !postIDExists {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "You cannot accept a passenger to your carpool if you haven't made one. That's just weird!",
			})
			return
		}
		exp := regexp.MustCompile("[0-9]+-[0-9]+")
		passengerID := exp.FindString(comparable)
		err := DB.AcceptPassenger(passengerID, postID.(uint64))
		if err != nil {
			res.WriteHeader(http.StatusUnprocessableEntity)
			writeJSON(res, JSON{
				"message": "There was an error in accepting this passenger. Error: " + err.Error(),
			})
			return
		}
		writeJSON(res, JSON{
			"message": "You successfully accepted the passenger with ID " + passengerID + ". What else would you like to do?",
		})
		return
	} else if strings.Contains(comparable, "directions") {
		if !createFound || !postIDExists {
			res.WriteHeader(http.StatusUnauthorized)
			writeJSON(res, JSON{
				"message": "I can't give you the directions because you don't have a carpool created.",
			})
			return
		}
		if session["fromGUC"].(bool) {
			directions, err := DirectionsAPI.GetRoute("German University IN cairo", strconv.FormatFloat(session["latitude"].(float64), 'f', -1, 64)+","+strconv.FormatFloat(session["longitude"].(float64), 'f', -1, 64))
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				writeJSON(res, JSON{
					"message": "An error occured while recieving the directions. Error: " + err.Error(),
				})
				return
			}
			writeJSON(res, JSON{
				"message": directions,
			})
			return
		}
		directions, err := DirectionsAPI.GetRoute(strconv.FormatFloat(session["latitude"].(float64), 'f', -1, 64)+","+strconv.FormatFloat(session["longitude"].(float64), 'f', -1, 64), "German University IN cairo")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			writeJSON(res, JSON{
				"message": "An error occured while recieving the directions. the location you entered can not be found",
			})
			return
		}
		writeJSON(res, JSON{
			"message": directions,
		})
		return
	}
	res.WriteHeader(http.StatusUnprocessableEntity)
	writeJSON(res, JSON{
		"message": "I did not understand what you said. Would you like to view all the available carpools,  cancel your request, edit your request or choose an available carpool?",
	})
	return
}

// Function to get notification from databse.
func getNotifications(session Session) (string, error) {
	//should be called only whesession Sessionn gucid is set
	notificationString := ""
	passengerRequests, err := DB.GetPassengerRequestsByGUCID(session["gucID"].(string))
	if err != nil {

		return "", fmt.Errorf("error")
	}
	for index := 0; index < len(passengerRequests); index++ {
		passengerRequest := passengerRequests[index]
		if passengerRequest.Notify == 0 { //Rejected
			notificationString += "-I'm sorry, but your last carpool request couldn't be made.You can joining another one.-"
			// remove from session with this guc mail and DB

			err = DB.DeletePassengerRequest(passengerRequest.PostID, session["gucID"].(string))
			if err != nil {
				return "", fmt.Errorf("error")
			}
			delete(session, "myChoice")

		} else if passengerRequest.Notify == 2 { //Accepted
			notificationString += "-Your request has been accepted! have fun-"
		}
	}
	postID, postIDExists := session["postID"]
	if postIDExists == true {
		passengerRequests, err = DB.GetPassengerRequestsByPostID(postID.(uint64))
		if err != nil {

			return "", fmt.Errorf("error")
		}
		for i := 0; i < len(passengerRequests); i++ {
			currentPassenger := passengerRequests[i]
			if currentPassenger.Notify == 3 {
				notificationString += "-The passenger with ID " + currentPassenger.Passenger.GUCID + " and name " + currentPassenger.Passenger.Name + " has cancelled his request. You can accept another one in their place.-"
				//remove him from db
				DB.DeletePassengerRequest(postID.(uint64), currentPassenger.Passenger.GUCID)
			}
		}

		CarpoolPost, err := DB.GetPostByID(session["postID"].(uint64))
		if err == nil {
			//no problem
			possiblePassengers := CarpoolPost[0].PossiblePassengers
			for i := 0; i < len(possiblePassengers); i++ {
				notificationString += "-The passenger with ID " + possiblePassengers[i] + " wants to ride with you-"
			}
		}
	}
	if notificationString == "" {
		return "-There are no new notifications-", nil
	}
	return notificationString, nil
}

// Function to write out a JSON response.
func writeJSON(res http.ResponseWriter, data JSON) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
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
