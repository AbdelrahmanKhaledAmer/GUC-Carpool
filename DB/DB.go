package DB

import (
	"fmt"
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
	IsDrop = true
)

//CarpoolRequest begin
/////////////////////////////

//UpdateDB : update a carpool post
func UpdateDB(postid uint64, Longitude float64, Latitude float64, FromGUC bool, AvailableSeats int, CurrentPassengers []string, PossiblePassengers []string) error {
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}
	c := session.DB("Carpool").C("CarpoolRequest")
	colQuerier := bson.M{"_id": postid}
	change := bson.M{"$set": bson.M{"currentpassengers": CurrentPassengers, "possiblepassengers": PossiblePassengers, "longitude": Longitude, "latitude": Latitude, "fromguc": FromGUC, "availableseats": AvailableSeats, "time": time.Now()}}
	err = c.Update(colQuerier, change)
	if err != nil {
		fmt.Println(err.Error())
	}
	if err != nil {
		return err
	}
	fmt.Println("update succ")
	return nil
}

//QueryAll return all the requests in the DB
func QueryAll() ([]CarpoolRequest, error) { //TODO should be renamed with the package name
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}

	c := session.DB("Carpool").C("CarpoolRequest")
	var results []CarpoolRequest
	err = c.Find(bson.M{}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetPostByID : return 1 post matching specific ID
func GetPostByID(PostID uint64) ([]CarpoolRequest, error) {
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}

	c := session.DB("Carpool").C("CarpoolRequest")
	var results []CarpoolRequest
	err = c.Find(bson.M{"_id": PostID}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}

	return results, nil

}

//InsertDB insert func
func InsertDB(req *CarpoolRequest) error {

	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}

	c := session.DB("Carpool").C("CarpoolRequest")

	//auto incr
	ai.Connect(c)
	req.PostID = ai.Next("CarPoolRequest")
	//auto incr

	err = c.Insert(req)
	if err != nil {
		return err
	}
	fmt.Println("insertion succ")
	return nil
}

//DeleteDB : Delete carpool request
func DeleteDB(PostID uint64) error {

	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}

	c := session.DB("Carpool").C("CarpoolRequest")

	err = c.Remove(bson.M{"postid": PostID})
	if err != nil {
		fmt.Printf("remove fail %v\n", err)
		return err
	}
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
	defer session.Close()
	if err != nil {
		return err
	}
	pass, _ := NewPassenger(GUCID, Name)
	c := session.DB("Carpool").C("PassengerRequest")
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
	return nil
}

//QueryAllPassengerRequests return all the requests in the DB
func QueryAllPassengerRequests() ([]PassengerRequest, error) { //TODO should be renamed with the package name
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}

	c := session.DB("Carpool").C("PassengerRequest")
	var results []PassengerRequest
	err = c.Find(bson.M{}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}

	return results, nil

}

// GetPassengerRequestByGUCID : return 1 passenger matching specific ID
func GetPassengerRequestByGUCID(GUCID string) ([]PassengerRequest, error) {
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}

	c := session.DB("Carpool").C("PassengerRequest")
	var results []PassengerRequest
	err = c.Find(bson.M{"passenger.gucid": GUCID}).All(&results) //c.Find(bson.M{"name": "Ahmed"}).All(&results) for filtering
	if err != nil {
		return nil, err
	}

	return results, nil

}

//InsertPassengerRequest insert func
func InsertPassengerRequest(req *PassengerRequest) error {

	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}

	c := session.DB("Carpool").C("PassengerRequest")

	err = c.Insert(req)
	if err != nil {

		return err

	}
	fmt.Println("insertion succ")
	return nil
}

//DeletePassengerRequest : Delete passengerrequest
func DeletePassengerRequest(GUCID string) error {

	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}

	c := session.DB("Carpool").C("PassengerRequest")

	err = c.Remove(bson.M{"passenger.gucid": GUCID})
	if err != nil {
		fmt.Printf("remove passenger fail %v\n", err)
		return err
	}
	return nil
}

//Passenger request functions end

func initDBSession() (*mgo.Session, error) {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	// Drop Database
	if IsDrop {
		err = session.DB("gucCarpool").DropDatabase()
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}
