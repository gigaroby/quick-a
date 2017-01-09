package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gigaroby/quick-a/model"
)

const CategoriesPerGame = 6

type metadata struct {
	Categories map[int]string `json:"categories"`
}

func getCategories(modelServerURL string) (map[int]string, error) {
	res, err := http.Get(strings.TrimRight(modelServerURL, "/") + "/metadata/")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata endpoint returned %d: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	meta := metadata{}
	err = json.NewDecoder(res.Body).Decode(&meta)
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
	requested, err := strconv.Atoi(req.URL.Query().Get("n"))
	if err != nil {
		requested = 6
	}
	if requested > len(c.categories) {
		requested = len(c.categories)
	}

	cat := make(model.Categories, 0)
	for k, v := range c.categories {
		cat = append(cat, model.Category{Index: k, Name: v})
	}

	perm := rand.Perm(len(cat))
	for i := 0; i < requested; i++ {
		cat[i], cat[perm[i]] = cat[perm[i]], cat[i]
	}

	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(cat[:requested])
	if err != nil {
		log.Printf("error serving categories: %s", err)
	}
}
