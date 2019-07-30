package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/anthonynsimon/bild/transform"
	minio "github.com/minio/minio-go"
)

var s3Client *minio.Client // = s3Init()

// For public buckets
func constructURL() string {
	var url strings.Builder
	if config.S3.TLS {
		url.WriteString("https://")
	} else {
		url.WriteString("http://")
	}
	url.WriteString(strings.Join([]string{config.S3.Host, config.S3.Bucket}, "/"))
	return url.String()
}

const (
	batchMultipartParseMax      = 30000000
	batchMultipartSingleFileMax = 7500000
	singleMultipartParseMax     = 9000000
	acao                        = "*"
)

func generateURL(c *minio.Client, bucket string, object string) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\"xo.jpg\"")
	url, e := c.PresignedGetObject(bucket, object, time.Second*60*60*24*7, reqParams)
	if e != nil {
		fmt.Println(e.Error())
	}
	log.Println(url.Path)
}

func setPolicy(Client *minio.Client, bucket string) {
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucket + `/*"]}]}`
	e := Client.SetBucketPolicy(bucket, policy)
	if e != nil {
		log.Println(e.Error())
	}
}

func s3Init() *minio.Client {
	endpoint := config.S3.Host
	apiKey := config.S3.Key
	secretKey := config.S3.Secret
	Client, e := minio.New(endpoint, apiKey, secretKey, config.S3.TLS)
	if e != nil {
		log.Println(e.Error())
	}
	if config.S3.Bucket != "" {
		makeBucket(Client, config.S3.Bucket)
		setPolicy(Client, config.S3.Bucket)
	} else {
		log.Println("No default bucket set, skipping creation")
	}
	return Client
}

func makeBucket(Client *minio.Client, bucket string) {
	location := ""
	err := Client.MakeBucket(bucket, location)
	if err != nil {
		log.Println(err)
		exists, err := Client.BucketExists(bucket)
		if err != nil && exists {
			log.Println(bucket, "exists")
		} else {
			log.Println(err)
		}
	} else {
		log.Println("Bucket", bucket, "created.")
	}
}

func getPolicy(Client *minio.Client, bucket string) {
	policy, e := Client.GetBucketPolicy(bucket)
	if e != nil {
		log.Println(e.Error())
	}
	policyReader := strings.NewReader(policy)
	fl, _ := os.Create("policy.json")
	io.Copy(fl, policyReader)
}

// Cleans up the file after successful PUT.
func putFile(c *minio.Client, filePath string, bucket string, keyName string, contentType string) error {
	fileObject, e := os.Open(filePath)
	if e != nil {
		return e
	}
	fileInfo, e := fileObject.Stat()
	if e != nil {
		return e
	}
	n, e := c.PutObject(bucket, keyName, fileObject, fileInfo.Size(), minio.PutObjectOptions{ContentType: contentType})
	if e != nil {
		return e
	}
	log.Println("Uploaded file of size ", n)
	fileObject.Close()
	return os.Remove(filePath)
}

func deleteFile(c *minio.Client, bucket, name string) error {
	if e := c.RemoveObject(bucket, name); e != nil {
		log.Println(e.Error())
		return e
	}
	return nil
}

func getFile(c *minio.Client, bucket string, keyName string, newLocalFile string) {
	reader, e := c.GetObject(bucket, keyName, minio.GetObjectOptions{})
	if e != nil {
		log.Println(e.Error())
	}
	defer reader.Close()
	newLocFile, _ := os.Create(newLocalFile)
	defer newLocFile.Close()
	info, e := reader.Stat()
	if e != nil {
		log.Println(e.Error())
	}
	_, e = io.CopyN(newLocFile, reader, info.Size)
	if e != nil {
		log.Println(e.Error())
	}
}

// This should be used as multipurpose single image processing endpoint.
// Parses up to 9 MB, single image files.
// Check forwards: jpeg/jpg/png/gif
// Keys: `image`, `width`, `height`, `bucket`, `app`
func sharedImageHandler(wr http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(singleMultipartParseMax)
	_, fh, e := r.FormFile("image")
	width := r.FormValue("width")
	height := r.FormValue("height")
	bucket := r.FormValue("bucket")
	for _, v := range r.Form {
		if v[0] == "" {
			wr.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(wr, "Invalid form fields.")
		}
	}
	if e != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(wr, "Error while reading file.")
		return
	}
	if fh != nil {
		if fh.Size < singleMultipartParseMax {
			multipartFile, e := fh.Open()
			if e != nil {
				log.Println(e.Error())
				wr.WriteHeader(http.StatusForbidden)
				fmt.Fprint(wr, "Bad file or too large.")
				return
			}
			b, e := ioutil.ReadAll(multipartFile)
			if e != nil {
				log.Println(e.Error())
				return
			}
			savedFilePath, e := tmpSaveFile(b)
			if e != nil {
				log.Println(e.Error())
				return
			}
			w, _ := strconv.Atoi(width)
			h, _ := strconv.Atoi(height)
			if e = resize(savedFilePath, w, h); e != nil {
				log.Println(e.Error())
				return
			}
			p := strings.Split(savedFilePath, "/")
			randNameWithExtension := p[len(p)-1]
			ext := strings.Split(randNameWithExtension, ".")
			if e := putFile(s3Client, savedFilePath, bucket, randNameWithExtension, "image/"+ext[1]); e != nil {
				log.Println(e.Error())
				wr.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(wr, "Error while processing file.")
			}
			imageurl := strings.Join([]string{constructURL(), randNameWithExtension}, "/")
			if e := store(imageurl, r.RemoteAddr, r.FormValue("app"), bucket); e != nil {
				wr.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(wr, "Internal server error.")
				return
			}
			wr.WriteHeader(http.StatusOK)
			wr.Header().Set("Content-Type", "text/plain")
			wr.Write([]byte(imageurl))
			return
		}
		msg := fmt.Sprintf("Error %s Only jpeg or png extensions are accepted, below 9MB.", fh.Filename)
		log.Println(msg)
		wr.WriteHeader(http.StatusForbidden)
		fmt.Fprint(wr, msg)
		return
	}
	log.Println(fmt.Sprintf("Error %s, nil file header.", fh.Filename))
	wr.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(wr, "Not valid multipart.")
	return
}

