package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestTmpSaveFile(t *testing.T) {
	response, e := http.Get("https://via.placeholder.com/1500")
	if e != nil {
		t.Error(e.Error())
	}
	var b2 []byte
	b, _ := ioutil.ReadAll(response.Body)
	location, e := tmpSaveFile(b)
	if e != nil {
		t.Fail()
	}
	if b2, e = ioutil.ReadFile(location); e != nil || len(b) == 0 {
		t.Error("Fail: ", e.Error())
	}
	if len(b) != len(b2) {
		t.Error(len(b), len(b2))
	}
}

func TestResize(t *testing.T) {
	response, e := http.Get("https://via.placeholder.com/1500")
	if e != nil {
		t.Error(e.Error())
	}
	b, _ := ioutil.ReadAll(response.Body)
	initialLength := len(b) // inital length
	filename := string(makeRandomString(15))
	path, _ := os.UserHomeDir()
	testFile := strings.Join([]string{path, "/", filename, ".png"}, "")
	e = ioutil.WriteFile(testFile, b, os.ModePerm)
	if e != nil {
		t.Error(e.Error())
	}
	resize(testFile, 150, 150)
	b, e = ioutil.ReadFile(testFile)
	if initialLength <= len(b) {
		t.Error("Fail: expected smaller byte slice.", "Initial:", initialLength, ", resized:", len(b))
	}
}
