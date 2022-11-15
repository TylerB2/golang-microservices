package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
)

var mu sync.Mutex

//Determine the json message structure
type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:",omitempty"`
}

//Function to read JSON
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

//takes data and write it to json
func (app *Config) writeJSON(w http.ResponseWriter, Status int, data any, headers ...http.Header) error {
	//convert the data to json
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	//add headers if they are provided
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	//write the actual json Message
	mu.Lock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Status)
	_, err = w.Write(out)
	mu.Unlock()
	if err != nil {
		return err
	}
	return nil

}

func (app *Config) errorJSON(w http.ResponseWriter, err error, Status ...int) error {
	//set statusCode
	statusCode := http.StatusBadRequest

	//Check if Status is Provided
	if len(Status) > 0 {
		statusCode = Status[0]
	}
	//Write error
	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