// Up to 10 images can be processed, key example: `image0` up to `image9`.
// Individual images should have less than `7.5 MB`.
// Total multipart form should be less than `30 MB`.
// Additional field keys excluding the files mentioned above: `bucket`, `app`, `resize` ? `yes`
func sharedBatchImageHandler(wr http.ResponseWriter, r *http.Request) {
	var w, h int
	var e error
	r.ParseMultipartForm(batchMultipartParseMax)
	for k, v := range r.Form {
		if v[0] == "" {
			wr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(wr, "Loop check iteration [%s] - Invalid form fields.", k)
			return
		}
	}
	if strings.TrimSpace(r.FormValue("resize")) == "yes" {
		w, e = strconv.Atoi(r.FormValue("width"))
		if e != nil {
			wr.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(wr, "Invalid resize form values.")
			return
		}
		h, e = strconv.Atoi(r.FormValue("height"))
		if e != nil {
			wr.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(wr, "Invalid resize form values.")
			return
		}
	}
	bucket := r.FormValue("bucket")
	var s3UploadedBatch []string
	for i := 0; i < 10; i++ {
		key := strings.Join([]string{"image", strconv.Itoa(i)}, "")
		_, fh, e := r.FormFile(key)
		if e != nil {
			fmt.Println(e.Error())
		}
		if fh != nil {
			if fh.Size < batchMultipartSingleFileMax {
				multipartFile, e := fh.Open()
				if e != nil {
					fmt.Println(e.Error())
				}
				b, e := ioutil.ReadAll(multipartFile)
				if e != nil {
					fmt.Println(e.Error())
				}
				detectedType := strings.Split(http.DetectContentType(b), "/")
				if detectedType[0] == "image" {
					localCopyPath, e := tmpSaveFile(b)
					if e != nil {
						log.Println(e.Error())
						return
					}
					if strings.TrimSpace(r.FormValue("resize")) == "yes" {
						if e := resize(localCopyPath, w, h); e != nil {
							log.Println(e.Error())
							return
						}
					}
					p := strings.Split(localCopyPath, "/")
					randNameWithExtension := p[len(p)-1]
					ext := strings.Split(randNameWithExtension, ".")
					if e := putFile(s3Client, localCopyPath, bucket, randNameWithExtension, "image/"+ext[1]); e != nil {
						log.Println(e.Error())
						wr.WriteHeader(http.StatusInternalServerError)
						fmt.Fprint(wr, "Error while processing file.")
					}
					imageurl := strings.Join([]string{constructURL(), randNameWithExtension}, "/")
					s3UploadedBatch = append(s3UploadedBatch, imageurl)
					if e := store(imageurl, r.RemoteAddr, r.FormValue("app"), bucket); e != nil {
						log.Println(e.Error())
						wr.WriteHeader(http.StatusInternalServerError)
						fmt.Fprint(wr, "Internal server error.")
						return
					}

				} else {
					fmt.Fprint(wr, fh.Filename, " Only images are accepted.")
				}
			} else {
				fmt.Fprint(wr, fh.Filename, " Only images under 2.5MB are accepted.")
			}
		}

	}
	b, e := json.Marshal(&s3UploadedBatch)
	if e != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(wr, "Error while processing your request.")
	}
	wr.Write(b)
	return
}

// POST keys: "app", "name"(the url is needed), "bucket"
func deleteFileHandler(wr http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for _, v := range r.PostForm {
		if v[0] == "" {
			wr.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(wr, "Invalid form fields.")
			return
		}
	}
	b, e := ioutil.ReadFile("sql/queries/delete.sql")
	if e != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(wr, "[1]Error while processing your request.")
		return
	}
	_, e = db.Exec(string(b), r.FormValue("app"), r.FormValue("name"))
	if e != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		log.Println(e.Error())
		fmt.Fprint(wr, "[2]Error while processing your request.")
		return
	}
	name := strings.Split(r.FormValue("name"), "/")
	if e = deleteFile(s3Client, r.FormValue("bucket"), name[len(name)-1]); e != nil {
		fmt.Fprint(wr, "[3]Error while processing your request.")
	}
	fmt.Fprint(wr, "Deleted.")
}
