package main

import (
	"database/sql"
	"fmt"
	"github.com/Weso1ek/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) Hits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>\n", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	if cfg.platform != "dev" {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
	}

	err := cfg.dbQueries.DeleteUsers(r.Context())

	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, errDb := sql.Open("postgres", dbURL)

	if errDb != nil {
		log.Fatal(errDb)
	}

	const filepathRoot = "."
	const port = "8080"

	var cfg apiConfig

	cfg.dbQueries = database.New(db)
	cfg.platform = os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("GET /api/healthz", http.HandlerFunc(Health))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(cfg.Hits))
	mux.Handle("POST /admin/reset", http.HandlerFunc(cfg.Reset))
	mux.Handle("POST /api/login", http.HandlerFunc(cfg.handlerLogin))

	//mux.Handle("POST /api/validate_chirp", http.HandlerFunc(ValidateChirp))

	mux.Handle("POST /api/users", http.HandlerFunc(cfg.handlerUsersCreate))
	mux.Handle("POST /api/chirps", http.HandlerFunc(cfg.handlerChirpsCreate))
	mux.Handle("GET /api/chirps", http.HandlerFunc(cfg.handlerChirps))
	mux.Handle("GET /api/chirps/{chirpID}", http.HandlerFunc(cfg.handlerGetChirp))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func Health(w http.ResponseWriter, r *http.Request) {
	body := "OK"

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}
