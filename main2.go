package main

import (
	"html/template"
	"os"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	templates.ExecuteTemplate(os.Stdout, "event_new.html", nil)
}
