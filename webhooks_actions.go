package main

import (
	"encoding/json"
	"fmt"
	"github.com/Weso1ek/chirpy/internal/auth"
	"github.com/google/uuid"
	"net/http"
)

func (cfg *apiConfig) handlerUserUpgrade(w http.ResponseWriter, r *http.Request) {
	type parametersData struct {
		UserId string `json:"user_id"`
	}

	type parameters struct {
		Event string         `json:"event"`
		Data  parametersData `json:"data"`
	}

	token, errToken := auth.GetAPIKey(r.Header)
	if errToken != nil {
		respondWithError(w, http.StatusUnauthorized, "No token", errToken)
		return
	}

	if token != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "No token not equal", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Event not supported", fmt.Errorf("Event not supported"))
	}

	userId, _ := uuid.Parse(params.Data.UserId)

	_, errUpdate := cfg.dbQueries.UpdateUserRed(r.Context(), userId)
	if errUpdate != nil {
		respondWithError(w, http.StatusNotFound, "Failed to update user", errUpdate)
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
