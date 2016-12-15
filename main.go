package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"

	"crypto/md5"
	"image/png"
	"io/ioutil"
	"strings"
)

const (
	maxImageHeight = 400
	maxImageWidth  = 400
)

var (
	htmlRoot = flag.String("html-root", "html-root", "directory containing files to serve")
)

func badRequest(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(msg))
}

func handleImage(rw http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		badRequest(rw, "error parsing data")
		return
	}

	parts := strings.SplitN(req.Form.Get("image"), ",", 2)
	if len(parts) != 2 {
		badRequest(rw, "error decoding image")
		return
	}
	imageBase64 := parts[1]
	// TODO: limit size of the image
	dec, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		log.Println(err)
		badRequest(rw, "error decoding image")
		return
	}
	r := bytes.NewReader(dec)
	config, err := png.DecodeConfig(r)
	if err != nil || config.Height > maxImageHeight || config.Width > maxImageWidth {
		badRequest(rw, "invalid image")
		return
	}

	name := fmt.Sprintf("images/%x.png", md5.Sum(dec))
	err = ioutil.WriteFile(name, dec, 0666)
	if err != nil {
		log.Printf("error writing image to file: %s\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func main() {
	flag.Parse()
	http.Handle("/", http.FileServer(http.Dir(*htmlRoot)))
	http.Handle("/image", http.HandlerFunc(handleImage))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
