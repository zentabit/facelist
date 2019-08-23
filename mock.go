package main

import (
	"encoding/json"
	"io/ioutil"
)

func getMockedUsers() UserList {
	var userlist = UserList{SlackTeam: "tink"}
	file, _ := ioutil.ReadFile("mocks/users.list.json")

	_ = json.Unmarshal([]byte(file), &userlist)
	return userlist
}
