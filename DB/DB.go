package DB

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/night-codes/mgo-ai"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
you should go get these packages first
"gopkg.in/mgo.v2"
"github.com/night-codes/mgo-ai"
"gopkg.in/mgo.v2/bson"
*/

var (
	//IsDrop show the status of DB
	IsDrop = false
)

//CarpoolRequest begin
/////////////////////////////

//UpdateDB : update a carpool post
func UpdateDB(postid uint64, Longitude float64, Latitude float64, FromGUC bool, AvailableSeats int, CurrentPassengers []string, PossiblePassengers []string, Time time.Time) error {
	session, err := initDBSession()
	if err != nil {
		return err
	}
	c := session.DB("carpool").C("CarpoolRequest")
	colQuerier := bson.M{"_id": postid}
	change := bson.M{"$set": bson.M{"starttime": Time, "currentpassengers": CurrentPassengers, "possiblepassengers": PossiblePassengers, "longitude": Longitude, "latitude": Latitude, "fromguc": FromGUC, "availableseats": AvailableSeats, "time": time.Now()}}
	err = c.Update(colQuerier, change)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err != nil {
		return err
	}
	fmt.Println("update succ")
	defer session.Close()
	return nil
}

//QueryAll return all the requests in the DB
func QueryAll() ([]CarpoolRequest, error) { //TODO should be renamed with the package name
	session, err := initDBSession()
	if err != nil {
		return nil, err
	}

	c := session.DB("carpool").C("CarpoolRequest")
	var results []CarpoolRequest
	err = c.Find(bson.M{}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return results, nil
}

// GetPostByID : return 1 post matching specific ID
func GetPostByID(PostID uint64) ([]CarpoolRequest, error) {
	session, err := initDBSession()
	if err != nil {
		return nil, err
	}

	c := session.DB("carpool").C("CarpoolRequest")
	var results []CarpoolRequest
	err = c.Find(bson.M{"_id": PostID}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return results, nil
}

//InsertDB insert func
func InsertDB(req *CarpoolRequest) error {

	session, err := initDBSession()
	if err != nil {
		return err
	}

	c := session.DB("carpool").C("CarpoolRequest")

	//auto incr
	ai.Connect(c)
	req.PostID = ai.Next("CarPoolRequest")
	//auto incr

	err = c.Insert(req)
	if err != nil {
		return err
	}
	fmt.Println("insertion succ")
	defer session.Close()
	return nil
}

//DeleteDB : Delete carpool request
func DeleteDB(PostID uint64) error {

	session, err := initDBSession()
	if err != nil {
		return err
	}

	c := session.DB("carpool").C("CarpoolRequest")

	err = c.Remove(bson.M{"_id": PostID})
	if err != nil {
		fmt.Printf("remove fail %v\n", err)
		return err
	}
	defer session.Close()
	return nil
}

//CarpoolRequest functions end
//////////////////////////////////////////////////////////////////

// PassengerRequest Functions begin

//since passenger can ride with only one carpool the GUCID will be a unique Identifier
//TODO enforce unique GUCID

//UpdatePassengerRequest : update passenger
func UpdatePassengerRequest(GUCID string, Name string, PostID uint64, Notify uint8) error {
	session, err := initDBSession()
	if err != nil {
		return err
	}
	pass, _ := NewPassenger(GUCID, Name)
	c := session.DB("carpool").C("PassengerRequest")
	colQuerier := bson.M{"passenger.gucid": GUCID}
	change := bson.M{"$set": bson.M{"passenger": pass, "postid": PostID, "notify": Notify}}
	err = c.Update(colQuerier, change)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err != nil {
		return err
	}
	fmt.Println(" passenger update succ")
	defer session.Close()
	return nil
}

//QueryAllPassengerRequests return all the requests in the DB
func QueryAllPassengerRequests() ([]PassengerRequest, error) { //TODO should be renamed with the package name
	session, err := initDBSession()
	if err != nil {
		return nil, err
	}

	c := session.DB("carpool").C("PassengerRequest")
	var results []PassengerRequest
	err = c.Find(bson.M{}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return results, nil
}

// GetPassengerRequestByGUCID : return 1 passenger matching specific ID
func GetPassengerRequestByGUCID(GUCID string) ([]PassengerRequest, error) {
	session, err := initDBSession()
	if err != nil {
		return nil, err
	}

	c := session.DB("carpool").C("PassengerRequest")
	var results []PassengerRequest
	err = c.Find(bson.M{"passenger.gucid": GUCID}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return results, nil
}

// GetPassengerRequestsByPostID : return all passengers matching specific postID
func GetPassengerRequestsByPostID(postID uint64) ([]PassengerRequest, error) {
	session, err := initDBSession()
	if err != nil {
		return nil, err
	}

	c := session.DB("carpool").C("PassengerRequest")
	var results []PassengerRequest
	err = c.Find(bson.M{"passenger.postid": postID}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return results, nil
}

//InsertPassengerRequest insert func
func InsertPassengerRequest(req *PassengerRequest) error {

	session, err := initDBSession()
	if err != nil {
		return err
	}

	c := session.DB("carpool").C("PassengerRequest")

	err = c.Insert(req)
	if err != nil {

		return err

	}
	fmt.Println("insertion succ")
	defer session.Close()
	return nil
}

//DeletePassengerRequest : Delete passengerrequest
func DeletePassengerRequest(GUCID string) error {
	session, err := initDBSession()
	if err != nil {
		return err
	}

	c := session.DB("carpool").C("PassengerRequest")

	err = c.Remove(bson.M{"passenger.gucid": GUCID})
	if err != nil {
		fmt.Printf("remove passenger fail %v\n", err)
		return err
	}
	defer session.Close()
	return nil
}

//RejectPassenger : Removes a passenger from the carpool, and sets the notification variable.
func RejectPassenger(GUCID string, PostID uint64) error {
	carpoolRequests, err := GetPostByID(PostID)
	if err != nil {
		return err
	}
	if len(carpoolRequests) == 0 {
		return errors.New("no post with this ID")
	}
	wasCurrent := false
	carpoolRequest := carpoolRequests[0]
	possiblePassengers := carpoolRequest.PossiblePassengers
	currentPassengers := carpoolRequest.CurrentPassengers
	for idx, val := range possiblePassengers {
		if strings.EqualFold(val, GUCID) {
			possiblePassengers = append(possiblePassengers[:idx], possiblePassengers[idx+1:]...)
			break
		}
	}
	for idx, val := range currentPassengers {
		if strings.EqualFold(val, GUCID) {
			currentPassengers = append(currentPassengers[:idx], currentPassengers[idx+1:]...)
			wasCurrent = true
			break
		}
	}
	availableSeats := carpoolRequest.AvailableSeats
	if wasCurrent {
		availableSeats++
	}
	err = UpdateDB(carpoolRequest.PostID, carpoolRequest.Longitude, carpoolRequest.Latitude, carpoolRequest.FromGUC, availableSeats, currentPassengers, possiblePassengers, carpoolRequest.StartTime)
	if err != nil {
		return err
	}

	passengerRequests, err := GetPassengerRequestByGUCID(GUCID)
	if err != nil {
		return err
	}

	if len(passengerRequests) == 0 {
		return errors.New("The passenger you're trying to reject did not request a carpool")
	}
	if PostID != passengerRequests[0].PostID {
		return errors.New("You can not reject a passenger that did not request your carpool")
	}

	passengerRequest := passengerRequests[0]
	err = UpdatePassengerRequest(passengerRequest.Passenger.GUCID, passengerRequest.Passenger.Name, passengerRequest.PostID, 0) //notify
	if err != nil {
		return err
	}

	return nil
}

// AcceptPassenger : adds a passenger to the currentPassengers in the carpool, and sets the notification variable.
func AcceptPassenger(GUCID string, PostID uint64) error {

	posts, err := GetPostByID(PostID)
	if err != nil {
		return err
	}
	if (len(posts)) == 0 {
		return errors.New("no post with this id")
	}
	possiblepassengers := posts[0].PossiblePassengers
	currentpassengers := posts[0].CurrentPassengers
	availableseats := posts[0].AvailableSeats

	passengers, err := GetPassengerRequestByGUCID(GUCID)
	if err != nil {
		return err
	}
	if len(passengers) == 0 {
		return errors.New("The passenger you're trying to accept did not request your carpool")
	}
	if PostID != passengers[0].PostID {
		return errors.New("you can not accept a passenger that did not request your carpool")
	}

	if availableseats == 0 {
		return errors.New("no seat available")
	}

	for index := 0; index < len(possiblepassengers); index++ {
		if possiblepassengers[index] == GUCID {
			currentpassengers = append(currentpassengers, GUCID)
			possiblepassengers = append(possiblepassengers[:index], possiblepassengers[index+1:]...)
			UpdatePassengerRequest(GUCID, passengers[0].Passenger.Name, PostID, 2) //notify
			UpdateDB(PostID, posts[0].Longitude, posts[0].Latitude, posts[0].FromGUC, availableseats-1, currentpassengers, possiblepassengers, posts[0].Time)
			return nil
		}
	}
	return errors.New("not a possible passenger")
}

//Passenger request functions end

func initDBSession() (*mgo.Session, error) {
	uri := "mongodb://carpool:carpool@ds245715.mlab.com:45715/carpool"
	session, err := mgo.Dial(uri)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	session.SetSafe(&mgo.Safe{})
	// Drop Database
	if IsDrop {
		err = session.DB("carpool").DropDatabase()
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}
