package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/olimeme/helpers"
	"github.com/olimeme/internal/auth"
	"github.com/olimeme/internal/database"
)

type ApiConfig struct {
    FileserverHits atomic.Int32
    Database       *database.Queries
    Platform       string
    JwtSecret      string
    PolkaKey       string
}

type User struct {
    ID          uuid.UUID `json:"id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Email       string    `json:"email"`
    IsChirpyRed bool      `json:"is_chirpy_red"`
}


func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.FileserverHits.Add(1)
		log.Println(cfg.FileserverHits.Load())
		next.ServeHTTP(res, req)
	})
}

func (cfg *ApiConfig) GetNumberOfReqs(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, cfg.FileserverHits.Load())))
}

func (cfg *ApiConfig) ClearNumberOfReqs(res http.ResponseWriter, req *http.Request) {
    if cfg.Platform != "dev" {
        res.WriteHeader(http.StatusForbidden)
        res.Write([]byte("Forbidden"))
        return
    }

    if err := cfg.Database.DeleteAllUsers(req.Context()); err != nil {
        res.WriteHeader(http.StatusInternalServerError)
        res.Write([]byte("could not delete users"))
        return
    }

    cfg.FileserverHits.Store(0)
    res.WriteHeader(http.StatusOK)
    res.Write([]byte("Reset metrics and deleted all users"))
}

func (cfg *ApiConfig) CreateUser(res http.ResponseWriter, req *http.Request) {
    type requestBody struct {
        Password string `json:"password"`
        Email    string `json:"email"`
    }

    var body requestBody
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, "Something went wrong")
        return
    }

    hashedPassword, err := auth.HashPassword(body.Password)
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not hash password")
        return
    }

    dbUser, err := cfg.Database.CreateUser(req.Context(), database.CreateUserParams{
        Email:          body.Email,
        HashedPassword: hashedPassword,
    })
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not create user")
        return
    }

    helpers.RespondWithJSON(res, http.StatusCreated, User{
        ID:        dbUser.ID,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
        IsChirpyRed:     dbUser.IsChirpyRed,
    })
}

func (cfg *ApiConfig) Login(res http.ResponseWriter, req *http.Request) {
    type requestBody struct {
        Password string `json:"password"`
        Email    string `json:"email"`
    }

    var body requestBody
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, "Something went wrong")
        return
    }

    dbUser, err := cfg.Database.GetUserByEmail(req.Context(), body.Email)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Incorrect email or password")
        return
    }

    match, err := auth.CheckPasswordHash(body.Password, dbUser.HashedPassword)
    if err != nil || !match {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Incorrect email or password")
        return
    }

    accessToken, err := auth.MakeJWT(dbUser.ID, cfg.JwtSecret, time.Hour)
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not create access token")
        return
    }

    refreshToken, err := auth.MakeRefreshToken()
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not create refresh token")
        return
    }

    _, err = cfg.Database.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
        Token:     refreshToken,
        UserID:    dbUser.ID,
        ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
    })
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not store refresh token")
        return
    }

    type response struct {
        User
        Token        string `json:"token"`
        RefreshToken string `json:"refresh_token"`
    }

    helpers.RespondWithJSON(res, http.StatusOK, response{
        User: User{
            ID:        dbUser.ID,
            CreatedAt: dbUser.CreatedAt,
            UpdatedAt: dbUser.UpdatedAt,
            Email:     dbUser.Email,
            IsChirpyRed: dbUser.IsChirpyRed,
        },
        Token:        accessToken,
        RefreshToken: refreshToken,
    })
}

func (cfg *ApiConfig) Refresh(res http.ResponseWriter, req *http.Request) {
    tokenString, err := auth.GetBearerToken(req.Header)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Missing token")
        return
    }

    dbUser, err := cfg.Database.GetUserFromRefreshToken(req.Context(), tokenString)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Invalid or expired refresh token")
        return
    }

    accessToken, err := auth.MakeJWT(dbUser.ID, cfg.JwtSecret, time.Hour)
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not create access token")
        return
    }

    helpers.RespondWithJSON(res, http.StatusOK, map[string]string{
        "token": accessToken,
    })
}

func (cfg *ApiConfig) Revoke(res http.ResponseWriter, req *http.Request) {
    tokenString, err := auth.GetBearerToken(req.Header)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Missing token")
        return
    }

    if err := cfg.Database.RevokeRefreshToken(req.Context(), tokenString); err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not revoke token")
        return
    }

    res.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) UpdateUser(res http.ResponseWriter, req *http.Request) {
    tokenString, err := auth.GetBearerToken(req.Header)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Missing or invalid token")
        return
    }

    userID, err := auth.ValidateJWT(tokenString, cfg.JwtSecret)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Invalid token")
        return
    }

    type requestBody struct {
        Password string `json:"password"`
        Email    string `json:"email"`
    }

    var body requestBody
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, "Something went wrong")
        return
    }

    hashedPassword, err := auth.HashPassword(body.Password)
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not hash password")
        return
    }

    dbUser, err := cfg.Database.UpdateUser(req.Context(), database.UpdateUserParams{
        Email:          body.Email,
        HashedPassword: hashedPassword,
        ID:             userID,
    })
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not update user")
        return
    }

    helpers.RespondWithJSON(res, http.StatusOK, User{
        ID:        dbUser.ID,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
        IsChirpyRed: dbUser.IsChirpyRed,
    })
}

func (cfg *ApiConfig) DeleteChirp(res http.ResponseWriter, req *http.Request) {
    tokenString, err := auth.GetBearerToken(req.Header)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Missing or invalid token")
        return
    }

    userID, err := auth.ValidateJWT(tokenString, cfg.JwtSecret)
    if err != nil {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Invalid token")
        return
    }

    chirpID, err := uuid.Parse(req.PathValue("chirpID"))
    if err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, "Invalid chirp ID")
        return
    }

    dbChirp, err := cfg.Database.GetChirpByID(req.Context(), chirpID)
    if err != nil {
        helpers.RespondWithError(res, http.StatusNotFound, "Chirp not found")
        return
    }

    if dbChirp.UserID != userID {
        helpers.RespondWithError(res, http.StatusForbidden, "You are not the author of this chirp")
        return
    }

    if err := cfg.Database.DeleteChirp(req.Context(), chirpID); err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not delete chirp")
        return
    }

    res.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) PolkaWebhook(res http.ResponseWriter, req *http.Request) {
    apiKey, err := auth.GetAPIKey(req.Header)
    if err != nil || apiKey != cfg.PolkaKey {
        helpers.RespondWithError(res, http.StatusUnauthorized, "Invalid API key")
        return
    }

    type webhookData struct {
        UserID uuid.UUID `json:"user_id"`
    }
    type requestBody struct {
        Event string      `json:"event"`
        Data  webhookData `json:"data"`
    }

    var body requestBody
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, "Something went wrong")
        return
    }

    if body.Event != "user.upgraded" {
        res.WriteHeader(http.StatusNoContent)
        return
    }

    if err := cfg.Database.UpgradeUserToChirpyRed(req.Context(), body.Data.UserID); err != nil {
        helpers.RespondWithError(res, http.StatusNotFound, "User not found")
        return
    }

    res.WriteHeader(http.StatusNoContent)
}