package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gigaroby/quick-a/imageutil"
	"github.com/gigaroby/quick-a/model"
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

func (c *classifyHandler) processImage(dataURL string) (img image.Image, err error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, errors.New("malformed data URL")
	}

	b64 := parts[1]
	// image is too big. computeB64Size computes the max size of the image when converted to base64
	if len(b64) > computeB64Size(c.MaxImageWidth*c.MaxImageHeigth*4) {
		return nil, errors.New("image too large")
	}

	dec, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	image, err := png.Decode(bytes.NewReader(dec))
	if err != nil {
		return nil, err
	}

	// converts the image into a format the backend can understand
	return imageutil.ConvertTo(imageutil.ConvertTo(image, imageutil.AlphaAsWhite), color.GrayModel), nil
}

func (c *classifyHandler) saveOriginal(basePath, session, expectedCategory string, pngData []byte) {
	category := expectedCategory
	if category == "" {
		category = "unknown"
	}
	dir := filepath.Join(basePath, category)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("error saving original image: %s\n", err)
		return
	}

	name := session
	if session == "" {
		h := fnv.New32()
		h.Write(pngData)
		name = hex.EncodeToString(h.Sum(nil))
	}
	name += ".png"

	if err := ioutil.WriteFile(filepath.Join(dir, name), pngData, 0644); err != nil {
		log.Printf("error saving original image: %s\n", err)
		return
	}
}

// prepareImagePOST creates a http request that will send image data as file
func prepareImagePOST(endpoint string, imageData []byte) (*http.Request, error) {
	b64 := base64.StdEncoding.EncodeToString(imageData)
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	w, err := mw.CreateFormFile("image", "image")
	if err != nil {
		return nil, err
	}
	_, err = w.Write([]byte(b64))
	if err != nil {
		return nil, err
	}
	ct := mw.FormDataContentType()
	mw.Close()

	req, _ := http.NewRequest("POST", endpoint, body)
	req.Header.Set("Content-Type", ct)
	return req, nil
}

func (c *classifyHandler) classify(imageData []byte) (model.Predictions, error) {
	req, err := prepareImagePOST(strings.TrimRight(c.modelServerURL, "/")+"/classify/", imageData)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
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
		log.Printf("error parsing form: %s\n", err)
		http.Error(rw, "invalid input data", http.StatusBadRequest)
		return
	}

	image, err := c.processImage(req.Form.Get("image"))
	if err != nil {
		log.Printf("error processing image: %s\n", err)
		http.Error(rw, "invalid image", http.StatusBadRequest)
		return
	}
	category := req.Form.Get("expected_category")
	session := req.Form.Get("session")

	imageData := new(bytes.Buffer)
	err = png.Encode(imageData, image)
	if err != nil {
		log.Printf("error encoding image: %s\n", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	go c.saveOriginal("images", session, category, imageData.Bytes())

	top3, err := c.classify(imageData.Bytes())
	if err != nil {
		log.Printf("error classifying image: %s\n", err)
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(top3)
}
