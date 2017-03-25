package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB() *sqlx.DB {
	db, err := sqlx.Open("sqlite3", "adb.db")
	if err != nil {
		panic(err)
	}

	return db
}

var DB = NewDB()

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

func getAutocompleteNames() []string {
	type Name struct {
		Name string `db:"Name"`
	}
	names := []Name{}
	err := DB.Select(&names, "SELECT Name FROM Activists")
	if err != nil {
		panic(err)
	}

	ret := []string{}
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

func AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New()
	names := getAutocompleteNames()
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

	fmt.Println("Listening on localhost:8080")
	http.ListenAndServe(":8080", n)
}
