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
	now := a.now()
	email, issuedAt, ok := readSession(r, now)
	if !ok {
		if _, err := r.Cookie(sessionCookieName); err == nil {
			clearCookie(w, sessionCookieName, "/")
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("{}"))
		return
	}
	// Sliding expiry: re-issue the cookie once it is more than a day old.
	if now.Sub(issuedAt) > sessionRefreshAfter {
		if err := saveSession(w, r, email, now); err != nil {
			log.Println("member auth: refresh session:", err)
		}
	}

	resp := meResponse{Email: email}
	// Hidden activists are merged-away duplicates; skip them. The email column
	// uses a case-insensitive collation.
	var row struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	err := a.store.db.GetContext(r.Context(), &row,
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
