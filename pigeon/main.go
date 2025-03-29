package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	pb "github.com/jeffmuter/muse/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

// TranscriptData represents the data structure for storing transcript information
type TranscriptData struct {
	transcriptID string
	summary      string
	topics       []*pb.TranscriptionTopic
	alerts       []*pb.TranscriptionAlert
	mu           sync.Mutex
}

// rpcServer implements the TranscriptServiceServer interface
type rpcServer struct {
	pb.UnimplementedPigeonServiceServer
	transcripts         map[string]*TranscriptData
	mu                  sync.Mutex
	notificationChannel chan *pb.TranscriptSummaryResponse
}

// ProcessTranscriptSummary implements the PigeonServiceServer interface
func (rpcS *rpcServer) ProcessTranscriptSummary(ctx context.Context, req *pb.TranscriptSummaryResponse) (*pb.ProcessingResult, error) {
	fmt.Printf("Received transcript summary for processing with ID: %s", req.GetTranscriptId())

	// Process the transcript summary data
	transcriptID := req.GetTranscriptId()
	summary := req.GetTranscriptionSummary()

	fmt.Printf("Processing summary: %s", summary)
	thisTranscriptData := &TranscriptData{
		transcriptID: transcriptID,
		summary:      summary,
		topics:       req.GetTranscriptionTopics(),
		alerts:       req.GetTranscriptionAlerts(),
	}

	// Store transcript data in memory
	rpcS.mu.Lock()
	rpcS.transcripts[transcriptID] = thisTranscriptData
	rpcS.mu.Unlock()

	// add to noti chan, which will trigger an email
	rpcS.notificationChannel <- req

	// Create and return a success result
	return &pb.ProcessingResult{
		Success: true,
		Message: fmt.Sprintf("Successfully processed transcript summary with ID: %s", transcriptID),
	}, nil
}

// GetTranscriptSummary implements the TranscriptService RPC method
func (rpcS *rpcServer) GetTranscriptSummary(ctx context.Context, req *pb.TranscriptRequest) (*pb.TranscriptSummaryResponse, error) {
	transcriptID := req.GetTranscriptId()
	log.Printf("Received request for transcript summary with ID: %s", transcriptID)

	// Create a response with transcript summary information
	response := &pb.TranscriptSummaryResponse{
		TranscriptionSummary: fmt.Sprintf("Summary for transcript %s", transcriptID),
		TranscriptionTopics: []*pb.TranscriptionTopic{
			{
				Name:        "Main Topic",
				Description: "Primary subject of the conversation",
			},
			{
				Name:        "Secondary Topic",
				Description: "Additional relevant subject discussed",
			},
		},
		TranscriptionAlerts: []*pb.TranscriptionAlert{
			{
				Type:        "Urgent",
				Description: "Potential urgent issue detected",
				Quote:       "I need this resolved immediately",
			},
		},
	}

	return response, nil
}

// sendEmail sends an email notification for an alert
func sendEmail(summary *pb.TranscriptSummaryResponse, accessKeyID, secretAccessKey string) error {
	// 1. Configure AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	// 2. Create an SES service client
	svc := ses.New(sess)

	// 3. Prepare email content
	recipient := "jefferymuter@yahoo.com" // Must be verified in SES
	sender := "muterjeffery@gmail.com"    // Must be verified if in sandbox mode

	// Convert alert count to string properly
	alertCount := fmt.Sprintf("%d", len(summary.TranscriptionAlerts))
	subject := "Muse Summary: " + alertCount + " Alerts Detected"

	// Create the HTML body with proper string formatting
	htmlBody := fmt.Sprintf(`<h2>Conversation Summary</h2>
		<br>

		<p>%s</p>
		<br>

		<h3>Alerts:</h3>
		<br>`, summary.TranscriptionSummary)

	// Add alerts if they exist
	for i := range summary.TranscriptionAlerts {
		htmlBody += fmt.Sprintf(`
		%s
		<br>
		<h3>Alert triggered by quote:</h3>
		<br>
		"%s"
		<br>`, summary.TranscriptionAlerts[i].Type, summary.TranscriptionAlerts[i].Quote)
	}

	// Create a text version of the email for clients that don't support HTML
	textBody := fmt.Sprintf("Conversation Summary\n\n%s\n\nAlerts:\n", summary.TranscriptionSummary)

	for i := range summary.TranscriptionAlerts {
		textBody += fmt.Sprintf("%s\nAlert triggered by quote: \"%s\"\n",
			summary.TranscriptionAlerts[i].Type,
			summary.TranscriptionAlerts[i].Quote)
	}

	// 4. Specify email parameters
	input := &ses.SendEmailInput{
		Source: aws.String(sender),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
			},
		},
	}

	// 5. Send the email
	result, err := svc.SendEmail(input)
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	fmt.Printf("Email sent successfully! Message ID: %s\n", *result.MessageId)
	return nil
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get credentials from .env
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Create notification channel for alerts that need emails
	//	notificationChannel := make(chan *pb.TranscriptionAlert, 100)

	notificationChannel := make(chan *pb.TranscriptSummaryResponse, 100)

	// Initialize the server
	server := &rpcServer{
		transcripts:         make(map[string]*TranscriptData),
		notificationChannel: notificationChannel,
	}

	// Create a TCP listener on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register our server as a TranscriptService
	pb.RegisterPigeonServiceServer(grpcServer, server)

	fmt.Println("Pigeon created gRPC server, listening on port 50051...")

	// Start the gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Main loop to process summaries and send emails
	for summary := range notificationChannel {
		err := sendEmail(summary, accessKeyID, secretAccessKey)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}
		time.Sleep(10 * time.Second) // sleep for 10 seconds, my purposes really should never need more.
	}
}
