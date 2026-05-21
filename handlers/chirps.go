package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/olimeme/helpers"
	"github.com/olimeme/internal/database"
)
type Chirp struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Body      string    `json:"body"`
    UserID    uuid.UUID `json:"user_id"`
}

func (cfg *ApiConfig) CreateChirp(res http.ResponseWriter, req *http.Request) {
    type requestBody struct {
        Body   string    `json:"body"`
        UserID uuid.UUID `json:"user_id"`
    }

    var body requestBody
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        helpers.RespondWithError(res, http.StatusBadRequest, err.Error())
        return
    }

    if len(body.Body) > 140 {
        helpers.RespondWithError(res, http.StatusBadRequest, "Chirp is too long")
        return
    }

    cleanedBody := replaceProfanity(body.Body)

    dbChirp, err := cfg.Database.CreateChirp(req.Context(), database.CreateChirpParams{
        Body:   cleanedBody,
        UserID: body.UserID,
    })
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, err.Error())
        return
    }

    helpers.RespondWithJSON(res, http.StatusCreated, Chirp{
        ID:        dbChirp.ID,
        CreatedAt: dbChirp.CreatedAt,
        UpdatedAt: dbChirp.UpdatedAt,
        Body:      dbChirp.Body,
        UserID:    dbChirp.UserID,
    })
}

func (cfg *ApiConfig) GetChirps(res http.ResponseWriter, req *http.Request) {
    dbChirps, err := cfg.Database.GetAllChirps(req.Context())
    if err != nil {
        helpers.RespondWithError(res, http.StatusInternalServerError, "Could not retrieve chirps")
        return
    }

    chirps := make([]Chirp, len(dbChirps))
    for i, dbChirp := range dbChirps {
        chirps[i] = Chirp{
            ID:        dbChirp.ID,
            CreatedAt: dbChirp.CreatedAt,
            UpdatedAt: dbChirp.UpdatedAt,
            Body:      dbChirp.Body,
            UserID:    dbChirp.UserID,
        }
    }

    helpers.RespondWithJSON(res, http.StatusOK, chirps)
}

func (cfg *ApiConfig) GetChirp(res http.ResponseWriter, req *http.Request) {
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

    helpers.RespondWithJSON(res, http.StatusOK, Chirp{
        ID:        dbChirp.ID,
        CreatedAt: dbChirp.CreatedAt,
        UpdatedAt: dbChirp.UpdatedAt,
        Body:      dbChirp.Body,
        UserID:    dbChirp.UserID,
    })
}