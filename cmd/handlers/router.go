package handlers

import (
	"html/template"
	"net/http"
)

var Tmpl *template.Template

func SetTemplate(t *template.Template) {
	Tmpl = t
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "index.html", nil)
}

func StreamPage(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "stream.html", nil)
}

func ViewerPage(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "viewer.html", nil)
}
