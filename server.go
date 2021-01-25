package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	bulmaPath = "./public/bulma.min.css"
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/css/bulma.min.css", bulmaCSS)
	http.Handle("/favicon.ico", http.NotFoundHandler())
}

func listen(port string) {
	if port == "" {
		port = "8080"
	}
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		indexPost(w, r)
		return
	}
	tpls.ExecuteTemplate(w, "index.gohtml", nil)
}

func indexPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		io.WriteString(w, "An error occured :(")
		return
	}
	fmt.Println(r.PostForm)
	io.WriteString(w, "ciao")
}

func bulmaCSS(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(bulmaPath)
	if err != nil {
		return
	}
	defer f.Close()
	w.Header().Set("content-type", "text/css")
	io.Copy(w, f)
}
