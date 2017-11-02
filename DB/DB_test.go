package DB

import (
	"fmt"
	"testing"
)

func TestGetPostByID(t *testing.T) {
	PostID := 70
	//does not exist
	res, err := GetPostByID(uint64(PostID))
	fmt.Println(res)
	if err != nil {
		t.Error("failed")
	}

	//exist
	PostID = 3
	res, err = GetPostByID(uint64(PostID))
	fmt.Println(res)
	if err != nil {
		t.Error("failed")
	}
}

func TestUpdate(t *testing.T) {

	//test update code
	var possiblePass []string
	var currentPass []string
	possiblePass = append(possiblePass, "Abdelrahman", "saher")
	currentPass = append(currentPass, "Ahmed", "Mohamed")
	//non existing post
	// err := UpdateDB(7, 31, 32, true, 4, currentPass, possiblePass)
	// if err == nil {
	// 	t.Error("problem")
	// }
	s1, err := QueryAll()
	fmt.Println(s1)
	//existing post
	err = UpdateDB(3, 25, 32, false, 3, currentPass, possiblePass)
	if err != nil {
		fmt.Println(err.Error())
	}

	s1, err = QueryAll()
	fmt.Println(s1)

}

func TestPassengerInsert(t *testing.T) {
	passreq, _ := NewPassengerRequest("3-4578", "koko", 2, true)
	InsertPassengerRequest(&passreq)
}

func TestPassengerQueryALL(t *testing.T) {
	fmt.Println(QueryAllPassengerRequests())
}

func TestPassengerQueryONE(t *testing.T) {
	fmt.Println(GetPassengerRequestByGUCID("4"))
	fmt.Println(GetPassengerRequestByGUCID("3-4578"))
}

func TestUpdatePassenger(t *testing.T) {
	fmt.Println(UpdatePassengerRequest("4", "wawa", 4, true))
	fmt.Println(UpdatePassengerRequest("3-4578", "hamada", 2, false))
}

func TestRemovePassenger(t *testing.T) {
	fmt.Println(DeletePassengerRequest("4"))
	fmt.Println(DeletePassengerRequest("3-4578"))
}
