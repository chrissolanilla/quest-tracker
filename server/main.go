package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

)

type server struct {
	db *sql.DB
}

type quest struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Difficulty  string  `json:"difficulty"`
	Completed   bool    `json:"completed"`
	CompletedBy *string `json:"completed_by,omitempty"`
}

type leaderboardRow struct {
	UserID string  `json:"user_id"`
	Name   string  `json:"name"`
	Points float64 `json:"points"`
}


func main() {
	_ = godotenv.Load(".env.api")
	// _ = godotenv.Load("../.env.api")
	// _ = godotenv.Load("../../.env.api")

	dsn := getenv("DATABASE_URL", "")
	port := getenv("API_PORT", "8080")
	origin := getenv("CORS_ORIGIN", "http://localhost:5173")
	if dsn == "" {
		log.Fatal("missing DATABASE_URL")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil { log.Fatal(err) }
	if err := db.Ping(); err != nil { log.Fatal(err) }
	if err := ensureAuthTables(db); err != nil { log.Fatal(err) }

	s := &server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /leaderboard", s.handleLeaderboard)
	mux.HandleFunc("GET /quests", s.handleQuests)

	s.mountAuth(mux, origin)
	s.mountMe(mux)
	s.mountLogout(mux)
	s.mountAsana(mux)

	handler := cors(origin, mux)

	log.Println("api listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func cors(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent); return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`
		select u.id, u.name, coalesce(sc.points,0)
		from users u
		left join scores sc on sc.user_id = u.id
		order by coalesce(sc.points,0) desc, u.name asc
		limit 10`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()

	var out []leaderboardRow
	for rows.Next() {
		var row leaderboardRow
		if err := rows.Scan(&row.UserID, &row.Name, &row.Points); err != nil {
			http.Error(w, err.Error(), 500); return
		}
		out = append(out, row)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}


func (s *server) handleQuests(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`
		select id, name, difficulty, completed, completed_by
		from quests
		order by completed asc, name asc
		limit 200`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()

	var out []quest
	for rows.Next() {
		var q quest
		if err := rows.Scan(&q.ID, &q.Name, &q.Difficulty, &q.Completed, &q.CompletedBy); err != nil {
			http.Error(w, err.Error(), 500); return
		}
		out = append(out, q)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}


// call once after opening DB
func ensureAuthTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
		  id TEXT PRIMARY KEY,
		  name TEXT NOT NULL,
		  avatar_url TEXT
		);

		CREATE TABLE IF NOT EXISTS oauth_accounts (
		  user_id TEXT NOT NULL REFERENCES users(id),
		  provider TEXT NOT NULL,
		  access_token TEXT NOT NULL,
		  refresh_token TEXT,
		  scope TEXT,
		  expires_at TIMESTAMPTZ,
		  PRIMARY KEY (user_id, provider)
		);

		CREATE TABLE IF NOT EXISTS sessions (
		  id TEXT PRIMARY KEY,
		  user_id TEXT NOT NULL REFERENCES users(id),
		  created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS sessions_meta (
		  id TEXT PRIMARY KEY,
		  state TEXT,
		  code_verifier TEXT,
		  created_at TIMESTAMPTZ DEFAULT now()
		);
	`)
	return err
}

