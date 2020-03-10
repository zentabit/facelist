/*
Copyright 2018 Tink AB

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	cfg           config
	userlist      UserList
	IndexTemplate = template.Must(template.ParseFiles("templates/index.html"))
)

type (
	config struct {
		EmailFilter   string `yaml:"emailFilter"`
		SlackAPIToken string `yaml:"slackAPIToken"`
		SlackTeam     string `yaml:"slackTeam"`
	}

	UserList struct {
		SlackTeam string
		Members   []User `json:"members"`
	}

	User struct {
		Name    string  `json:"name"`
		Id      string  `json:"id"`
		TeamId  string  `json:"team_id"`
		IsBot   bool    `json:"is_bot"`
		Deleted bool    `json:"deleted"`
		Profile Profile `json:"profile"`
	}

	Profile struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		RealName  string `json:"real_name"`
		Title     string `json:"title"`
		Image     string `json:"image_192"`
		Phone     string `json:"phone"`
		Email     string `json:"email"`
		Status    string `json:"status_text"`
	}
)

func init() {
	log.Println("Starting facelist")

	configFile := flag.String("config", "facelist.yaml", "Configuration file to load")
	flag.Parse()
	b, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Unable to read config: %v\n", err)
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalf("Unable to decode config: %v\n", err)
	}

	if cfg.SlackTeam == "" {
		log.Fatalf("SLACK_TEAM is not set!")
		os.Exit(1)
	}
	if cfg.SlackAPIToken == "" {
		log.Fatalf("SLACK_API_TOKEN is not set!")
		os.Exit(1)
	}
	userlist.SlackTeam = cfg.SlackTeam
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	client := http.Client{Timeout: time.Duration(5 * time.Second)}

	// Use mocked data for local dev
	if cfg.SlackAPIToken == "<SECRET_API_TOKEN_GOES_HERE>" {
		userlist = getMockedUsers()
	} else {
		req, err := http.NewRequest("GET", "https://slack.com/api/users.list/", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Add("Authorization", "Bearer "+cfg.SlackAPIToken)

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &userlist)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Filter out deleted accounts, bots and users without @tink.se email addresses
	filteredUsers := []User{}
	for _, user := range userlist.Members {
		if !user.Deleted && !user.IsBot && strings.HasSuffix(user.Profile.Email, cfg.EmailFilter) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Sort users on first name
	sort.SliceStable(filteredUsers, func(i, j int) bool {
		return strings.ToLower(filteredUsers[i].Profile.RealName) < strings.ToLower(filteredUsers[j].Profile.RealName)
	})

	userlist.Members = filteredUsers
	if err := IndexTemplate.Execute(w, userlist); err != nil {
		log.Printf("Failed to execute index template: %v\n", err)
		http.Error(w, "Oops. That's embarrassing. Please try again later.", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
