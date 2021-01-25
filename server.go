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

	uuid "github.com/satori/go.uuid"
	"golang.org/x/image/draw"
)

const (
	bulmaPath     = "./public/bulma.min.css"
	fileField     = "file"
	defaultSize   = 30
	defaultColor  = "#FFEB3B"
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
	scolor := r.FormValue("scolor")[1:] + "FF"
	direction := r.FormValue("direction")
	isTransparent := r.FormValue("transparent") != ""

	validFiles := []multipart.File{}

	// we handle files by hand
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, file := range r.MultipartForm.File[fileField] {
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

	if !isTransparent {
		b, err := hex.DecodeString(scolor)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Not transparent background", b)
		clr := color.NRGBA{b[0], b[1], b[2], b[3]}
		//
		for i := 0; i < x; i++ {
			for j := 0; j < y; j++ {
				img.Set(i, j, clr)
			}
		}
	} else {
		log.Println("Transparent background")
	}

	// todo: allow user to select scaler based on quality
	scaler := draw.NearestNeighbor

	// https://pkg.go.dev/golang.org/x/image/draw#Scaler
	for i, f := range validFiles {
		src, _, err := image.Decode(f)
		if err != nil {
			log.Println(err)
			resError(w, r, err)
			return
		}
		// todo: flip based on direction
		x0 := i*(newSize+ssize) + ssize
		y0 := ssize
		if direction == "v" {
			x0, y0 = y0, x0
		}
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
		resError(w, r, err)
		return
	}
	defer dstFile.Close()
	// encode as .png to the file

	err = png.Encode(dstFile, img)
	if err != nil {
		log.Println(err)
		resError(w, r, err)
	}

	tpls.ExecuteTemplate(w, "img.gohtml", fileName)
}

func resError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err)
	http.Redirect(w, r, "/?error=true", http.StatusSeeOther)
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
