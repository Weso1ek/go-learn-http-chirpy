package main

import (
	"encoding/json"
	"github.com/Weso1ek/chirpy/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")

	chirpUUID, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't list chirps", err)
		return
	}

	chirp, errDb := cfg.dbQueries.GetChirp(r.Context(), chirpUUID)

	type response struct {
		Chirp
	}

	if errDb != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp", errDb)
	}

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.ListChirps(r.Context())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't list chirps", err)
		return
	}

	var chirpsResp []Chirp

	for _, j := range chirps {
		chirpsResp = append(chirpsResp, Chirp{
			ID:        j.ID,
			CreatedAt: j.CreatedAt.Time,
			UpdatedAt: j.UpdatedAt.Time,
			Body:      j.Body,
			UserId:    j.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsResp)
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
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

	// CREATE CHIRP
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Time,
			UpdatedAt: chirp.UpdatedAt.Time,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
}
