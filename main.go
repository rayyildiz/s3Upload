package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	fileRoot := flag.String("directory-root", "./", "Directory root path")
	extensions := flag.String("file-extension", "*", "File extension")
	bucket := flag.String("s3-bucket", "", "S3 bucket name")
	region := flag.String("s3-region", "eu-west-1", "S3 region name")
	storageClass := flag.String("s3-storage-class", "ONEZONE_IA", "S3 Storage class")
	flag.Parse()

	if *bucket == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Root folder :%s", *fileRoot)
	fmt.Printf("File Extension :%s", *extensions)
	fmt.Printf("Bucket name :%s", *bucket)
	fmt.Printf("S3 Region :%s", *region)
	fmt.Printf("S3 Storage Class :%s", *storageClass)

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(*region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		log.Fatal(err)
	}

	svc := s3.New(s)

	i := 0

	files := fileList(*fileRoot, *extensions)
	for _, strFile := range files {
		f, err := os.Open(strFile)
		if err != nil {
			log.Printf("couldn't get file info for %s, %v", strFile, err)
		}
		defer f.Close()

		fileName := f.Name()

		newName := fileName[len(*fileRoot):]
		err = checkIfFileExist(svc, newName, *bucket)
		log.Printf("%s file status %v", newName, err)
		if err == nil {
			err = uploadFile(svc, *fileRoot, newName, *bucket, *storageClass)
			if err == nil {
				i++
			} else {
				log.Printf("%s upload status %v", newName, err)
			}
		}

		// f.Close()
	}

	log.Printf("totally uploaded file number %d", i)
}

func fileList(rootDir, extension string) []string {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {

		name := strings.ToLower(info.Name())
		if !info.IsDir() && (extension == "*" || strings.HasSuffix(name, extension)) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("error getting files %v", err)
	}

	return files
}

func uploadFile(client *s3.S3, fileRoot, fileDir, bucket, storageClass string) error {

	fileDir = strings.Replace(fileDir, "\\", "/", 1)
	file, err := os.Open(fileRoot + fileDir)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(fileDir),
		ACL:           aws.String("private"),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		StorageClass:  aws.String(storageClass),
	})
	return err
}

func checkIfFileExist(client *s3.S3, fileName, bucket string) error {
	fileName = strings.Replace(fileName, "\\", "/", 1)

	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)

		if ok && aerr.Code() == "NotFound" {
			return nil
		}
	}
	return fmt.Errorf("file exist %s", fileName)
}
