package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/olimeme/internal/database"
)

type ApiConfig struct {
    FileserverHits atomic.Int32
    Database       *database.Queries
    Platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
        Email string `json:"email"`
    }

    var body requestBody
    decoder := json.NewDecoder(req.Body)
    if err := decoder.Decode(&body); err != nil {
        res.Header().Set("Content-Type", "application/json")
        res.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
        return
    }

    dbUser, err := cfg.Database.CreateUser(req.Context(), body.Email)
    if err != nil {
        res.Header().Set("Content-Type", "application/json")
        res.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(res).Encode(map[string]string{"error": err.Error()})
        return
    }

    user := User{
        ID:        dbUser.ID,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
    }

    res.Header().Set("Content-Type", "application/json")
    res.WriteHeader(http.StatusCreated)
    json.NewEncoder(res).Encode(user)
}