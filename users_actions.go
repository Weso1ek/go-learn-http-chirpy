package main

import (
	"database/sql"
	"encoding/json"
	"github.com/Weso1ek/chirpy/internal/auth"
	"github.com/Weso1ek/chirpy/internal/database"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.dbQueries.GetUserByLogin(r.Context(), sql.NullString{String: params.Email, Valid: true})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found", nil)
	}

	errCompare := auth.CheckPasswordHash(user.HashedPassword.String, params.Password)
	if errCompare != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found", nil)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email:     user.Email.String,
		},
	})
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, errPass := auth.HashPassword(params.Password)
	if errPass != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", errPass)
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          sql.NullString{String: params.Email, Valid: true},
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
			Email:     user.Email.String,
		},
	})
}
