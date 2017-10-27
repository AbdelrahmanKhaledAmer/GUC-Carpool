package DB

import (
	"fmt"
	"time"

	"github.com/night-codes/mgo-ai"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
you should go get these packages first
"gopkg.in/mgo.v2"
"github.com/night-codes/mgo-ai"
"gopkg.in/mgo.v2/bson"
to test code in main func change the package to main and put it in separate folder
*/

var (
	//IsDrop show the status of DB
	IsDrop = true
)

//CarpoolRequest : request made by students to,from guc
type CarpoolRequest struct {
	GUCID              string
	Longitude          float64
	Latitude           float64
	PostID             uint64 `bson:"_id,omitempty"`
	Time               time.Time
	StartTime          time.Time // time parsing done outside database for multiple format
	CurrentPassengers  []string
	PossiblePassengers []string
	Name               string
	FromGUC            bool
	AvailableSeats     int
}

//NewCarpool create new carpool request return the newly created request
func NewCarpool(GUCID string, Longitude float64, Latitude float64, Name string, FromGUC bool, AvailableSeats int, StartTime string) (req CarpoolRequest, err error) {
	mySlice1 := make([]string, 0)
	stTime, err := time.Parse("Jan 2, 2006 at 3:04pm (EET)", StartTime)
	if err != nil {
		return req, err
	}
	req = CarpoolRequest{
		GUCID:              GUCID,
		Longitude:          Longitude,
		Latitude:           Latitude,
		Time:               time.Now(),
		StartTime:          stTime,
		CurrentPassengers:  mySlice1,
		PossiblePassengers: mySlice1,
		Name:               Name,
		FromGUC:            FromGUC,
		AvailableSeats:     AvailableSeats,
	}
	return req, nil
}

/*func main() {
	// test Insert Data
	//should check if start time is a valid time format first
	newC, err := NewCarpool("34-111", 31.25, 32.56, "mohamed", true, 5, "Jan 2, 2006 at 3:04pm (EET)") //newC will have default values
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(newC)
	}
	err = InsertDB(&newC)
	if err != nil {
		fmt.Println(err.Error())
	}
	//test Query code
	rs, err := QueryAll()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(rs)
	//test update code
	err = updateDB(1, 31, 32, true, 4)
	if err != nil {
		fmt.Println(err.Error())
	}
}*/

func updateDB(postid, Longitude float64, Latitude float64, FromGUC bool, AvailableSeats int) error {
	session, err := initDBSession()
	defer session.Close()
	if err != nil {
		return err
	}
	c := session.DB("Carpool").C("CarpoolRequest")
	colQuerier := bson.M{"_id": postid}
	change := bson.M{"$set": bson.M{"longitude": Longitude, "latitude": Latitude, "fromguc": FromGUC, "availableseats": AvailableSeats, "time": time.Now()}}
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
