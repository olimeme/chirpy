package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/olimeme/handlers"
	"github.com/olimeme/internal/database"
)

const port = ":8080"
const root = "."


func checkHealth(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)
	
	mux := http.NewServeMux()
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	apiCfg := handlers.ApiConfig{
		FileserverHits: atomic.Int32{},
		Database:       dbQueries,
		Platform:       platform,
	}

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(root)))))
	mux.HandleFunc("GET /api/healthz", checkHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.GetNumberOfReqs)
	mux.HandleFunc("POST /admin/reset", apiCfg.ClearNumberOfReqs)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)
	
	server := &http.Server{
		Handler: mux,
		Addr:           port,
	}
	log.Printf("Server started at port: %s", port)
	log.Fatal(server.ListenAndServe())
}