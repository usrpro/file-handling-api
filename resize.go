package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

func genPseudoRand() *rand.Rand {
	rInt := rand.New(rand.NewSource(rand.Int63() * time.Now().UnixNano()))
	return rInt
}

// makeRandomString generates a pseudo random string with the length specified as parameter.
func makeRandomString(bytesLength int) []byte {
	byteVar := make([]byte, bytesLength)
	chars := "abcdefghijklmnopqrstuvwxyz123456789" // our posibilities
	for i := range byteVar {
		x := genPseudoRand()
		byteVar[i] = chars[x.Intn(len(chars))]
	}
	return byteVar
}

// Attempts to create tmp folder in $HOME and save the file with random name.
func tmpSaveFile(b []byte) (string, error) {
	var ext string
	switch http.DetectContentType(b) {
	case "image/jpg":
		ext = ".jpg"
	case "image/jpeg":
		ext = ".jpeg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	}
	path, e := os.UserHomeDir()
	if e != nil {
		return "", e
	}
	folder := filepath.Join(path, "tmp")
	os.Mkdir(folder, os.ModePerm)
	location := filepath.Join(folder, strings.Join([]string{string(makeRandomString(30)), ext}, ""))
	if e = ioutil.WriteFile(location, b, os.ModePerm); e != nil {
		return "", e
	}
	log.Println(location)
	return location, nil
}

// Uses bild library to open convert and write the image to the same path.
func resize(imagePath string, w, h int) error {
	i, e := imgio.Open(imagePath)
	if e != nil {
		return e
	}
	resized := transform.Resize(i, w, h, transform.Linear)
	e = imgio.Save(imagePath, resized, imgio.JPEGEncoder(100))
	return e
}
