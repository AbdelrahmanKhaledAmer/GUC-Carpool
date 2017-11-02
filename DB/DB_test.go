package DB

import (
	"fmt"
	"testing"
)

// func TestGetPostByID(t *testing.T) {
// 	PostID := 70
// 	//does not exist
// 	res, err := GetPostByID(uint64(PostID))
// 	fmt.Println(res)
// 	if err != nil {
// 		t.Error("failed")
// 	}

// 	//exist
// 	PostID = 3
// 	res, err = GetPostByID(uint64(PostID))
// 	fmt.Println(res)
// 	if err != nil {
// 		t.Error("failed")
// 	}
// }

// func TestUpdate(t *testing.T) {

// 	//test update code
// 	var possiblePass []string
// 	var currentPass []string
// 	possiblePass = append(possiblePass, "34-9791", "34-14269")
// 	currentPass = append(currentPass, "31-1111", "34-6141")
// 	//non existing post
// 	// err := UpdateDB(7, 31, 32, true, 4, currentPass, possiblePass,time.Now())
// 	// if err == nil {
// 	// 	t.Error("problem")
// 	// }
// 	s1, err := QueryAll()
// 	fmt.Println(s1)
// 	//existing post
// 	err = UpdateDB(2, 25, 32, false, 3, currentPass, possiblePass, time.Now())
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	s1, err = QueryAll()
// 	fmt.Println(s1)

// }

// func TestPassengerInsert(t *testing.T) {
// 	passreq, _ := NewPassengerRequest("3-4578", "koko", 2, 2)
// 	InsertPassengerRequest(&passreq)
// }

// func TestPassengerQueryALL(t *testing.T) {
// 	fmt.Println(QueryAllPassengerRequests())
// }

// func TestPassengerQueryONE(t *testing.T) {
// 	fmt.Println(GetPassengerRequestByGUCID("4"))
// 	fmt.Println(GetPassengerRequestByGUCID("3-4578"))
// }

// func TestUpdatePassenger(t *testing.T) {
// 	fmt.Println(UpdatePassengerRequest("4", "wawa", 4, 2))
// 	fmt.Println(UpdatePassengerRequest("3-4578", "hamada", 2, 1))
// }

// func TestRemovePassenger(t *testing.T) {
// 	fmt.Println(DeletePassengerRequest("4"))
// 	fmt.Println(DeletePassengerRequest("3-4578"))
// }

func TestAcceptPassenger(t *testing.T) {
	fmt.Println(AcceptPassenger("34-9791", 2))
}
