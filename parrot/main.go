package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	receiveMessages(sqsClient, queueUrl)
}

func receiveMessages(svc *sqs.SQS, queueUrl string) (string, error) {
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
		return "", err
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
		return
	}
	fmt.Println("Message Deleted")

	// Parse your SQS message into this struct
	json.Unmarshal([]byte(resp.Messages[0].Body), &sqsMessage)
}

func getTranscriptionFileFromS3(fileName string) error {
	bucket := "https://musetranscriptions.s3.us-east-2.amazonaws.com/"

	file, err := s3.GetObjectOutput(fileName)

	return nil
}
