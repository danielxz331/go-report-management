package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadFileToSpace uploads a file to DigitalOcean Space
func UploadFileToSpace(filePath, fileName string) error {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpoint := "https://sfo3.digitaloceanspaces.com"
	region := "sfo3"
	bucket := "reportstesting"
	folder := "reports"

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filePath, err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)

	// Initialize a session in the sfo3 region that the DigitalOcean Space resides in.
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true), // Required for DigitalOcean Spaces
	})
	if err != nil {
		return fmt.Errorf("failed to create session, %v", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Upload the file to the DigitalOcean Space
	key := fmt.Sprintf("%s/%s", folder, fileName)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   fileBytes,
		ACL:    aws.String("public-read"), // Change to "private" if you don't want the file to be publicly accessible
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to space, %v", err)
	}

	log.Printf("Successfully uploaded %q to %q\n", fileName, key)
	return nil
}
