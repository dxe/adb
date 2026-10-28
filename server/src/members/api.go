package members

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type meResponse struct {
	Email      string  `json:"email"`
	Name       *string `json:"name"`
	ActivistID *int    `json:"activistId"`
}

// me returns the signed-in member, or 401 {}.
func (a *authServer) me(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	w.Header().Set("Content-Type", "application/json")
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("{}"))
		return
	}
	now := a.now()
	email, refreshed, err := a.store.lookupSession(r.Context(), c.Value, now)
	if err != nil {
		log.Println("member auth: lookup session:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if email == "" {
		clearCookie(w, sessionCookieName, "/")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("{}"))
		return
	}
	if refreshed {
		a.setSessionCookie(w, c.Value, now)
	}

	resp := meResponse{Email: email}
	// Hidden activists are merged-away duplicates; skip them. The email column
	// uses a case-insensitive collation.
	var row struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	err = a.store.db.GetContext(r.Context(), &row,
		`SELECT id, name FROM activists WHERE email = ? AND hidden = 0 ORDER BY id LIMIT 1`, email)
	switch {
	case err == nil:
		resp.Name, resp.ActivistID = &row.Name, &row.ID
	case errors.Is(err, sql.ErrNoRows):
	default:
		log.Println("member auth: activist lookup:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}
