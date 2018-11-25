package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3_REGION = "eu-west-1"
	S3_BUCKET = "rayyildiz-photos"
	FS_ROOT   = "G:\\Photos\\Downloads\\" // "G:\\test\\"
)

func main() {

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(S3_REGION),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Fatal(err)
	}

	svc := s3.New(s)

	i := 0

	files := getFileList()
	for _, strFile := range files {
		f, err := os.Open(strFile)
		if err != nil {
			log.Printf("could get file info ")
		}
		// log.Printf("file info %s", f.Name())

		fileName := f.Name()

		newName := fileName[len(FS_ROOT):]
		err = CheckFileExist(svc, newName)
		log.Printf("%s file status %v", newName, err)
		if err == nil {
			err = AddFileToS3(svc, newName)
			if err == nil {
				i++
			} else {
				log.Printf("%s upload status %v", newName, err)
			}
		}

		f.Close()
	}

	log.Printf("totally uploaded file number %d", i)
}

func getFileList() []string {

	var files []string
	err := filepath.Walk(FS_ROOT, func(path string, info os.FileInfo, err error) error {
		dir := path
		if !(strings.Contains(dir, "2013") || strings.Contains(dir, "2014") || strings.Contains(dir, "2015") ||
			strings.Contains(dir, "2016") || strings.Contains(dir, "2013")) {

			name := strings.ToLower(info.Name())
			if !info.IsDir() && (strings.HasSuffix(name, ".mp4") || strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png")) {
				files = append(files, path)
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("error getting files %v", err)
	}

	return files
}

func AddFileToS3(client *s3.S3, fileDir string) error {

	fileDir = strings.Replace(fileDir, "\\", "/", 1)
	// Open the file for use
	file, err := os.Open(FS_ROOT + fileDir)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(S3_BUCKET),
		Key:           aws.String(fileDir),
		ACL:           aws.String("private"),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		StorageClass:  aws.String("ONEZONE_IA"),
	})
	return err
}

func CheckFileExist(client *s3.S3, fileName string) error {

	fileName = strings.Replace(fileName, "\\", "/", 1)

	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(fileName),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)

		if ok && aerr.Code() == "NotFound" {
			return nil
		}

	}
	// log.Printf("result %v", result)

	return fmt.Errorf("file exist %s", fileName)
}
