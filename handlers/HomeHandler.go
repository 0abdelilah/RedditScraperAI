package handlers

import (
	"fmt"
	"net/http"
	"text/template"
)

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/static/",
		http.FileServer(http.Dir("./templates/static")),
	).ServeHTTP(w, r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "This page does not exist.", 404)
		return
	}

	tmpt, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}

	tmpt.Execute(w, nil)
}
