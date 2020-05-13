package main

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"fmt"
	"net/url"
	"net/http"
	"time"
	"strconv"
	"bytes"
	"os"
)

// ImgCacher represents the settings for the 
type ImgCacher struct {
	TenantID 		string `json: "TenantID"`
	ApplicationID   string `json: "ApplicationID"`
	ClientSecret    string `json: "ClientSecret"`

	tok Token
}

// Token struct holds the Microsoft Graph API authentication token used by GraphClient to authenticate API-requests to the ms graph API
type Token struct {
	Token_Type   string    `json: "token_type"`// should always be "Bearer" for msgraph API-calls
	Resource     string    `json: "resource"`// will most likely always be https://graph.microsoft.com, hence the BaseURL
	Access_Token string    `json: "access_token"`// the access-token itself
}

// LoginBaseURL represents the basic url used to acquire a token for the msgraph api
const LoginBaseURL string = "https://login.microsoftonline.com"

// BaseURL represents the URL used to perform all ms graph API-calls
const BaseURL string = "https://graph.microsoft.com"

// APIVersion represents the APIVersion of msgraph used by this implementation
const APIVersion string = "v1.0"

// ModMin: the most amount of Minutes a file can be Modified before it has to be updated
const ModMin = 1000

// gets a token, copied from go-msgraph
func (c *ImgCacher) getToken(v interface{}) error{
	if c.TenantID == "" {
		return fmt.Errorf("Tenant ID is empty")
	}
	resource := fmt.Sprintf("/%v/oauth2/token", c.TenantID)
	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("client_id", c.ApplicationID)
	data.Add("client_secret", c.ClientSecret)
	data.Add("resource", BaseURL)

	u, err := url.ParseRequestURI(LoginBaseURL)
	if err != nil {
		return fmt.Errorf("Unable to parse URI: %v", err)
	}

	u.Path = resource
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))

	if err != nil {
		return fmt.Errorf("HTTP Request Error: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))


	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP response error: %v of http.Request: %v", err, req.URL)
	}
	defer resp.Body.Close() // close body when func returns

	body, err := ioutil.ReadAll(resp.Body) // read body first to append it to the error (if any)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Hint: this will mostly be the case if the tenant ID can not be found, the Application ID can not be found or the clientSecret is incorrect.
		// The cause will be described in the body, hence we have to return the body too for proper error-analysis
		return fmt.Errorf("StatusCode is not OK: %v. Body: %v ", resp.StatusCode, string(body))
	}

	//fmt.Println("Body: ", string(body))

	if err != nil {
		return fmt.Errorf("HTTP response read error: %v of http.Request: %v", err, req.URL)
	}

	return json.Unmarshal(body, &v)
}

// DownloadImage downloads the image with user ID id, puts it in directory ./img/, modfified version of go-mgraph's makeAPIGETCall
func (c *ImgCacher) DownloadImage(id string) error{
	imgPath := "./img/" + id + ".jpg"
	
	// check if file exists 
	if fi, err := os.Stat(imgPath); err == nil && time.Now().Sub(fi.ModTime()).Minutes() < ModMin {
		return nil
	  
	} else if os.IsNotExist(err) || time.Now().Sub(fi.ModTime()).Minutes() > ModMin {
		reqURL, err := url.ParseRequestURI(BaseURL)
		if err != nil {
			return fmt.Errorf("Unable to parse URI %v: %v", BaseURL, err)
		}

		// Add Version to API-Call, the leading slash is always added by the calling func
		reqURL.Path = "/" + APIVersion + "/users/" + id + "/photo/$value"
		

		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			return fmt.Errorf("HTTP request error: %v", err)
		}	

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", c.tok.Access_Token)

		getParams := url.Values{}
		getParams.Add("$top", strconv.Itoa(999))
		req.URL.RawQuery = getParams.Encode() // set query parameters

		httpClient := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("HTTP response error: %v of http.Request: %v", err, req.URL)
		}
		defer resp.Body.Close() // close body when func returns

		body, err := ioutil.ReadAll(resp.Body) // read body first to append it to the error (if any)
		if resp.StatusCode == 404 {
			return fmt.Errorf("image not found, probably because the user didn't set one")
		} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
			// Hint: this will mostly be the case if the tenant ID can not be found, the Application ID can not be found or the clientSecret is incorrect.
			// The cause will be described in the body, hence we have to return the body too for proper error-analysis
			return fmt.Errorf("StatusCode is not OK: %v. Body: %v ", resp.StatusCode, string(body))
		}

		if err != nil {
			return fmt.Errorf("HTTP response read error: %v of http.Request: %v", err, req.URL)
		}

		err = ioutil.WriteFile(imgPath, body, 0644)
		if(err!=nil){
			log.Println(err)
		}
		return err

	  } else {
		// Schrödinger: file may or may not exist. See err for details.
		fmt.Println(err)
		return nil
	  }


	
}