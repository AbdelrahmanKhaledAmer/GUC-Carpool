package DB

import (
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