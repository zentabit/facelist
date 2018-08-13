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
    "fmt"
    "github.com/kelseyhightower/envconfig"
    "google.golang.org/appengine"
    "google.golang.org/appengine/urlfetch"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "sort"
    "strings"
)

const (
    // IndexTmpl is the template for the index page.
    IndexTmpl = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="application/json">
    <title>Facelist</title>
    <style>
        .container {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr 1fr;
            grid-gap: 10px;
            background-color: #fff;
            color: #444;
        }

        .user {
          //background-color: #444;
          //color: #fff;
          //border-radius: 5px;
          padding: 10px;
        }
    </style>
  </head>
  <body>
    <div class="container">
    {{range .Members}}
        <div class="user">
            <h3>{{if .Profile.RealName}}{{.Profile.RealName}}{{else}}{{.Name}}{{end}}
            <a href="https://{{- $.SlackTeam -}}.slack.com/messages/{{.Id}}/team/{{.Id}}/" target="_blank">
                <img src="https://a.slack-edge.com/436da/marketing/img/meta/favicon-32.png" title="Contact {{.Profile.FirstName}} on Slack" width="16" height="16"/>
            </a>
            </h3>
            <a href="https://{{- $.SlackTeam -}}.slack.com/messages/{{.Id}}/team/{{.Id}}/" target="_blank">
                <img src="{{.Profile.Image}}" title="Contact {{.Profile.FirstName}} on Slack"/>
            </a>
        </div>
    {{end}}
    </div>
  </body>
</html>
`
)

var (
    cfg           config
    userlist      UserList
    IndexTemplate = template.Must(template.New("index").Parse(IndexTmpl))
)

type (
    config struct {
        EmailFilter   string `envconfig:"EMAIL_FILTER" default:""`
        SlackApiToken string `envconfig:"SLACK_API_TOKEN"`
        SlackTeam     string `envconfig:"SLACK_TEAM"`
    }

    UserList struct {
        SlackTeam string
        Members   []User `json:members`
    }

    User struct {
        Name    string  `json:"name"`
        Id      string  `json:"id"`
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
    if err := envconfig.Process("facelist", &cfg); err != nil {
        log.Fatalf("failed to parse config: %v\n", err)
    }
    if cfg.SlackTeam == "" {
        log.Fatalf("SLACK_TEAM is not set!")
        os.Exit(1)
    }
    if cfg.SlackApiToken == "" {
        log.Fatalf("SLACK_API_TOKEN is not set!")
        os.Exit(1)
    }
    userlist.SlackTeam = cfg.SlackTeam
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

    ctx := appengine.NewContext(r)
    client := urlfetch.Client(ctx)
    url := fmt.Sprintf("https://slack.com/api/users.list?token=%s", cfg.SlackApiToken)
    resp, err := client.Get(url)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    body, _ := ioutil.ReadAll(resp.Body)

    err = json.Unmarshal(body, &userlist)
    if err != nil {
        log.Fatal(err)
    }

    // Filter out deleted accounts, bots and users without @tink.se email adresses
    filteredUsers := []User{}
    for i := range userlist.Members {
        user := userlist.Members[i]
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
    appengine.Main()
}
