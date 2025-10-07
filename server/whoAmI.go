package main

import (
	"net/http"
	"encoding/json"
	"fmt"
)

func (s *server) mountMe(mux *http.ServeMux) {
	mux.HandleFunc("GET /me", s.handleMe)
}

func (s *server) handleMe(w http.ResponseWriter, r *http.Request) {
	sid := getSessionCookie(r)
	fmt.Println("SID:", sid)

	if sid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		fmt.Println("SID empty")
		return
	}
	var userID, name string
	err := s.db.QueryRow(`
		select u.id, u.name
		from sessions s
		join users u on u.id = s.user_id
		where s.id=$1`, sid).Scan(&userID, &name)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		fmt.Println("our error is: ", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user_id": userID,
		"name":    name,
	})
}

func (s *server) mountLogout(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/logout", s.handleLogout)
	fmt.Println("Logout mounted")
}

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sid := getSessionCookie(r)
	if sid != "" {
		_, _ = s.db.Exec(`delete from sessions where id=$1`, sid)
		http.SetCookie(w, &http.Cookie{
			Name:     "sid",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

