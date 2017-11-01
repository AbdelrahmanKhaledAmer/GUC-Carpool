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
