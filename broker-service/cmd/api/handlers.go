package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

//Json Format for Sending data from any request
type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

//Create json format to handle authentication
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	log.Println("In the Broker Handlers")
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}
	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) TestAuthenticate(w http.ResponseWriter, r *http.Request) {
	//Marshal data into json
	jsonData, err := json.Marshal(r.Body)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//Create a new http request
	request, err := http.NewRequest("POST", "http://authentication-service:8081", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//Create a new http Client and get a response
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	//Decode the json response from other service
	var responseFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&responseFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//Write Response From the Service
	var payload jsonResponse
	payload.Error = false
	payload.Message = responseFromService.Message

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	//Read the request we into Request Payload
	var requestPayload RequestPayload

	//check error
	err := app.readJSON(w, *r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
	}

	//take action based on request we are getting
	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)

	default:
		app.errorJSON(w, errors.New("unknown action"))

	}
}

//local function to help with authentication
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	//Create json to send to the auth services
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	//Creating a new http request and use bytes to convert data
	request, err := http.NewRequest("POST", "http://authentication-service:8081/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	//create a new client to call service
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	//make sure we get back the correct status

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	//Read the response body into a variable
	var jsonFromService jsonResponse

	//decode response for auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
	}

	//sent data back
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}
