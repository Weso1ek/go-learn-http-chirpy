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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: sql.NullTime{Time: time.Now().UTC().Add(time.Hour * 24 * 60), Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt.Time,
			UpdatedAt:   user.UpdatedAt.Time,
			Email:       user.Email.String,
			IsChirpyRed: user.IsChirpyRed.Bool,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	hashedPassword, errPass := auth.HashPassword(params.Password)
	if errPass != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password", errPass)
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          sql.NullString{String: params.Email, Valid: true},
		HashedPassword: sql.NullString{String: hashedPassword, Valid: true},
		ID:             userID,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:          userID,
			Email:       user.Email.String,
			IsChirpyRed: user.IsChirpyRed.Bool,
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
			ID:          user.ID,
			CreatedAt:   user.CreatedAt.Time,
			UpdatedAt:   user.UpdatedAt.Time,
			Email:       user.Email.String,
			IsChirpyRed: user.IsChirpyRed.Bool,
		},
	})
}
