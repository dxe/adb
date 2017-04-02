package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/directactioneverywhere/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/negroni"

	_ "github.com/mattn/go-sqlite3"
)

func router() *mux.Router {
	main := MainController{db: model.NewDB("adb.db")}

	router := mux.NewRouter()
	router.HandleFunc("/", main.UpdateEventHandler)
	router.HandleFunc("/update_event/{event_id:[0-9]+}", main.UpdateEventHandler)
	router.HandleFunc("/list_events", main.ListEventsHandler)

	// API
	router.HandleFunc("/activist_names/get", main.AutocompleteActivistsHandler)
	router.HandleFunc("/event/save", main.EventSaveHandler)

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, req *http.Request) {
	events, err := model.GetEvents(c.db)
	if err != nil {
		panic(err)
	}
	renderTemplate(w, "event_list", map[string]interface{}{
		"Events": events,
	})
}

var templates = template.Must(template.New("").Funcs(
	template.FuncMap{
		"formatdate": func(date time.Time) string {
			return date.Format(model.EventDateLayout)
		},
		"datenotzero": func(date time.Time) bool {
			return !time.Time{}.Equal(date)
		},
	}).ParseGlob("templates/*.html"))

func renderTemplate(w io.Writer, name string, data interface{}) {
	if err := templates.ExecuteTemplate(w, name+".html", data); err != nil {
		panic(err)
	}
}

func writeJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (c MainController) UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	var event model.Event
	if eventIDStr, ok := vars["event_id"]; ok {
		eventID, err := strconv.Atoi(eventIDStr)
		if err != nil {
			panic(err)
		}
		event, err = model.GetEvent(c.db, eventID)
		if err != nil {
			panic(err)
		}
	}

	renderTemplate(w, "event_new", map[string]interface{}{
		"Event": event,
	})
}

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	names := model.GetAutocompleteNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) EventSaveHandler(w http.ResponseWriter, req *http.Request) {
	event, err := model.CleanEventData(c.db, req.Body)
	if err != nil {
		panic(err)
	}

	err = model.InsertEvent(c.db, event)
	if err != nil {
		panic(err)
	}
}

func main() {
	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r := router()
	n.UseHandler(r)

	fmt.Println("Listening on localhost:8080")
	http.ListenAndServe(":8080", n)
}
