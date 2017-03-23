package main

import (

	"net/http"

	"github.com/urfave/negroni"
	"github.com/unrolled/render"
	"github.com/gorilla/mux"
)

var Render = render.New()

func router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)

	router.HandleFunc("/get_autocomplete_activist_names", AutocompleteActivistsHandler)

	router.HandleFunc("/update_event", UpdateEventHandler)

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{
		Layout: "layout",
	})
	r.HTML(w, http.StatusOK, "event_new", nil)
}

func AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New()
	names := []string{
		"Samer Masterson", "Jake Hobbs", "Jake Something", "yoyoyo",
	}
	r.JSON(w, http.StatusOK, map[string][]string{
		"activist_names": names,
	})
}

func UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	
}

func main() {
	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r := router()
	n.UseHandler(r)

	http.ListenAndServe(":8080", n)
}
