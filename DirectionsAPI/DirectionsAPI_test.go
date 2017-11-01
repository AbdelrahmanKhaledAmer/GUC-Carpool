package DirectionsAPI

import (
	"fmt"
	"testing"
)

//testing function

func TestDirection(t *testing.T) {

	//success with address
	s, err := GetRoute("German University IN cairo", "Cairo Festival City")
	if err != nil {
		fmt.Println("Error:" + err.Error())
		t.Error("failed test1 ")

	} else {
		fmt.Println(s)
	}

	//success with latitude longitude pairs
	fmt.Println("*****************************************")
	s, err = GetRoute("30.0320,31.4085", "29.9866,31.4414")
	if err != nil {
		fmt.Println("Error:" + err.Error())
		t.Error("failed test1 ")

	} else {
		fmt.Println(s)
	}
	fmt.Println("*****************************************")
	//fail
	s, err = GetRoute("German Universijkdsdkjty IN cairo", "Cairo Festjkdsjksdival City")
	if err != nil {
		fmt.Println("Error:" + err.Error())
	} else {
		fmt.Println(s)
		t.Error("failed test1 ")

	}
	fmt.Println("*****************************************")
}
