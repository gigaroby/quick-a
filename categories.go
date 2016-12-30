package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
)

const CategoriesPerGame = 6

type metadata struct {
	Categories map[int]string `json:"categories"`
}

func getCategories(metadataPath string) (map[int]string, error) {
	f, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}

	meta := metadata{}
	err = json.NewDecoder(f).Decode(&meta)
	if err != nil {
		return nil, err
	}

	return meta.Categories, nil
}

// categoryHandler handles client requests for a random set
// of categories
type categoryHandler struct {
	categories map[int]string
}

func (c *categoryHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	cats := make([]int, 0)
	for k, _ := range c.categories {
		cats = append(cats, k)
	}

	if len(cats) < 6 {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Fatalf(
			"not enough categories, expcted at least %d, got %d",
			CategoriesPerGame, len(cats))
	}

	perm := rand.Perm(len(cats))
	result := make(map[int]string)

	for i := 0; i < 6; i++ {
		cat := cats[perm[i]]
		result[cat] = c.categories[cat]
	}

	rw.WriteHeader(http.StatusOK)
	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Printf("error serving categories: %s", err)
	}
}
