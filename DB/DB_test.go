package DB

import (
	"fmt"
	"testing"
)

func TestGetPostByID(t *testing.T) {
	PostID := 3
	res, err := GetPostByID(uint64(PostID))
	fmt.Println(res[0])
	if err != nil || res == nil || len(res) == 0 {
		t.Error("failed")
	}
}

func TestUpdate(t *testing.T) {

	//test update code
	var possiblePass []string
	var currentPass []string
	possiblePass = append(possiblePass, "Abdelrahman", "saher")
	currentPass = append(currentPass, "Ahmed", "Mohamed")
	err := UpdateDB(3, 31, 32, true, 4, currentPass, possiblePass)
	if err != nil {
		fmt.Println(err.Error())
	}

}
