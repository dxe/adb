package members

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dxe/adb/config"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

func Route(r *mux.Router, db *sqlx.DB) {
	handle := func(path string, method func(*server)) {
		r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			method(&server{db, w, r})
		})
	}

	handle("/", (*server).index)
	handle("/login", (*server).login)
	handle("/auth", (*server).auth)
}

type server struct {
	db *sqlx.DB
	w  http.ResponseWriter
	r  *http.Request
}

func (s *server) queryJSON(data interface{}, query string, args ...interface{}) error {
	var buf []byte
	if err := s.db.QueryRowContext(s.r.Context(), query, args...).Scan(&buf); err != nil {
		return err
	}
	return json.Unmarshal(buf, data)
}

func (s *server) error(err error) {
	s.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(s.w, err)
}

func (s *server) render(tmpl *template.Template, data interface{}) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		s.error(err)
		return
	}
	s.w.Header().Set("Content-Type", "text/html; charset=utf-8")
	s.w.Write(buf.Bytes())
}

func (s *server) redirect(dest string) {
	http.Redirect(s.w, s.r, dest, http.StatusFound)
}

func absURL(path string) string {
	if config.IsProd {
		return "https://members.dxesf.org" + path
	}
	return "http://localhost:8080/members" + path
}
