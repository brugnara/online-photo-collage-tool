package main

import (
	"log"
	"os"
)

func init() {

}

func main() {
	log.Println("Starting stuff")
	listen(os.Getenv("PORT"))
}
