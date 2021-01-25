package main

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/image/draw"
)

const (
	bulmaPath     = "./public/bulma.min.css"
	fileField     = "file"
	defaultSize   = 30
	defaultColor  = "#FFFFFF"
	defaultHeight = 300
	maxSSize      = 100
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/css/bulma.min.css", bulmaCSS)
	http.HandleFunc("/tmp/", imageShow)
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
	tpls.ExecuteTemplate(w, "index.gohtml", struct {
		FileField     string
		DefaultSize   int
		DefaultColor  string
		DefaultHeight int
		MaxSSize      int
	}{fileField, defaultSize, defaultColor, defaultHeight, maxSSize})
}

func indexPost(w http.ResponseWriter, r *http.Request) {
	// 1 << 20 = 1MB
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		log.Fatalln(err)
		return
	}

	ssize, err := strconv.Atoi(r.FormValue("ssize"))
	if err != nil || ssize > maxSSize {
		ssize = defaultSize
	}
	newSize, err := strconv.Atoi(r.FormValue("height"))
	if err != nil || newSize < ssize*2 {
		newSize = defaultHeight
	}
	scolor := strings.Replace(r.FormValue("scolor")+"FF", "#", "", -1)
	direction := r.FormValue("direction")

	log.Println(ssize, scolor, direction)
	// todo:
	//  - read params
	//  - read file
	//  - gen && save image
	//  - return to visualizer

	validFiles := []multipart.File{}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, file := range r.MultipartForm.File[fileField] {
			// f, err := fhs[0].Open()
			f, err := file.Open()
			if err != nil {
				log.Println(err)
				continue
			}
			defer f.Close()
			validFiles = append(validFiles, f)
		}
	}

	log.Println(validFiles)

	// gen image based on validFiles len
	x := len(validFiles)*(ssize+newSize) + ssize
	y := newSize + 2*ssize
	if direction == "v" {
		x, y = y, x
	}
	log.Println("Image size:", x, y)
	img := image.NewRGBA(image.Rect(0, 0, x, y))
	// todo: fill image with scolor
	// leave if wanted transparent
	b, err := hex.DecodeString(scolor)
	if err != nil {
		log.Fatal(err)
	}
	clr := color.RGBA{b[0], b[1], b[2], 0}
	//
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			img.Set(i, j, clr)
		}
	}
	//
	log.Printf("%T\n", img)

	// todo: allow user to select scaler based on quality
	scaler := draw.NearestNeighbor

	// https://pkg.go.dev/golang.org/x/image/draw#Scaler
	for i, f := range validFiles {
		src, _, err := image.Decode(f)
		if err != nil {
			log.Fatal(err)
			return
		}
		// todo: flip based on direction
		x0 := i*(newSize+ssize) + ssize
		y0 := ssize
		x1 := x0 + newSize
		y1 := y0 + newSize
		dr := image.Rect(x0, y0, x1, y1)
		fmt.Println(dr, src.Bounds())
		scaler.Scale(img, dr, src, src.Bounds(), draw.Over, nil)
	}

	// open file to save
	fileName := fmt.Sprintf("./tmp/%s.png", uuid.NewV4())
	dstFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()
	// encode as .png to the file

	err = png.Encode(dstFile, img)
	if err != nil {
		log.Fatal(err)
	}

	tpls.ExecuteTemplate(w, "img.gohtml", fileName)
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

func imageShow(w http.ResponseWriter, r *http.Request) {
	log.Println("showing image:", r.URL.Path)
	f, err := os.Open("./" + r.URL.Path)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	io.Copy(w, f)
}
