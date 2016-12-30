package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gigaroby/quick-a/model"
)

var (
	malformedDataURL = errors.New("malformed data URL")
	invalidImage     = errors.New("invalid image")
)

const (
	defaultMaxWidth  = 400
	defaultMaxHeight = 400
)

// classifyHandler receives a data-url for a png image
// and responds with the classification output for that image
type classifyHandler struct {
	MaxImageWidth  int
	MaxImageHeigth int

	modelServerURL string
	categories     map[int]string
}

func badRequest(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(msg))
}

// computeB64Size computes the number of bytes necessary to represent size bytes to base64
func computeB64Size(size int) int {
	// see http://stackoverflow.com/questions/4715415/base64-what-is-the-worst-possible-increase-in-space-usage
	return ((size + 2) / 3) * 4
}

func (c *classifyHandler) checkImage(dataURL string) (original string, data []byte, err error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", nil, malformedDataURL
	}

	b64 := parts[1]
	// image is too big. computeB64Size computes the max size of the image when converted to base64
	if len(b64) > computeB64Size(c.MaxImageWidth*c.MaxImageHeigth*4) {
		return "", nil, invalidImage
	}

	dec, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", nil, err
	}

	config, err := png.DecodeConfig(bytes.NewReader(dec))
	if err != nil || config.Height > c.MaxImageHeigth || config.Width > c.MaxImageWidth {
		return "", nil, invalidImage
	}

	return b64, dec, nil
}

func (c *classifyHandler) saveOriginal(basePath, category string, data []byte) {
	dir := filepath.Join(basePath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("error saving original image: %s\n", err)
		return
	}
	h := fnv.New32()
	h.Write(data)
	name := hex.EncodeToString(h.Sum(nil)) + ".png"
	if err := ioutil.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		log.Printf("error saving original image: %s\n", err)
		return
	}
}

func (c *classifyHandler) classify(b64Image string) (model.Predictions, error) {
	res, err := http.PostForm(c.modelServerURL+"/classify/", url.Values{
		"image": []string{b64Image},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	pred := model.Predictions{}
	err = json.NewDecoder(res.Body).Decode(&pred)
	if err != nil {
		return nil, err
	}
	return pred.Top(3).FillNames(c.categories), nil
}

func (c *classifyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(rw, "invalid input data", http.StatusBadRequest)
		return
	}

	b64, data, err := c.checkImage(req.Form.Get("image"))
	if err != nil {
		http.Error(rw, "invalid image", http.StatusBadRequest)
		return
	}

	category := req.Form.Get("expected_category")
	if category == "" {
		category = "unknown"
	}

	go c.saveOriginal("images", category, data)

	top3, err := c.classify(b64)
	if err != nil {
		log.Printf("error classifying image: %s\n", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(top3)
}
