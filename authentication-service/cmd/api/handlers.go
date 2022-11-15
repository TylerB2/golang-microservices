package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) AuthTest(w http.ResponseWriter, r *http.Request) {
	log.Println("In the AUTHTest Handlers")
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Hit the Authentication Service"

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	log.Print("in the authenticate Handler")
	var requestpayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestpayload)
	if err != nil {
		log.Print("Unable to read Json")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	//validate user against database
	user, err := app.Models.User.GetByEmail(requestpayload.Email)
	if err != nil {
		log.Print("Invalid Email ")
		app.errorJSON(w, errors.New("invalid Credentials"), http.StatusBadRequest)
		return
	}

	//Check password
	valid, err := user.PasswordMatches(requestpayload.Password)
	if err != nil {
		log.Print("lllogg error ")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	if err != nil || !valid {
		log.Print("Passwords do not match")
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	//Log Authentication in the logger service
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	//return or write the response
	var payload jsonResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Logged in user %s", user.Email)
	payload.Data = user

	log.Println("Logged in user ", user.Email)
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service:8082/log"

	//Create request

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Problem with writing to logger service", err)
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		log.Println("Problem with writing to logger service ", err)
		return err
	}
	return nil
}
