package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get credentials from .env
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	queueUrl := os.Getenv("AWS_SQS_QUEUE_URL")

	// Create custom credentials
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")

	// Create session with credentials
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String("us-east-2"),
			Credentials: creds,
		},
	}))

	sqsClient := sqs.New(sess)
	bucketName, fileName, err := receiveMessages(sqsClient, queueUrl)
	if err != nil {
		fmt.Printf("no message read, or error: %v", err)
		return
	}
	fmt.Printf("message received. bucketname: %s. fileName: %s.\n", bucketName, fileName)

	err = getTranscriptionFileFromS3(sess, bucketName, fileName)
	if err != nil {
		fmt.Printf("getting file from s3 failed: %v\n", err)
		return
	}
	fmt.Println("process finished")
}

// getTranscriptionFileFromS3 takes in a session, name of an s3 bucket, and a file in that bucket, and creates it in a local dir
func getTranscriptionFileFromS3(sess *session.Session, bucketName, fileName string) error {
	fmt.Println("begin to get file...")

	localPath := filepath.Join("transcripts", filepath.Base(fileName))

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("could not create file: %v\n", err)
	}

	defer file.Close()

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		})
	if err != nil {
		return fmt.Errorf("could not download file from s3 bucket: %v\n", err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return nil
}

// TODO: better err handling
// receiveMessages takes in queue url, and if a message exists, returns the bucket name, file name of the desired file.
func receiveMessages(svc *sqs.SQS, queueUrl string) (string, string, error) {
	var sqsMessage struct {
		Records []struct {
			S3 struct {
				Bucket struct {
					Name string `json:"name"`
				} `json:"bucket"`
				Object struct {
					Key string `json:"key"`
				} `json:"object"`
			} `json:"s3"`
		} `json:"Records"`
	}

	resp, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(2),
	})
	if err != nil {
		fmt.Println("Error receiving messages:", err)
		return "", "", err
	}
	if len(resp.Messages) == 0 {
		fmt.Println("receiving messages failed, no messages.")
		return "", "", fmt.Errorf("receiving messages failed, no messages.\n")
	}

	// Process the message
	fmt.Printf("Processing message: %s\n", aws.StringValue(resp.Messages[0].Body))

	// Delete the message
	_, delErr := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: resp.Messages[0].ReceiptHandle,
	})
	if delErr != nil {
		fmt.Println("Delete Error", delErr)
		return "", "", err
	}
	fmt.Println("Message Deleted")

	// Parse your SQS message into this struct
	json.Unmarshal([]byte(*resp.Messages[0].Body), &sqsMessage)

	return sqsMessage.Records[0].S3.Bucket.Name, sqsMessage.Records[0].S3.Object.Key, nil
}
