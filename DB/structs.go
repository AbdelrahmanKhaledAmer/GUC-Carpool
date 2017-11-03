package DB

import (
	"strconv"
	"time"
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

// CarpoolToString : Take a Carpool Request as a subject and returns a string describing it.
func (c *CarpoolRequest) CarpoolToString() string {
	str := "{ PostID: " + strconv.FormatUint(c.PostID, 10)
	str += ",</br> Name: " + c.Name
	str += ",</br> GUCID: " + c.GUCID
	str += ",</br> FromGUC: " + strconv.FormatBool(c.FromGUC)
	str += ",</br> Latitude: " + strconv.FormatFloat(c.Latitude, 'f', -1, 64)
	str += ",</br> Longitude: " + strconv.FormatFloat(c.Longitude, 'f', -1, 64)
	str += ",</br> Start Time: " + c.StartTime.Format("Jan 2, 2006 at 3:04pm (EET)")
	str += ",</br> Available Seats: " + strconv.FormatInt(int64(c.AvailableSeats), 10)
	str += ",</br> Current Passengers: [ "
	for i := 0; i < len(c.CurrentPassengers); i++ {
		str += c.CurrentPassengers[i]
		if i != (len(c.CurrentPassengers) - 1) {
			str += ", "
		}
	}
	str += " ],</br> Possible Passengers: [ "
	for i := 0; i < len(c.PossiblePassengers); i++ {
		str += c.PossiblePassengers[i]
		if i != (len(c.PossiblePassengers) - 1) {
			str += ", "
		}
	}
	str += " ] }"
	return str
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

//Passenger : current or possible Passenger riding the car
type Passenger struct {
	GUCID string
	Name  string
}

//PassengerRequest : request made by students to,from guc
type PassengerRequest struct {
	Passenger Passenger
	PostID    uint64
	Notify    uint8
}

//NewPassenger create new Passenger  return the newly created Passenger
func NewPassenger(GUCID string, Name string) (req Passenger, err error) {

	req = Passenger{
		GUCID: GUCID,
		Name:  Name,
	}
	return req, nil
}

//NewPassengerRequest : request made by students to,from guc
func NewPassengerRequest(GUCID string, Name string, PostID uint64, Notify uint8) (req PassengerRequest, err error) {
	pass, _ := NewPassenger(GUCID, Name)
	req = PassengerRequest{
		Passenger: pass,
		PostID:    PostID,
		Notify:    Notify,
	}
	return req, nil
}
