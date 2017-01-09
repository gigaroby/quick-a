package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	maxImageHeight = 600
	maxImageWidth  = 1000
)

var (
	htmlRoot       = flag.String("html-root", "html-root", "directory containing files to serve")
	port           = flag.Int("port", 8080, "port to listen on")
	modelServerURL = flag.String("model-server-url", "http://localhost:8000", "model server URL")
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	log.Println("requesting categories from model server")
	categories, err := getCategories(*modelServerURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d categories loaded\n", len(categories))

	http.Handle("/", http.FileServer(http.Dir(*htmlRoot)))
	http.Handle("/classify", &classifyHandler{
		MaxImageHeigth: maxImageHeight,
		MaxImageWidth:  maxImageWidth,

		modelServerURL: *modelServerURL,
		categories:     categories,
	})

	http.Handle("/categories", &categoryHandler{
		categories: categories,
	})

	log.Println("ready to serve requests")
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil))
}
