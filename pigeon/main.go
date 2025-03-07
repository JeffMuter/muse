package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	pb "github.com/jeffmuter/muse/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
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

type rpcServer struct {
	pb.UnimplementedParrotServiceServer
	lastAlert alertData
	mu        sync.Mutex
	alertChan chan alertData
}

func (rpcS *rpcServer) SendAlertData(ctx context.Context, req *pb.AlertDataRequest) (*pb.AlertDataResponse, error) {
	alert := alertData{
		alertTitle:          req.AlertTitle,
		deviceType:          req.DeviceType,
		deviceName:          req.DeviceName,
		eventTime:           req.EventTime,
		eventDate:           req.EventDate,
		deviceLocation:      req.DeviceLocation,
		conversationSummary: req.ConversationSummary,
		alertQuote:          req.AlertQuote,
		fileUrl:             req.FileUrl,
	}

	rpcS.mu.Lock()
	rpcS.lastAlert = alert
	rpcS.mu.Unlock()

	// Send the alert data through the channel
	rpcS.alertChan <- alert

	// Return success response
	return &pb.AlertDataResponse{
		Success: true,
		Message: "AlertData received successfully",
	}, nil
}

func sendEmail(alertData alertData, accessKeyID, secretAccessKey string) error {
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
	subject := "Muse Alert: " + alertData.alertTitle + " Detected"
	textBody := "This is an alert from Muse."

	htmlBody := `<h1>Muse: ` + alertData.alertTitle + ` Detected</h1>
		<br>
		<p>Muse has discovered an ` + alertData.alertTitle + ` from: ` + alertData.deviceType + ` ` + alertData.deviceName + `</p>
		<br>
		<h3>Alert triggered by quote:</h3>
		<br> 
		` + alertData.alertQuote + `
		<br>
		<h3>Conversation Summary:</h3>
		<br>
		` + alertData.conversationSummary + `
		<br>
		<h2>Estimated Event Details:</h3>
		<br>
		<h3>Time: ~` + alertData.eventTime + `</h3>
		<br>
		<h3>Date: ` + alertData.eventDate + `</h3>
		<br>
		<h3>Link to file ` + alertData.fileUrl + `<br>`

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
	alertChan := make(chan alertData)

	server := &rpcServer{
		alertChan: alertChan,
	}

	// Create a TCP listener on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register our server
	pb.RegisterParrotServiceServer(grpcServer, server)

	fmt.Println("pigeon created grpcServer, listening...")

	// Start the gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Main loop to process alerts and send emails
	for alertData := range alertChan {
		fmt.Println("Received alert data, sending email...")
		if err := sendEmail(alertData, accessKeyID, secretAccessKey); err != nil {
			log.Printf("Failed to send email: %v", err)
		}
	}
}
