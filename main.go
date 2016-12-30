package main

import (
	"flag"
	"log"
	"net/http"
)

const (
	maxImageHeight = 400
	maxImageWidth  = 400
)

var (
	htmlRoot       = flag.String("html-root", "html-root", "directory containing files to serve")
	metadataPath   = flag.String("metadata", "METADATA", "file containing metadata about the model")
	modelServerURL = flag.String("model-server-url", "http://localhost:8090", "model server URL")
)

func main() {
	flag.Parse()

	categories, err := getCategories(*metadataPath)
	if err != nil {
		log.Fatal(err)
	}

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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
