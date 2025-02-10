package cloud

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

// take the unique filepath for this audio file, open, then send the data to our aws bucket
func UploadFileToS3(filePath string) error {
	// aws config
	region := "us-east-2"

	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading env package")
	}
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	fmt.Println(awsAccessKey)

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			awsAccessKey,
			awsSecretKey,
			"",
		),
	})
	if err != nil {
		fmt.Printf("error while creating aws session: %v\n", err)
		return fmt.Errorf("error while creating aws session: %v\n", err)
	}

	svc := s3.New(awsSession)
	bucket := "mutermuse"

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file:", err)
		return fmt.Errorf("error while trying to open file: %v\n", err)
	}
	defer file.Close()

	// Read the contents of the file into a buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		return fmt.Errorf("Error reading file: %v\n", err)
	}

	// This uploads the contents of the buffer to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		fmt.Println("Error uploading file:", err)
		return fmt.Errorf("Error uploading file: %v\n", err)
	}

	fmt.Println("File uploaded successfully.")
	return nil
}
