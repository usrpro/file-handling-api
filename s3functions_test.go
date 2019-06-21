package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestS3Init(t *testing.T) {
	if o, e := s3Init().BucketExists(s3Bucket); e != nil || !o {
		t.Errorf("S3Init test failed: %s", e.Error())
	}
}
func TestMakeBucket(t *testing.T) {
	bkt := makeRandomString(15)
	s3 := s3Init()
	if e := s3.MakeBucket(string(bkt), ""); e != nil {
		t.Errorf("Failed at bucket creation test: %s", e.Error())
	}
	if o, e := s3.BucketExists(string(bkt)); e != nil || !o {
		t.Errorf("S3Init test failed: %s", e.Error())
	}
}
func TestPutFile(t *testing.T) {
	s3 := s3Init()
	var err []error
	response, e := http.Get("https://via.placeholder.com/1500")
	err = append(err, e)
	b, e := ioutil.ReadAll(response.Body)
	err = append(err, e)
	fileName := string(makeRandomString(15))
	folder, e := os.UserHomeDir()
	err = append(err, e)
	saveFilePath := strings.Join([]string{folder, "/", fileName, ".png"}, "")
	err = append(err, ioutil.WriteFile(saveFilePath, b, os.ModePerm))
	err = append(err, putFile(s3, saveFilePath, s3Bucket, fileName, "image/png"))
	for _, v := range err {
		if v != nil {
			t.Errorf("S3 PUT test failed: %s", v.Error())
		}
	}
	url := strings.Join([]string{constructURL(), fileName}, "")
	response2, e := http.Get(url)
	err = append(err, e)
	b2, e := ioutil.ReadAll(response2.Body)
	err = append(err, e)
	imgType := http.DetectContentType(b2)
	if imgType != "image/png" {
		t.Errorf("S3 PUT test failed at image type check: %s", imgType)
	}
	os.Remove(saveFilePath)
}

// Generates a multipart request populated with one FormFile with specified fieldname.
// Sets an aditional form field.
func generateMultipartRequest(fileFieldName string, aditionalFormFields map[string]string) (bytes.Buffer, string, error) {

	response, e := http.Get("https://via.placeholder.com/1500")
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	b, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	filename := string(makeRandomString(15))
	path, e := os.UserHomeDir()
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	testFile := strings.Join([]string{path, "/", filename, ".png"}, "")
	e = ioutil.WriteFile(testFile, b, os.ModePerm)
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	fl, e := os.Open(testFile)
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	var bfl bytes.Buffer
	mfw := multipart.NewWriter(&bfl)
	contentType := mfw.FormDataContentType()
	for k, v := range aditionalFormFields {
		mfw.WriteField(k, v)
	}
	fileWriter, e := mfw.CreateFormFile(fileFieldName, "image.png")
	if e != nil {
		return bytes.Buffer{}, "", e
	}
	if n, e := io.Copy(fileWriter, fl); e != nil || n == 0 {
		return bytes.Buffer{}, "", e
	}
	defer fl.Close()
	mfw.Close()
	if e = os.Remove(testFile); e != nil {
		return bytes.Buffer{}, "", e
	}
	return bfl, contentType, nil
}

func TestSharedImageHandler(t *testing.T) {
	wr := httptest.NewRecorder()
	m := make(map[string]string)
	m["bucket"] = "magick-crop"
	m["width"] = "450"
	m["height"] = "350"
	bfl, contentType, e := generateMultipartRequest("image", m)
	if e != nil {
		t.Errorf("%s", e.Error())
	}
	r := httptest.NewRequest("POST", "/", &bfl)
	r.Header.Set("Content-Type", contentType)
	http.HandlerFunc(sharedImageHandler).ServeHTTP(wr, r)
	result := wr.Result()
	if result.StatusCode != 200 {
		t.Errorf("%s", result.Header)
	}
	resultBody, e := ioutil.ReadAll(result.Body)
	if e != nil {
		t.Error(e.Error())
	}
	s3file := strings.Split(string(resultBody), "/")
	if s3file[2] != s3Host || s3file[3] != s3Bucket {
		t.Errorf("%s", string(resultBody))
	}
}

func TestSharedImageHandlerFail(t *testing.T) {
	bufFake := new(bytes.Buffer)
	r := httptest.NewRequest("POST", "/", bufFake)
	r.Header.Set("Content-Type", "multipart/form-data")
	wr := httptest.NewRecorder()
	http.HandlerFunc(sharedImageHandler).ServeHTTP(wr, r)
	resultFalse, e := ioutil.ReadAll(wr.Result().Body)
	if e != nil {
		t.Errorf("%s", e.Error())
	}
	expectedFail := "Error while reading file."
	if wr.Result().StatusCode != 500 || string(resultFalse) != expectedFail {
		t.Errorf("Test sharedPhotoHandler failed, got %s expected %s", resultFalse, expectedFail)
	}
}

func TestSharedBatchImageHandler(t *testing.T) {
	wr := httptest.NewRecorder()
	m := make(map[string]string)
	m["bucket"] = "magick-crop"
	m["width"] = "450"
	m["height"] = "350"
	bfl, contentType, e := generateMultipartRequest("image0", m)
	if e != nil {
		t.Errorf("%s", e.Error())
	}
	r := httptest.NewRequest("POST", "/", &bfl)
	r.Header.Set("Content-Type", contentType)
	http.HandlerFunc(sharedBatchImageHandler).ServeHTTP(wr, r)
	result := wr.Result()
	b, e := ioutil.ReadAll(result.Body)
	if e != nil {
		t.Error(e.Error())
	}
	if e != nil {
		t.Errorf("%s", e.Error())
	}
	var img [10]string
	e = json.Unmarshal(b, &img)
	if !strings.Contains(img[0], "minio") {
		t.Errorf("Fail: %s, %s", img[0], e.Error())
	}
}
