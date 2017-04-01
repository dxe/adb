package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/directactioneverywhere/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/negroni"

	_ "github.com/mattn/go-sqlite3"
)

func router() *mux.Router {
	main := MainController{db: model.NewDB("adb.db")}

	router := mux.NewRouter()
	router.HandleFunc("/", main.IndexHandler)
	router.HandleFunc("/get_autocomplete_activist_names", main.AutocompleteActivistsHandler)
	router.HandleFunc("/update_event", main.UpdateEventHandler)
	router.HandleFunc("/list_events", main.ListEventsHandler)

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, req *http.Request) {
	// events, err := model.GetEvents(c.db)
	// if err != nil {
	// 	panic(err)
	// }
	// writeJSON(w, map[string]interface{}{
	// 	"events": events,
	// })
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w io.Writer, name string, data interface{}) {
	if err := templates.ExecuteTemplate(w, name+".html", data); err != nil {
		panic(err)
	}
}

func writeJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (c MainController) IndexHandler(w http.ResponseWriter, req *http.Request) {
	renderTemplate(w, "event_new", nil)
}

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	names := model.GetAutocompleteNames(c.db)
	writeJSON(w, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	event, err := model.CleanEventData(c.db, req.Body)
	if err != nil {
		panic(err)
	}

	err = model.InsertNewEvent(c.db, event)
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
