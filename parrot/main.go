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

type TranscriptSummaryRequest struct {
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
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get API key from environment
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("ANTHROPIC_API_KEY is not set in the environment")
	}

	// check aws sqs for messages
	// Get credentials from .env
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	queueUrl := os.Getenv("AWS_SQS_QUEUE_URL")

	// Create custom credentials
	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")

	// Create session with credentials
	sqsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String("us-east-2"),
			Credentials: creds,
		},
	}))

	mapOfMessageDetails, err := receiveMessages(sqsSession, queueUrl)
	if err != nil {
		fmt.Printf("error receiving messages: %v\n", err)
	}

	fmt.Printf("map of details length: %d\n", len(mapOfMessageDetails))

	// open s3 session
	s3Session, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatalf("error creating new aws session: %v\n", err)
	}

	protoConn, err := grpc.Dial("pigeon:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer protoConn.Close()

	// open protobuff client to send notifications to email service.
	protoClient := pb.NewPigeonServiceClient(protoConn)

	// loop through map, and get transcript from now local filename
	err = sendEmailsFromMessages(mapOfMessageDetails, s3Session, protoClient)
	if err != nil {
		fmt.Printf("error sending emails using map of s3 details recieved from sqs: %v\n", err)
	}

	fmt.Printf("error sending emails: %v\n", err)

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
		MaxNumberOfMessages: aws.Int64(3),
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

// transcriptserviceclient, may be wrong type.
func sendEmailsFromMessages(messages map[string]string, sess *session.Session, client pb.PigeonServiceClient) error {
	// loop through messages, sending emails
	for fileName, bucketName := range messages {
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
		fmt.Printf("transcript raw: %s\n", transcriptString)

		anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		anthropicClient := anthropic.NewClient(anthropicKey)

		summaryData, err := getAISummaryFromTranscript(transcriptString, anthropicKey, *anthropicClient)
		if err != nil {
			return fmt.Errorf("error getting AI summary from transcript: %v", err)
		}

		// Create the gRPC message with Anthropic data
		summary := &pb.TranscriptSummaryRequest{
			TranscriptionSummary: summaryData.TranscriptionSummary,
			TranscriptId:         filepath.Base(fileName), // Using the new field we added
			TranscriptionTopics:  make([]*pb.TranscriptionTopic, len(summaryData.TranscriptionTopics)),
			TranscriptionAlerts:  make([]*pb.TranscriptionAlert, len(summaryData.TranscriptionAlerts)),
		}

		// Convert topics
		for i, topic := range summaryData.TranscriptionTopics {
			summary.TranscriptionTopics[i] = &pb.TranscriptionTopic{
				Name:        topic.Name,
				Description: topic.Description,
			}
		}

		// Convert alerts
		for i, alert := range summaryData.TranscriptionAlerts {
			summary.TranscriptionAlerts[i] = &pb.TranscriptionAlert{
				Type:        alert.Type,
				Description: alert.Description,
				Quote:       alert.Quote,
			}
		}

		// Send the complete summary to the next service for processing
		result, err := client.ProcessTranscriptSummary(context.Background(), summary)
		if err != nil {
			return fmt.Errorf("error sending transcript summary for processing: %v", err)
		}

		// Check if processing was successful
		if !result.Success {
			fmt.Printf("Warning: Processing failed for transcript %s: %s\n",
				filepath.Base(fileName), result.Message)
		} else {
			fmt.Printf("Successfully processed transcript %s: %s\n",
				filepath.Base(fileName), result.Message)
		}
	}

	return nil
}

func createTranscriptSummaryTool() map[string]interface{} {
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
								"type":        "string",
								"enum":        []string{"sensitive", "urgent", "action_required", "follow_up", "custom"},
								"description": "The type of alert identified in the transcript.",
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

func getAISummaryFromTranscript(transcript string, anthropicKey string, client anthropic.Client) (TranscriptSummaryRequest, error) {

	var summary TranscriptSummaryRequest

	// Create tool definition
	transcriptTool := createTranscriptSummaryTool()

	// Create the message content
	messageContent := "Please analyze and summarize this transcript: " + transcript

	// Make the API call with tool calling
	response, err := client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model:     anthropic.ModelClaude3Opus20240229,
		MaxTokens: 4096,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					{
						Type: anthropic.MessagesContentTypeText,
						Text: &messageContent,
					},
				},
			},
		},
		Tools: []anthropic.ToolDefinition{
			{
				Name:        transcriptTool["name"].(string),
				Description: transcriptTool["description"].(string),
				InputSchema: transcriptTool["input_schema"],
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

	// Process the response
	fmt.Println("Response received from Anthropic")

	// Check for tool calls in the response
	if len(response.Content) > 0 && response.Content[0].Type == "tool_use" {
		toolUse := response.Content[0].MessageContentToolUse
		if toolUse != nil {
			// The response contains a tool use with structured JSON
			fmt.Printf("Tool ID: %s\n", toolUse.ID)
			fmt.Printf("Tool Name: %s\n", toolUse.Name)

			// Parse the JSON response from the tool
			err = json.Unmarshal([]byte(toolUse.Input), &summary)
			if err != nil {
				log.Fatalf("Error unmarshalling tool response: %v", err)
			}

			// Print the structured data
			fmt.Printf("\nTranscription Summary: %s\n", summary.TranscriptionSummary)

			fmt.Println("\nTopics:")
			for _, topic := range summary.TranscriptionTopics {
				fmt.Printf("- %s: %s\n", topic.Name, topic.Description)
			}

			if len(summary.TranscriptionAlerts) > 0 {
				fmt.Println("\nAlerts:")
				for _, alert := range summary.TranscriptionAlerts {
					fmt.Printf("- %s: %s\n  Quote: %s\n",
						alert.Type, alert.Description, alert.Quote)
				}
			}

			// If you need to access the raw JSON
			rawJSON, _ := json.MarshalIndent(summary, "", "  ")
			fmt.Printf("\nRaw JSON:\n%s\n", string(rawJSON))
		}
	} else if len(response.Content) > 0 && response.Content[0].Text != nil {
		// Handle text response (fallback if tool wasn't used)
		fmt.Printf("Text response: %s\n", *response.Content[0].Text)
	} else {
		fmt.Println("Unexpected response format")
	}

	return summary, nil
}
