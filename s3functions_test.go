package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func init() {
	defTables()
	config.S3.Host = "play.minio.io:9000"
	config.S3.Bucket = "magick-crop"
	config.S3.Key = "Q3AM3UQ867SPQQA43P2F"
	config.S3.Secret = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	s3Client = s3Init()
}

func TestS3Init(t *testing.T) {
	if o, e := s3Init().BucketExists(config.S3.Bucket); e != nil || !o {
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
	if e != nil {
		t.Error("Fail[1]:", e.Error())
	}
	fileName := string(makeRandomString(15))
	folder, e := os.UserHomeDir()
	if e != nil {
		t.Error("Fail[2]:", e.Error())
	}
	saveFilePath := strings.Join([]string{folder, "/", fileName, ".png"}, "")
	if e = ioutil.WriteFile(saveFilePath, b, os.ModePerm); e != nil {
		t.Error("Fail[3]:", e.Error())
	}
	if e = putFile(s3, saveFilePath, config.S3.Bucket, fileName, "image/png"); e != nil {
		t.Error("Fail[4]:", e.Error())
	}
	url := strings.Join([]string{"https:/", config.S3.Host, config.S3.Bucket, fileName}, "/")
	response2, e := http.Get(url)
	if e != nil {
		t.Error("Fail[5]:", e.Error())
	}
	b2, e := ioutil.ReadAll(response2.Body)
	if e != nil {
		t.Error("Fail[6]:", e.Error())
	}
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
	m["bucket"] = config.S3.Bucket
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
	if rows, e := db.Query("select bucket from files_stored where name = $1;", string(resultBody)); e != nil || !rows.Next() {
		t.Error("Fail: ", e.Error())
	}
	if s3file[2] != config.S3.Bucket+"."+config.S3.Host {
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
	m["bucket"] = config.S3.Bucket
	m["width"] = "50"
	m["height"] = "50"
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
	if img[0] == "" {
		t.Errorf("Fail: %s, %s", img[0], e.Error())
	}
}
func TestDeleteFileHandler(t *testing.T) {
	wr := httptest.NewRecorder()
	r := new(http.Request)
	body := make(url.Values)
	body.Set("app", "test")
	body.Set("bucket", config.S3.Bucket)
	response, e := http.Get("https://via.placeholder.com/1500")
	if e != nil {
		t.Error(e.Error())
	}
	b, e := ioutil.ReadAll(response.Body)
	if e != nil {
		t.Error(e.Error())
	}
	loc, e := tmpSaveFile(b)
	if e != nil {
		t.Error(e.Error())
	}
	body.Set("name", loc)
	r.PostForm = body
	name := strings.Split(loc, "/")
	s3 := s3Init()
	putFile(s3, loc, config.S3.Bucket, name[len(name)-1], "image/png")
	store(loc, r.RemoteAddr, "test", config.S3.Bucket)
	http.HandlerFunc(deleteFileHandler).ServeHTTP(wr, r)
	if wr.Result().Status != "200 OK" {
		t.Error("Fail: ", wr.Result().Status)
	}
}
