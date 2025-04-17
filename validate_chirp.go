package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type returnVals struct {
		//Valid bool `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	type returnValsError struct {
		Error string `json:"error"`
	}

	if len(params.Body) > 140 {
		respBodyError := returnValsError{
			Error: "Chirp is too long",
		}

		datErr, _ := json.Marshal(respBodyError)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(datErr)
	}

	cleanedBody := strings.Split(params.Body, " ")
	for i, j := range cleanedBody {
		if strings.ToLower(j) == "kerfuffle" || strings.ToLower(j) == "sharbert" || strings.ToLower(j) == "fornax" {
			cleanedBody[i] = "****"
		}
	}

	// profane words
	respBody := returnVals{
		CleanedBody: strings.Join(cleanedBody, " "),
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
