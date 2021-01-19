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
	"github.com/zentabit/go-msgraph"
	//"github.com/k3a/html2text"
	"gopkg.in/yaml.v2"
	//"time"
)

var (
	cfg           config
	userlist 	  msgraph.Users
	IndexTemplate = template.Must(template.ParseFiles("templates/index.html"))
	am			= "data/aboutme.json"
)
type (
	config struct {
		EmailFilter   string `yaml:"emailFilter"`
		GraphAPIToken string `yaml:"graphAPIToken"`
		ApplicationID string `yaml:"applicationID"`
		TenantID      string `yaml:"tenantID"`
		GroupID		  string `yaml:"groupID"`
	}
)
/*
aboutMes struct {
	Email string
	Role string 
}
*/


func init() {
	log.Println("Starting facelist")
	
	configFile := flag.String("config", "scouterna.yaml", "Configuration file to load")
	flag.Parse()
	b, err := ioutil.ReadFile(*configFile)

	if err != nil {
		log.Fatalf("Unable to read config: %v\n", err)
	}
	
	err = yaml.Unmarshal(b, &cfg)

	if err != nil {
		log.Fatalf("Unable to decode config: %v\n", err)
	}
	
	if cfg.ApplicationID == "" {
		log.Fatalf("appID is not set!")
		os.Exit(1)
	}
	if cfg.TenantID == "" {
		log.Fatalf("tenantID is not set!")
		os.Exit(1)
	}
	log.Println("Config initialised")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//client := http.Client{Timeout: time.Duration(5 * time.Second)}
	//var graphClient msgraph.GraphClient
	// Use mocked data for local dev
	if cfg.GraphAPIToken == "" {
		userlist = nil
	} else {
		graphClient, err := msgraph.NewGraphClient(cfg.TenantID, cfg.ApplicationID, cfg.GraphAPIToken)
		if err != nil {
    		log.Println("Credentials are probably wrong or system time is not synced: ", err)
		}
		
		var g msgraph.Group
		g, err = graphClient.GetGroup(cfg.GroupID)
		log.Println("Group fetched")

		userlist, err = g.ListMembers()
		log.Println("Users fetched")
		userlist2 := []msgraph.User{}

		// create temp aboutme store
		var aboutMes = make(map[string] string)
		//aboutMes["foo"] = "bar"
		// check if file is stale
		/*
		fi, err := os.Stat(am)
		if err == nil && time.Now().Sub(fi.ModTime()).Minutes() > 1000 {
			os.Remove(am)
		}

		// if file exists, use that info and load the userlist
		if _, err := os.Stat(am); err == nil {
			s,_ := ioutil.ReadFile(am)
			yaml.Unmarshal(s, aboutMes)
			for _,u := range userlist {
				u.AboutMe.Value = strings.SplitAfterN(html2text.HTML2Text(aboutMes[u.ID]), "\n", 2)[0]
				userlist2 = append(userlist2, u)
				
			}
		// otherwise, create the file again
		} else if os.IsNotExist(err) {
			
			for _,u := range userlist {
				tempU, _ := graphClient.GetUser(u.ID)
				tempU.AboutMe.Value = strings.SplitAfterN(html2text.HTML2Text(tempU.AboutMe.Value), "\n", 2)[0]
				userlist2 = append(userlist2, tempU)
				aboutMes[u.ID] = tempU.AboutMe.Value
			}
			s, _ := yaml.Marshal(aboutMes)
			ioutil.WriteFile(am, s, 0644)

		} else {
			// TODO implementera nån lösning här
		}
		*/

		if _, err := os.Stat(am); err == nil {
			s,_ := ioutil.ReadFile(am)
			json.Unmarshal(s, &aboutMes)

			for _,u := range userlist {
				log.Println(aboutMes[u.Mail])
				u.AboutMe.Value = aboutMes[u.Mail]
				//log.Println(u.AboutMe.Value)
				userlist2 = append(userlist2, u)
			}
		}

		log.Println("Aboutmes fetched")
		userlist = userlist2

		if err != nil {
			log.Printf(err.Error())
		}
	}

	// Filter out deleted accounts, bots and users without @tink.se email addresses
	
	filteredUsers := []msgraph.User{}
	
	for _, user := range userlist {
		if strings.HasSuffix(user.Mail, cfg.EmailFilter) {
			filteredUsers = append(filteredUsers, user)
		}
	}
	
	
	//Sort users on first name
	sort.SliceStable(filteredUsers, func(i, j int) bool {
		return strings.ToLower(filteredUsers[i].DisplayName) < strings.ToLower(filteredUsers[j].DisplayName)
	})
	

	userlist = filteredUsers
	if err := IndexTemplate.Execute(w, userlist); err != nil {
		log.Printf("Failed to execute index template: %v\n", err)
		http.Error(w, "Oops. That's embarrassing. Please try again later.", http.StatusInternalServerError)
	}

	igc := ImgCacher{}
	
	igc.ApplicationID = cfg.ApplicationID
	igc.ClientSecret = cfg.GraphAPIToken
	igc.TenantID = cfg.TenantID
	err := igc.getToken(&igc.tok)
	if(err!=nil){
		log.Println("fuck you")
	}
	log.Println("starting image download")
	for _, user := range userlist{
		//log.Println("Hej")
		err = igc.DownloadImage(user.ID)
		if(err!=nil){
			log.Println(err)
		}
	}
	log.Println("Finished!")
}

func main() {
	http.HandleFunc("/", indexHandler)
	fs := http.FileServer(http.Dir("img/"))
	http.Handle("/img/", http.StripPrefix("/img/", fs))
	http.ListenAndServe(":8080", nil)
}
