package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	pb "github.com/jeffmuter/muse/proto"
	"github.com/joho/godotenv"
	"github.com/liushuangls/go-anthropic/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type alertData struct {
	alertTitle          string
	deviceType          string
	deviceName          string
	eventTime           string
	eventDate           string
	deviceLocation      string
	conversationSummary string
	alertQuote          string
	fileUrl             string
}

type TranscriptJson struct {
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
	} `json:"results"`
}

type ClaudeTranscriptSummary struct {
	Summary string
	Topics  []TranscriptTopicDetails
	Alerts  []TranscriptAlertDetails
}

type TranscriptTopicDetails struct {
	Name      string
	Timestamp time.Time
	Summary   string
}

type TranscriptAlertDetails struct {
	Type     string
	Summary  string
	Severity string
}

// currently runs once TODO to do this where a sleep of xseconds waits to check for a new file, then run again with new transcript
func main() {
	// create dummy alert data.
	alertData := alertData{
		alertTitle:          "pizza rolls detected",
		deviceType:          "camera",
		deviceName:          "camera213",
		eventTime:           "02:13am",
		eventDate:           "01/02/2025",
		deviceLocation:      "London",
		conversationSummary: "this is the conversation summary",
		alertQuote:          "this is a quote",
		fileUrl:             "pretendthisisafileUrl",
	}

	// Set up a connection to pigeon service
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewParrotServiceClient(conn)

	// Load .env file
	err = godotenv.Load()
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

	fileBucketDetailMap, err := receiveMessages(sess, queueUrl)
	if err != nil {
		fmt.Printf("no message read, or error: %v", err)
		return
	}

	for bucketName, fileName := range fileBucketDetailMap {
		err = getTranscriptionFileFromS3(sess, bucketName, fileName)
		if err != nil {
			fmt.Printf("getting file from s3 failed: %v\n", err)
			return
		}

		fmt.Println("process finished")

		transcriptString, err := getTranscriptFromJsonFile(fileName)
		if err != nil {
			fmt.Printf("error getting transcript from json file: %v\n", err)
			return
		}

		fmt.Println("transcript retreived successfully")

		anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		anthropicClient := anthropic.NewClient(anthropicKey)

		prompt := `
	Summarize the following transcript by returning a JSON object in this format with the following fields filled out: {
		"summary": "Overall conversation summary",
		"topics": [
			{
				"name": "topic_name",
				"summary": "summary of this topic discussion"
			}
		],
		"alerts": [
			{
				"type": "crime|sensitive_topic",
				"quote": "the exact substring that triggered the alert",
				"summary": "details about the alert",
				"severity": "high|medium|low",
				"mentioned_at": "how far into the transcript",
			}
		]
	} 
	Here is the transcript:` + transcriptString

		claudeResponse, err := anthropicClient.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model: anthropic.ModelClaude3Haiku20240307,
			Messages: []anthropic.Message{
				anthropic.NewUserTextMessage(prompt),
			},
			MaxTokens: 1000,
		})
		if err != nil {
			var e *anthropic.APIError
			if errors.As(err, &e) {
				fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
			} else {
				fmt.Printf("Messages error: %v\n", err)
			}
			return
		}

		fmt.Printf("printing this to stop error %v\n", claudeResponse)

		// Create request from our data
		req := &pb.AlertDataRequest{
			AlertTitle:          alertData.alertTitle,
			DeviceType:          alertData.deviceType,
			DeviceName:          alertData.deviceName,
			EventTime:           alertData.eventTime,
			EventDate:           alertData.eventDate,
			DeviceLocation:      alertData.deviceLocation,
			ConversationSummary: alertData.conversationSummary,
			AlertQuote:          alertData.alertQuote,
			FileUrl:             alertData.fileUrl,
		}

		// Send the request
		resp, err := client.SendAlertData(context.Background(), req)
		if err != nil {
			log.Printf("Failed to send data: %v", err)
		} else {
			log.Printf("Response from service1: %s", resp.Message)
		}
	}

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

// receiveMessages takes in queue url, and if a message exists, returns the bucket name, file name of the desired file in a map.
func receiveMessages(sess *session.Session, queueUrl string) (map[string]string, error) {
	mapOfFileDetails := make(map[string]string)

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

	sqsClient := sqs.New(sess)

	resp, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(2),
	})
	if err != nil {
		fmt.Println("Error receiving messages:", err)
		return mapOfFileDetails, err
	}
	if len(resp.Messages) == 0 {
		fmt.Println("receiving messages failed, no messages.")
		return mapOfFileDetails, fmt.Errorf("receiving messages failed, no messages.\n")
	}

	fmt.Println("length of messages: " + strconv.Itoa(len(resp.Messages)))

	// Process the messages
	fmt.Printf("Processing messages: %s\n", aws.StringValue(resp.Messages[0].Body))

	for _, message := range resp.Messages {
		// Delete the message
		_, delErr := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueUrl),
			ReceiptHandle: message.ReceiptHandle,
		})
		if delErr != nil {
			fmt.Println("Delete Error", delErr)
			return mapOfFileDetails, err
		}

		// Parse your SQS message into this struct
		json.Unmarshal([]byte(*message.Body), &sqsMessage)

		mapOfFileDetails[sqsMessage.Records[0].S3.Object.Key] = sqsMessage.Records[0].S3.Bucket.Name
	}

	return mapOfFileDetails, nil
}

// getTranscriptFromJsonFile takes in a json file name, finds the transcript within, and returns the transcript
func getTranscriptFromJsonFile(fileName string) (string, error) {

	if !strings.HasPrefix(fileName, "transcripts/") {
		return "", fmt.Errorf("error: fileName did not begin with transcripts/ as expected...")
	}

	if !strings.HasSuffix(fileName, ".json") {
		return "", fmt.Errorf("error, non .json file passed in.")
	}

	fileData, err := os.ReadFile("./" + fileName)
	if err != nil {
		return "", fmt.Errorf("error reading from file: %s, error: %v\n", fileName, err)
	}

	var jsonTranscript TranscriptJson
	err = json.Unmarshal(fileData, &jsonTranscript)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling fileData from filename: %s, fileData: %v, err: %v\n", fileName, fileData, err)
	}

	if len(jsonTranscript.Results.Transcripts) == 0 {
		return "", fmt.Errorf("error getting transcript from file, no transcript, length of 0... transcript struct: %v", jsonTranscript)
	}

	if len(jsonTranscript.Results.Transcripts) > 1 {
		return "", fmt.Errorf("error, unexpected length of json transcripts. len: %d, struct data: %v\n", len(jsonTranscript.Results.Transcripts), jsonTranscript)
	}

	return jsonTranscript.Results.Transcripts[0].Transcript, nil

}
