package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

type TranscriptJson struct {
	Results struct {
		Transcripts []struct {
			Transcript string `json:"transcript"`
		} `json:"transcripts"`
	} `json:"results"`
}

type anthropicResponseJson struct {
	ConversationSummary string `json:"conversationSummary"`
	Topics              []struct {
		Topic struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"topic"`
	} `json:"topics"`
	Alerts []struct {
		Alert struct {
			Type        string `json:"type"`
			Description string `json:"description"`
			Quote       string `json:"quote"`
		} `json:"alert"`
	} `json:"alerts"`
}

type TranscriptSummaryResponse struct {
	TranscriptionSummary string               `json:"transcriptionSummary"`
	TranscriptionTopics  []TranscriptionTopic `json:"transcriptionTopics"`
	TranscriptionAlerts  []TranscriptionAlert `json:"transcriptionAlerts,omitempty"`
}

type TranscriptionTopic struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TranscriptionAlert struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Quote       string `json:"quote"`
}

// currently runs once TODO to do this where a sleep of xseconds waits to check for a new file, then run again with new transcript
func main() {

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

	// get messages from sqs, stored in map
	fileBucketDetailMap, err := receiveMessages(sess, queueUrl)
	if err != nil {
		fmt.Printf("no message read, or error: %v", err)
		return
	}

	err = sendEmailsFromMessages(fileBucketDetailMap, sess, client)

}

// getTranscriptionFileFromS3 takes in a session, name of an s3 bucket, and a file in that bucket, and creates it in a local dir
func getTranscriptionFileFromS3(sess *session.Session, bucketName, fileName string) error {

	localPath := filepath.Join("transcripts", filepath.Base(fileName))

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("could not create file: %v\n", err)
	}

	defer file.Close()

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		})
	if err != nil {
		return fmt.Errorf("could not download file from s3 bucket: %v\n", err)
	}

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

		// dont add to map if not a json file. there are other misc file types that are auto added to bucket, like .temp that have caused issues
		if !strings.HasSuffix(sqsMessage.Records[0].S3.Object.Key, "json") {
			continue
		}
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

func sendEmailsFromMessages(messages map[string]string, sess *session.Session, client pb.ParrotServiceClient) error {

	// loop through messages, sending emails
	for fileName, bucketName := range messages {

		transcriptionTool := NewTranscriptSummaryTool()

		if !strings.HasSuffix(fileName, "json") {
			continue
		}
		err := getTranscriptionFileFromS3(sess, bucketName, fileName)
		if err != nil {
			return fmt.Errorf("getting file from s3 failed: %v\n", err)
		}

		transcriptString, err := getTranscriptFromJsonFile(fileName)
		if err != nil {
			return fmt.Errorf("error getting transcript from json file: %v\n", err)
		}

		messageContentText := "Please analyze and summarize this transcript: " + transcriptString

		anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		anthropicClient := anthropic.NewClient(anthropicKey)

		// set to anthropic response and use it for something.
		anthropicResponse, err := anthropicClient.CreateMessages(context.Background(), anthropic.MessagesRequest{
			Model:     anthropic.ModelClaude3Dot5HaikuLatest,
			MaxTokens: 4096,
			Messages: []anthropic.Message{
				{
					Role: anthropic.RoleUser,
					Content: []anthropic.MessageContent{
						{
							Type: anthropic.MessagesContentTypeText,
							Text: &messageContentText,
						},
					},
				},
			},
			Tools: []anthropic.ToolDefinition{
				{
					Name:        transcriptionTool["name"].(string),
					Description: transcriptionTool["description"].(string),
					InputSchema: transcriptionTool["input_schema"],
				},
			},

			ToolChoice: &anthropic.ToolChoice{
				Type: "tool",
				Name: "transcript_summary",
			},
		})

		if err != nil {
			log.Fatalf("Error calling Anthropic API: %v", err)
		}

		fmt.Printf("AnthropicResponse: %v\n", anthropicResponse.Content[0].Text)

		var anthropicAlert string

		if len(anthropicResponse.Content) > 0 {
			if anthropicResponse.Content[0].Text != nil {
				// Get the text if it's a text response
				anthropicAlert = *anthropicResponse.Content[0].Text
				fmt.Printf("anthropic response not nil: %v\n", anthropicAlert)
			} else {
				return fmt.Errorf("unexpected response format from Anthropic API")
			}
		} else {
			return fmt.Errorf("empty content in Anthropic API response")
		}

		var anthropicResponseJson anthropicResponseJson

		err = json.Unmarshal([]byte(anthropicAlert), &anthropicResponseJson)
		if err != nil {
			return fmt.Errorf("anthropic generated response is not json in intended format: %v\n", err)
		}

		// Create request from our data
		req := &pb.AlertDataRequest{}

		fmt.Printf("alertTitle set to: %v\n", req.AlertTitle)

		// Send the request
		resp, err := client.SendAlertData(context.Background(), req)
		if err != nil {
			log.Printf("Failed to send data: %v", err)
		} else {
			log.Printf("Response from service1: %s", resp.Message)
		}
	}

	return nil
}

// TranscriptSummaryTool creates a tool definition for generating transcript summaries
func NewTranscriptSummaryTool() map[string]interface{} {
	return map[string]interface{}{
		"name":        "transcript_summary",
		"description": "Generate a comprehensive summary of a transcript with topics and alerts using well-structured JSON.",
		"input_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transcriptionSummary": map[string]interface{}{
					"type":        "string",
					"description": "A concise summary of the entire transcript, highlighting the main points and key takeaways.",
				},
				"transcriptionTopics": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "The name or title of the topic discussed.",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "A brief description of the topic and what was discussed about it.",
							},
						},
						"required": []string{"name", "description"},
					},
					"description": "An array of key topics identified in the transcript.",
				},
				"transcriptionAlerts": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"type": map[string]interface{}{
								"type":               "string",
								"enum":               []string{"sensitive", "urgent", "action_required", "follow_up", "custom"},
								"descri       ption": "The type of alert identified in the transcript.",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "A description of the alert and why it was flagged.",
							},
							"quote": map[string]interface{}{
								"type":        "string",
								"description": "A direct quote from the transcript that triggered this alert.",
							},
						},
						"required": []string{"type", "description", "quote"},
					},
					"description": "An array of alerts or important items that require attention.",
				},
			},
			"required": []string{"transcriptionSummary", "transcriptionTopics"},
		},
	}
}
