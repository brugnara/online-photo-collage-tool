package main

import (
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
	// 1 << 20 = 1MB
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		log.Fatalln(err)
		return
	}

	key := "file"

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
			// f, err := fhs[0].Open()
			log.Println(fhs)
		}
	}

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
